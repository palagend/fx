import axios, { AxiosInstance, AxiosResponse, InternalAxiosRequestConfig, AxiosHeaders } from 'axios'
import { config } from '../config'

const API_BASE_URL = '/api'

const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

let isRefreshing = false
let refreshSubscribers: ((newToken: string) => void)[] = []
let getAccessToken: (() => string | null) | null = null
let getRefreshToken: (() => string | null) | null = null
let setTokensCallback: ((tokens: { access_token: string; refresh_token: string }) => void) | null = null
let clearTokensCallback: (() => void) | null = null
let openLoginModalCallback: (() => void) | null = null

export function setupAuthCallbacks(
  getAccess: () => string | null,
  getRefresh: () => string | null,
  setTokens: (tokens: { access_token: string; refresh_token: string }) => void,
  clearTokens: () => void,
  openLoginModal: () => void
): void {
  getAccessToken = getAccess
  getRefreshToken = getRefresh
  setTokensCallback = setTokens
  clearTokensCallback = clearTokens
  openLoginModalCallback = openLoginModal
}

function subscribeTokenRefresh(callback: (newToken: string) => void): void {
  refreshSubscribers.push(callback)
}

function onTokenRefreshed(newToken: string): void {
  refreshSubscribers.forEach(callback => callback(newToken))
  refreshSubscribers = []
}

apiClient.interceptors.request.use(
  (requestConfig: InternalAxiosRequestConfig): InternalAxiosRequestConfig => {
    if (config.isFrontend) {
      return requestConfig
    }
    
    if (!getAccessToken) {
      console.warn('Auth callbacks not initialized')
      return requestConfig
    }
    
    const accessToken = getAccessToken()
    if (accessToken) {
      requestConfig.headers = requestConfig.headers || {}
      requestConfig.headers.Authorization = `Bearer ${accessToken}`
    }
    return requestConfig
  },
  (error: unknown): Promise<never> => {
    return Promise.reject(error)
  }
)

apiClient.interceptors.response.use(
  (response: AxiosResponse): AxiosResponse => {
    return response
  },
  async (error: unknown): Promise<never> => {
    const axiosError = error as { 
      response?: { status?: number; data?: { error?: string } }; 
      config?: InternalAxiosRequestConfig;
      code?: string
    }
    const originalRequest = axiosError.config

    if (config.isFrontend) {
      return Promise.reject(error)
    }

    if (axiosError.response?.status !== 401) {
      return Promise.reject(error)
    }

    if (!getRefreshToken || !clearTokensCallback || !openLoginModalCallback) {
      console.warn('Auth callbacks not initialized')
      return Promise.reject(error)
    }

    const refreshToken = getRefreshToken()
    if (!refreshToken) {
      clearTokensCallback()
      openLoginModalCallback()
      return Promise.reject(error)
    }

    if (originalRequest?.headers?.['X-Retry-Count']) {
      clearTokensCallback()
      openLoginModalCallback()
      return Promise.reject(error)
    }

    if (isRefreshing && originalRequest) {
      return new Promise((resolve) => {
        subscribeTokenRefresh((newToken: string) => {
          if (!originalRequest.headers) {
            originalRequest.headers = new AxiosHeaders()
          }
          originalRequest.headers.Authorization = `Bearer ${newToken}`
          resolve(apiClient(originalRequest))
        })
      })
    }

    isRefreshing = true

    try {
      const response = await axios.post(`${API_BASE_URL}/auth/refresh`, {
        refresh_token: refreshToken
      }, {
        headers: { 'Content-Type': 'application/json' }
      })

      const responseData = response.data as { 
        access_token?: string; 
        refresh_token?: string;
        tokens?: { access_token?: string; refresh_token?: string }
      }

      const newToken = responseData.access_token || responseData.tokens?.access_token
      const newRefreshToken = responseData.refresh_token || responseData.tokens?.refresh_token

      if (!newToken) {
        throw new Error('刷新令牌失败')
      }

      if (setTokensCallback) {
        setTokensCallback({
          access_token: newToken,
          refresh_token: newRefreshToken || refreshToken
        })
      }

      onTokenRefreshed(newToken)

      if (originalRequest) {
        if (!originalRequest.headers) {
          originalRequest.headers = new AxiosHeaders()
        }
        originalRequest.headers.Authorization = `Bearer ${newToken}`
        originalRequest.headers.set('X-Retry-Count', '1')
        
        return apiClient(originalRequest)
      }
      return Promise.reject(error)
    } catch (refreshError) {
      clearTokensCallback()
      openLoginModalCallback()
      console.error('登录已过期，请重新登录')
      return Promise.reject(refreshError)
    } finally {
      isRefreshing = false
    }
  }
)

export { apiClient }