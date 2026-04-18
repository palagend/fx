import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import apiClient from '../api/axios'

export const useUserStore = defineStore('user', () => {
  const user = ref(null)
  const accessToken = ref(localStorage.getItem('accessToken') || '')
  const refreshToken = ref(localStorage.getItem('refreshToken') || '')
  const isLoading = ref(false)
  const error = ref(null)
  const showLoginModal = ref(false)

  const isLoggedIn = computed(() => !!accessToken.value && !!user.value)

  const setTokens = (tokens) => {
    accessToken.value = tokens.access_token
    refreshToken.value = tokens.refresh_token
    localStorage.setItem('accessToken', tokens.access_token)
    localStorage.setItem('refreshToken', tokens.refresh_token)
  }

  const clearTokens = () => {
    accessToken.value = ''
    refreshToken.value = ''
    user.value = null
    localStorage.removeItem('accessToken')
    localStorage.removeItem('refreshToken')
  }

  const register = async (username, email, password) => {
    isLoading.value = true
    error.value = null
    try {
      const response = await apiClient.post('/auth/register', {
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
      const response = await apiClient.post('/auth/login', {
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
      const response = await apiClient.get('/auth/me')
      user.value = response.data.user
    } catch (err) {
      // 401 错误由 axios 拦截器处理
      console.error('获取用户信息失败:', err)
    }
  }

  const logout = async () => {
    if (refreshToken.value) {
      try {
        await apiClient.post('/auth/logout', {
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
      await apiClient.post('/auth/logout-all')
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
      await apiClient.post('/auth/change-password', {
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
      await fetchUserInfo()
    }
  }

  const openLoginModal = () => {
    showLoginModal.value = true
  }

  const closeLoginModal = () => {
    showLoginModal.value = false
  }

  return {
    user,
    accessToken,
    refreshToken,
    isLoading,
    error,
    isLoggedIn,
    showLoginModal,
    register,
    login,
    logout,
    logoutAll,
    fetchUserInfo,
    changePassword,
    init,
    openLoginModal,
    closeLoginModal,
    setTokens,
    clearTokens
  }
})
