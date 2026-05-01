import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { config, userApi, apiClient } from '../api'
import type { User, UserSettings } from '../types'

export const useUserStore = defineStore('user', () => {
  const user = ref<User | null>(null)
  const accessToken = ref<string>(localStorage.getItem('accessToken') || '')
  const refreshToken = ref<string>(localStorage.getItem('refreshToken') || '')
  const isLoading = ref<boolean>(false)
  const error = ref<string | null>(null)
  const showLoginModal = ref<boolean>(false)

  const isLoggedIn = computed(() => {
    if (config.isFrontend) {
      return true
    }
    return !!accessToken.value && !!user.value
  })

  const setTokens = (tokens: { access_token: string; refresh_token: string }) => {
    if (config.isFrontend) {
      return
    }
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

  const backendRegister = async (username: string, email: string, password: string) => {
    const response = await apiClient.post('/auth/register', {
      username,
      email,
      password
    })
    user.value = response.data.user
    setTokens(response.data.tokens)
    return { success: true }
  }

  const backendLogin = async (username: string, password: string) => {
    const response = await apiClient.post('/auth/login', {
      username,
      password
    })
    user.value = response.data.user
    setTokens(response.data.tokens)
    return { success: true }
  }

  const backendFetchUserInfo = async () => {
    if (!accessToken.value) return
    const response = await apiClient.get('/auth/me')
    user.value = response.data.user
  }

  const backendLogout = async () => {
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

  const backendLogoutAll = async () => {
    await apiClient.post('/auth/logout-all')
    clearTokens()
    return { success: true }
  }

  const backendChangePassword = async (oldPassword: string, newPassword: string) => {
    await apiClient.post('/auth/change-password', {
      old_password: oldPassword,
      new_password: newPassword
    })
    return { success: true }
  }

  const register = async (username: string, email: string, password: string) => {
    if (config.isFrontend) {
      return { success: false, error: '前端模式不支持注册' }
    }
    isLoading.value = true
    error.value = null
    try {
      return await backendRegister(username, email, password)
    } catch (err) {
      error.value = err.response?.data?.error || '注册失败'
      return { success: false, error: error.value }
    } finally {
      isLoading.value = false
    }
  }

  const login = async (username: string, password: string) => {
    if (config.isFrontend) {
      return { success: false, error: '前端模式不支持登录' }
    }
    isLoading.value = true
    error.value = null
    try {
      return await backendLogin(username, password)
    } catch (err) {
      error.value = err.response?.data?.error || '登录失败'
      return { success: false, error: error.value }
    } finally {
      isLoading.value = false
    }
  }

  const fetchUserInfo = async () => {
    if (config.isBackend) {
      await backendFetchUserInfo()
    }
  }

  const logout = async () => {
    if (config.isFrontend) {
      return
    }
    await backendLogout()
  }

  const logoutAll = async () => {
    if (config.isFrontend) {
      return { success: false, error: '前端模式不支持此操作' }
    }
    try {
      return await backendLogoutAll()
    } catch (err) {
      return { success: false, error: err.response?.data?.error || '登出失败' }
    }
  }

  const changePassword = async (oldPassword: string, newPassword: string) => {
    if (config.isFrontend) {
      return { success: false, error: '前端模式不支持修改密码' }
    }
    isLoading.value = true
    error.value = null
    try {
      return await backendChangePassword(oldPassword, newPassword)
    } catch (err) {
      error.value = err.response?.data?.error || '修改密码失败'
      return { success: false, error: error.value }
    } finally {
      isLoading.value = false
    }
  }

  const init = async () => {
    if (config.isFrontend) {
      user.value = userApi.getUser()
    } else if (accessToken.value) {
      await backendFetchUserInfo()
    }
  }

  const openLoginModal = () => {
    if (config.isFrontend) {
      return
    }
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