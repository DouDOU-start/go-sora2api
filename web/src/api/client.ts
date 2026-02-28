import axios from 'axios'

const client = axios.create({
  timeout: 30000,
})

// 请求拦截器：自动附加 JWT Token
client.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截器：401 自动跳转登录
client.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401 && !window.location.pathname.includes('/login')) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default client

// 从 Axios 错误中提取后端返回的友好错误信息
export function getErrorMessage(err: unknown, fallback = '操作失败'): string {
  if (axios.isAxiosError(err)) {
    // 后端返回 { "error": "xxx" } 格式
    const data = err.response?.data
    if (data && typeof data === 'object' && 'error' in data) {
      return String((data as { error: string }).error)
    }
    if (data && typeof data === 'object' && 'message' in data) {
      return String((data as { message: string }).message)
    }
  }
  if (err instanceof Error) return err.message
  return fallback
}
