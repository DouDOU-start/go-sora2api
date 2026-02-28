// 通用 API 响应
export interface ApiResponse<T> {
  code?: number
  message?: string
  data?: T
  error?: string
}

// 分页响应
export interface PageResponse<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}
