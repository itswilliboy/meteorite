export type APIResponse<T> = {
  status: number
  data: T
}

export type DashboardStats = {
  total_images: number
  storage_usage: number
  monthly_uploads: number
  user_bandwidth: number
}

export type User = {
  id: number
  name: string
  created_at: Date
  enabled: boolean
  admin: boolean
}

export type Toast = {
  id: number
  title: string
  desc?: string
  colour: "info" | "success" | "warning" | "danger"
  delay: number
  popped: boolean
}
