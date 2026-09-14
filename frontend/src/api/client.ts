import { getStoredToken, removeStoredToken, removeStoredUser } from '../utils/storage.ts'

const BASE_URL = (import.meta.env.VITE_API_BASE_URL as string) || ''

export interface RequestOptions extends RequestInit {
  timeoutMs?: number
  params?: Record<string, string | number | boolean | undefined>
}

export class ApiError extends Error {
  status: number
  code?: string
  details?: unknown

  constructor(message: string, status: number, code?: string, details?: unknown) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.details = details
  }
}

export async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { timeoutMs = 15000, params, ...customConfig } = options

  let fullPath = path.startsWith('http') ? path : `${BASE_URL}${path}`

  if (params) {
    const queryParams = new URLSearchParams()
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined) {
        queryParams.append(key, String(value))
      }
    })
    const queryString = queryParams.toString()
    if (queryString) {
      fullPath += (fullPath.includes('?') ? '&' : '?') + queryString
    }
  }

  const token = getStoredToken()
  const headers = new Headers(customConfig.headers || {})
  headers.set('Accept', 'application/json')

  if (!(customConfig.body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  if (token && !headers.has('Authorization')) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  const controller = new AbortController()
  const timer = setTimeout(() => controller.abort(), timeoutMs)

  try {
    const response = await fetch(fullPath, {
      ...customConfig,
      headers,
      signal: controller.signal,
      credentials: 'include',
    })

    if (response.status === 401 && !path.includes('/auth/login')) {
      removeStoredToken()
      removeStoredUser()
      if (window.location.pathname !== '/login') {
        window.location.href = `/login?redirect=${encodeURIComponent(window.location.pathname)}`
      }
    }

    const contentType = response.headers.get('content-type')
    const isJson = contentType && contentType.includes('application/json')
    const data = isJson ? await response.json() : await response.text()

    if (!response.ok) {
      let message = response.statusText || 'Request failed'
      let code: string | undefined
      let details: unknown

      if (typeof data === 'object' && data !== null) {
        if ('error' in data) {
          if (typeof data.error === 'string') {
            message = data.error
          } else if (typeof data.error === 'object' && data.error !== null) {
            message = data.error.message || message
            code = data.error.code
            details = data.error.details
          }
        } else if ('message' in data && typeof data.message === 'string') {
          message = data.message
        }
      }

      throw new ApiError(message, response.status, code, details)
    }

    if (typeof data === 'object' && data !== null && 'success' in data && 'data' in data) {
      return data.data as T
    }

    return data as T
  } catch (error: unknown) {
    if (error instanceof ApiError) {
      throw error
    }
    if (error instanceof DOMException && error.name === 'AbortError') {
      throw new ApiError('Request timeout after ' + timeoutMs + 'ms', 408)
    }
    const message = error instanceof Error ? error.message : 'Network failure'
    throw new ApiError(message, 0)
  } finally {
    clearTimeout(timer)
  }
}

export const api = {
  get: <T>(path: string, options?: RequestOptions) => request<T>(path, { ...options, method: 'GET' }),
  post: <T>(path: string, body?: unknown, options?: RequestOptions) =>
    request<T>(path, {
      ...options,
      method: 'POST',
      body: body instanceof FormData ? body : body ? JSON.stringify(body) : undefined,
    }),
  put: <T>(path: string, body?: unknown, options?: RequestOptions) =>
    request<T>(path, {
      ...options,
      method: 'PUT',
      body: body ? JSON.stringify(body) : undefined,
    }),
  patch: <T>(path: string, body?: unknown, options?: RequestOptions) =>
    request<T>(path, {
      ...options,
      method: 'PATCH',
      body: body ? JSON.stringify(body) : undefined,
    }),
  delete: <T>(path: string, options?: RequestOptions) => request<T>(path, { ...options, method: 'DELETE' }),
}
