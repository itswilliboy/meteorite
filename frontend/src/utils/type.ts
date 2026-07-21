export type APIResponse<T> = {
  status: number
  data: T
}

export type PaginatedResponse<T> = {
  status: number
  data: T
  page: number
  pageSize: number
  hasNext: boolean
  hasPrev: boolean
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

export type Image = {
  id: string
  date: string
  mimetype: string
  views: number
  url: string
  filename: string | null
  size: number
  width: number | null
  height: number | null
  duration_ms: number | null
  bitrate: number | null
  codec: string | null
  framerate: number | null
  sample_rate: number | null
  channels: number | null
  has_cover: boolean
}

export type Colour = "info" | "success" | "warning" | "danger"

export type Toast = {
  id: number
  title: string
  desc?: string
  colour: Colour
  delay: number
  popped: boolean
}
