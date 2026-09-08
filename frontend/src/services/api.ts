import { useAuthStore } from '../stores/auth'

const BASE_URL = import.meta.env.VITE_API_URL || '/api/v1'

interface ApiResponse<T> {
  data: T
  status: number
}

interface ApiError extends Error {
  response?: {
    status: number
    data?: any
  }
}

async function request<T>(method: string, path: string, body?: any, isRetry = false): Promise<ApiResponse<T>> {
  const token = useAuthStore.getState().accessToken
  const url = path.startsWith('http') ? path : `${BASE_URL}${path}`

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const res = await fetch(url, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })

  if (res.status === 401 && !isRetry) {
    const refreshToken = useAuthStore.getState().refreshToken
    if (refreshToken) {
      try {
        const refreshRes = await fetch(`${BASE_URL}/auth/refresh`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refresh_token: refreshToken }),
        })
        if (refreshRes.ok) {
          const { access_token, refresh_token } = await refreshRes.json()
          useAuthStore.getState().setTokens(access_token, refresh_token)
          return request<T>(method, path, body, true)
        }
      } catch {}
    }
    useAuthStore.getState().logout()
  }

  const data = res.status !== 204 ? await res.json().catch(() => null) : null

  if (!res.ok) {
    const error: ApiError = new Error(data?.error || `Request failed with status ${res.status}`)
    error.response = { status: res.status, data }
    throw error
  }

  return { data, status: res.status }
}

const api = {
  get: <T = any>(url: string) => request<T>('GET', url),
  post: <T = any>(url: string, data?: any) => request<T>('POST', url, data),
  put: <T = any>(url: string, data?: any) => request<T>('PUT', url, data),
  delete: <T = any>(url: string) => request<T>('DELETE', url),
}

export default api

