import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'
import { useUserStore } from '../stores/user'
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

function subscribeTokenRefresh(callback: (newToken: string) => void): void {
  refreshSubscribers.push(callback)
}

function onTokenRefreshed(newToken: string): void {
  refreshSubscribers.forEach(callback => callback(newToken))
  refreshSubscribers = []
}

apiClient.interceptors.request.use(
  (requestConfig: AxiosRequestConfig): AxiosRequestConfig => {
    if (config.isFrontend) {
      return requestConfig
    }
    const userStore = useUserStore()
    if (userStore.accessToken) {
      requestConfig.headers = requestConfig.headers || {}
      requestConfig.headers.Authorization = `Bearer ${userStore.accessToken}`
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
    const axiosError = error as { response?: { status?: number }; config?: AxiosRequestConfig }
    const originalRequest = axiosError.config

    if (config.isFrontend) {
      return Promise.reject(error)
    }

    if (axiosError.response?.status !== 401) {
      return Promise.reject(error)
    }

    const userStore = useUserStore()

    if (!userStore.refreshToken) {
      userStore.clearTokens()
      userStore.openLoginModal()
      return Promise.reject(error)
    }

    if (isRefreshing && originalRequest) {
      return new Promise((resolve) => {
        subscribeTokenRefresh((newToken: string) => {
          if (originalRequest.headers) {
            originalRequest.headers.Authorization = `Bearer ${newToken}`
          }
          resolve(apiClient(originalRequest))
        })
      })
    }

    isRefreshing = true

    try {
      const response = await axios.post(`${API_BASE_URL}/auth/refresh`, {
        refresh_token: userStore.refreshToken
      })

      const newToken = response.data.tokens.access_token
      const newRefreshToken = response.data.tokens.refresh_token

      userStore.setTokens({
        access_token: newToken,
        refresh_token: newRefreshToken
      })

      onTokenRefreshed(newToken)

      if (originalRequest?.headers) {
        originalRequest.headers.Authorization = `Bearer ${newToken}`
      }
      if (originalRequest) {
        return apiClient(originalRequest)
      }
      return Promise.reject(error)
    } catch (refreshError) {
      userStore.clearTokens()
      userStore.openLoginModal()
      console.error('登录已过期，请重新登录')
      return Promise.reject(refreshError)
    } finally {
      isRefreshing = false
    }
  }
)

export { apiClient }