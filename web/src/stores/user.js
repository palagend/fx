import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'

const API_BASE_URL = '/api'

export const useUserStore = defineStore('user', () => {
  const user = ref(null)
  const accessToken = ref(localStorage.getItem('accessToken') || '')
  const refreshToken = ref(localStorage.getItem('refreshToken') || '')
  const isLoading = ref(false)
  const error = ref(null)

  const isLoggedIn = computed(() => !!accessToken.value && !!user.value)

  const setTokens = (tokens) => {
    accessToken.value = tokens.access_token
    refreshToken.value = tokens.refresh_token
    localStorage.setItem('accessToken', tokens.access_token)
    localStorage.setItem('refreshToken', tokens.refresh_token)
    setupAxiosInterceptors()
  }

  const clearTokens = () => {
    accessToken.value = ''
    refreshToken.value = ''
    user.value = null
    localStorage.removeItem('accessToken')
    localStorage.removeItem('refreshToken')
  }

  const setupAxiosInterceptors = () => {
    axios.defaults.headers.common['Authorization'] = `Bearer ${accessToken.value}`
  }

  const register = async (username, email, password) => {
    isLoading.value = true
    error.value = null
    try {
      const response = await axios.post(`${API_BASE_URL}/auth/register`, {
        username,
        email,
        password
      })
      user.value = response.data.user
      setTokens(response.data.tokens)
      return { success: true }
    } catch (err) {
      error.value = err.response?.data?.error || '注册失败'
      return { success: false, error: error.value }
    } finally {
      isLoading.value = false
    }
  }

  const login = async (username, password) => {
    isLoading.value = true
    error.value = null
    try {
      const response = await axios.post(`${API_BASE_URL}/auth/login`, {
        username,
        password
      })
      user.value = response.data.user
      setTokens(response.data.tokens)
      return { success: true }
    } catch (err) {
      error.value = err.response?.data?.error || '登录失败'
      return { success: false, error: error.value }
    } finally {
      isLoading.value = false
    }
  }

  const fetchUserInfo = async () => {
    if (!accessToken.value) return
    try {
      const response = await axios.get(`${API_BASE_URL}/auth/me`)
      user.value = response.data.user
    } catch (err) {
      if (err.response?.status === 401) {
        await refreshAccessToken()
      }
    }
  }

  const refreshAccessToken = async () => {
    if (!refreshToken.value) {
      clearTokens()
      return false
    }
    try {
      const response = await axios.post(`${API_BASE_URL}/auth/refresh`, {
        refresh_token: refreshToken.value
      })
      setTokens(response.data.tokens)
      await fetchUserInfo()
      return true
    } catch (err) {
      clearTokens()
      return false
    }
  }

  const logout = async () => {
    if (refreshToken.value) {
      try {
        await axios.post(`${API_BASE_URL}/auth/logout`, {
          refresh_token: refreshToken.value
        })
      } catch (err) {
        console.error('Logout error:', err)
      }
    }
    clearTokens()
  }

  const logoutAll = async () => {
    try {
      await axios.post(`${API_BASE_URL}/auth/logout-all`)
      clearTokens()
      return { success: true }
    } catch (err) {
      return { success: false, error: err.response?.data?.error || '登出失败' }
    }
  }

  const changePassword = async (oldPassword, newPassword) => {
    isLoading.value = true
    error.value = null
    try {
      await axios.post(`${API_BASE_URL}/auth/change-password`, {
        old_password: oldPassword,
        new_password: newPassword
      })
      return { success: true }
    } catch (err) {
      error.value = err.response?.data?.error || '修改密码失败'
      return { success: false, error: error.value }
    } finally {
      isLoading.value = false
    }
  }

  const init = async () => {
    if (accessToken.value) {
      setupAxiosInterceptors()
      await fetchUserInfo()
    }
  }

  return {
    user,
    accessToken,
    refreshToken,
    isLoading,
    error,
    isLoggedIn,
    register,
    login,
    logout,
    logoutAll,
    fetchUserInfo,
    refreshAccessToken,
    changePassword,
    init
  }
})
