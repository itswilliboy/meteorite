import { $fetch, type FetchOptions, type FetchRequest } from "ofetch"
import { type Router } from "vue-router"

import type { APIResponse, DashboardStats } from "./type"

export class HTTPException extends Error {
  status: number

  constructor(status: number, message: APIResponse<any>) {
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

    if (resp.status >= 200 && resp.status < 300) {
      return resp.data
    }

    if (resp.status === 401) {
      if (this.router.currentRoute.value.name !== "login") {
        this.router.push("/login")
      }
    }

    throw new HTTPException(resp.status, resp)
  }

  async login(username: string, password: string): Promise<string> {
    return this.request<string>("/api/login", { method: "POST", body: { username, password } })
  }

  async logout() {
    return this.request("/api/logout", { method: "POST" })
  }

  async ping() {
    return this.request("/api/ping")
  }

  async dashboardStats(): Promise<DashboardStats> {
    return this.request<DashboardStats>("/api/stats")
  }

  async resetToken(): Promise<string> {
    return this.request<string>("/api/reset-token", { method: "POST" })
  }
}
