import { $fetch, type FetchOptions, type FetchRequest } from "ofetch"
import { type Router } from "vue-router"

type APIResponse<T> = {
  status: number
  data: T
}

export class Client {
  constructor(private router: Router) {}

  async request<T>(request: FetchRequest, options?: FetchOptions<"json">): Promise<T> {
    const resp = await $fetch<APIResponse<T>>(request, {
      ...options,
      headers: { ...options?.headers, Authorisation: localStorage.getItem("token") || "" },
      ignoreResponseError: true
    })

    // TODO: Maybe cleanup
    const range = Math.floor(resp.status / 100)
    switch (range) {
      case 2:
        return resp.data

      case 4:
        if (resp.status == 401) {
          console.log(this.router.currentRoute)
          if (this.router.currentRoute.value.name !== "login") {
            localStorage.removeItem("token")
            this.router.push("/login")
          }
          return resp.data
        }

        console.log("Error status 4xx", resp.status)

      default:
        console.log(console.log("Error status", resp.status))
    }

    return resp.data
  }

  async dashboardStats() {
    return this.request<any>("/api/stats")
  }

  async login(username: string, password: string): Promise<string> {
    return this.request<string>("/api/login", { method: "POST", body: { username, password } })
  }
}
