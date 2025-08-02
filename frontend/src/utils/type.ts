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
