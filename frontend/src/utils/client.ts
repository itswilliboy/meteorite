import { $fetch, type FetchOptions, type FetchRequest } from "ofetch"
import { type Router } from "vue-router"
import type {
  AuthenticationResponseJSON,
  PublicKeyCredentialCreationOptionsJSON,
  PublicKeyCredentialRequestOptionsJSON,
  RegistrationResponseJSON
} from "@simplewebauthn/browser"

import { setCurrentUser } from "@/composables/useAuth"
import type {
  AdminStats,
  AdminUser,
  AdminUserDetail,
  APIResponse,
  DashboardStats,
  DashboardTimeseries,
  Folder,
  FolderListing,
  Image,
  PaginatedResponse,
  Passkey,
  User
} from "./type"

export class HTTPException extends Error {
  status: number

  constructor(status: number, message: APIResponse<unknown>) {
    super(`HTTPException ${status}: ${message.data ?? message}`)
    this.status = status
  }
}

export class Client {
  constructor(private router: Router) {}

  async request<T>(request: FetchRequest, options?: FetchOptions<"json">): Promise<T> {
    const resp = await $fetch<APIResponse<T>>(request, {
      ...options,
      ignoreResponseError: true
    })

    if (resp.status >= 200 && resp.status < 300) return resp.data

    if (resp.status === 401) {
      setCurrentUser(null)
      if (this.router.currentRoute.value.name !== "login") {
        this.router.push("/login")
      }
    }

    throw new HTTPException(resp.status, resp)
  }

  async requestPaginated<T>(request: FetchRequest, options?: FetchOptions<"json">): Promise<PaginatedResponse<T>> {
    const resp = await $fetch<PaginatedResponse<T>>(request, {
      ...options,
      ignoreResponseError: true
    })

    if (resp.status >= 200 && resp.status < 300) return resp

    if (resp.status === 401) {
      setCurrentUser(null)
      if (this.router.currentRoute.value.name !== "login") {
        this.router.push("/login")
      }
    }

    throw new HTTPException(resp.status, resp)
  }

  async login(username: string, password: string): Promise<User> {
    const resp = await this.request<User>("/api/login", {
      method: "POST",
      body: { username, password }
    })

    return {
      ...resp,
      created_at: new Date(resp.created_at)
    }
  }

  async logout() {
    return this.request("/api/logout", { method: "POST" })
  }

  async changePassword(oldPass: string, newPass: string) {
    const body = {
      old_password: oldPass,
      new_password: newPass
    }
    return this.request("/api/change-password", { method: "POST", body })
  }

  async ping() {
    return this.request("/api/ping")
  }

  async dashboardStats(): Promise<DashboardStats> {
    return this.request<DashboardStats>("/api/stats")
  }

  async dashboardTimeseries(): Promise<DashboardTimeseries> {
    return this.request<DashboardTimeseries>("/api/stats/timeseries")
  }

  async resetToken(): Promise<string> {
    return this.request<string>("/api/reset-token", { method: "POST" })
  }

  async getImages(page: number = 0, sort: string = "date_desc", folderId: string | null = null) {
    const query = new URLSearchParams({ page: String(page), sort })
    if (folderId) query.set("folder_id", folderId)
    return this.requestPaginated<Image[]>(`/api/get-images?${query.toString()}`)
  }

  async deleteImage(imageId: string): Promise<void> {
    return this.request<void>("/api/delete-image", { query: { id: imageId }, method: "POST" })
  }

  async bulkDeleteImages(ids: string[]): Promise<string[]> {
    return this.request<string[]>("/api/bulk-delete-images", { method: "POST", body: { ids } })
  }

  async bulkMoveImages(ids: string[], folderId: string | null): Promise<string[]> {
    return this.request<string[]>("/api/bulk-move-images", { method: "POST", body: { ids, folder_id: folderId } })
  }

  async getFolders(parentId: string | null = null): Promise<FolderListing> {
    const query = parentId ? `?parent_id=${encodeURIComponent(parentId)}` : ""
    return this.request<FolderListing>(`/api/folders${query}`)
  }

  async createFolder(name: string, parentId: string | null = null): Promise<Folder> {
    return this.request<Folder>("/api/folders", { method: "POST", body: { name, parent_id: parentId } })
  }

  async renameFolder(id: string, name: string): Promise<Folder> {
    return this.request<Folder>(`/api/folders/${id}/rename`, { method: "POST", body: { name } })
  }

  async moveFolder(id: string, parentId: string | null): Promise<Folder> {
    return this.request<Folder>(`/api/folders/${id}/move`, { method: "POST", body: { parent_id: parentId } })
  }

  async deleteFolder(id: string): Promise<void> {
    return this.request<void>(`/api/folders/${id}/delete`, { method: "POST" })
  }

  async adminStats(): Promise<AdminStats> {
    return this.request<AdminStats>("/api/admin/stats")
  }

  async adminListUsers(page: number = 0) {
    const resp = await this.requestPaginated<AdminUser[]>(`/api/admin/users?page=${page}`)
    return { ...resp, data: resp.data.map(u => ({ ...u, created_at: new Date(u.created_at) })) }
  }

  async adminCreateUser(username: string, password: string, admin: boolean = false): Promise<AdminUser> {
    const resp = await this.request<AdminUser>("/api/admin/users", { method: "POST", body: { username, password, admin } })
    return { ...resp, created_at: new Date(resp.created_at) }
  }

  async adminGetUser(id: number): Promise<AdminUserDetail> {
    const resp = await this.request<AdminUserDetail>(`/api/admin/users/${id}`)
    return { ...resp, created_at: new Date(resp.created_at) }
  }

  async adminSetUserEnabled(id: number, enabled: boolean): Promise<User> {
    return this.request<User>(`/api/admin/users/${id}/enabled`, { method: "POST", body: { enabled } })
  }

  async adminSetUserAdmin(id: number, admin: boolean): Promise<User> {
    return this.request<User>(`/api/admin/users/${id}/admin`, { method: "POST", body: { admin } })
  }

  async webauthnRegisterBegin(): Promise<PublicKeyCredentialCreationOptionsJSON> {
    return this.request<PublicKeyCredentialCreationOptionsJSON>("/api/webauthn/register/begin", { method: "POST" })
  }

  async webauthnRegisterFinish(response: RegistrationResponseJSON, name: string): Promise<void> {
    return this.request<void>(`/api/webauthn/register/finish?name=${encodeURIComponent(name)}`, {
      method: "POST",
      body: response
    })
  }

  async webauthnLoginBegin(): Promise<PublicKeyCredentialRequestOptionsJSON> {
    return this.request<PublicKeyCredentialRequestOptionsJSON>("/api/webauthn/login/begin", { method: "POST" })
  }

  async webauthnLoginFinish(response: AuthenticationResponseJSON): Promise<User> {
    const resp = await this.request<User>("/api/webauthn/login/finish", { method: "POST", body: response })
    return { ...resp, created_at: new Date(resp.created_at) }
  }

  async webauthnListCredentials(): Promise<Passkey[]> {
    const list = await this.request<Passkey[]>("/api/webauthn/credentials")
    return list.map(p => ({ ...p, created_at: new Date(p.created_at) }))
  }

  async webauthnDeleteCredential(id: string): Promise<void> {
    return this.request<void>(`/api/webauthn/credentials/${id}/delete`, { method: "POST" })
  }

  async uploadImage(imageData: Blob): Promise<{ URL: string }> {
    const formData = new FormData()
    formData.append("file", imageData)

    // only returns url
    return $fetch<{ URL: string }>("/api/upload", { method: "POST", body: formData })
  }
}
