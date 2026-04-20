import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { config, userApi, apiClient } from '../api'

export const useUserStore = defineStore('user', () => {
  // ========== 状态 ==========
  const user = ref(null)
  const accessToken = ref(localStorage.getItem('accessToken') || '')
  const refreshToken = ref(localStorage.getItem('refreshToken') || '')
  const isLoading = ref(false)
  const error = ref(null)
  const showLoginModal = ref(false)

  // ========== 计算属性 ==========
  const isLoggedIn = computed(() => {
    if (config.isFrontend) {
      return userApi.isLoggedIn() && !!user.value
    }
    return !!accessToken.value && !!user.value
  })

  // ========== 辅助函数 ==========
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

  // ========== 后端模式 API 调用 ==========
  const backendRegister = async (username, email, password) => {
    const response = await apiClient.post('/auth/register', {
      username,
      email,
      password
    })
    user.value = response.data.user
    setTokens(response.data.tokens)
    return { success: true }
  }

  const backendLogin = async (username, password) => {
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

  const backendChangePassword = async (oldPassword, newPassword) => {
    await apiClient.post('/auth/change-password', {
      old_password: oldPassword,
      new_password: newPassword
    })
    return { success: true }
  }

  // ========== 前端模式 API 调用 ==========
  const frontendRegister = async (username, email, password) => {
    const response = await userApi.register(username, email, password)
    user.value = response.data.user
    setTokens(response.data.tokens)
    return { success: true }
  }

  const frontendLogin = async (username, password) => {
    const response = await userApi.login(username, password)
    user.value = response.data.user
    setTokens(response.data.tokens)
    return { success: true }
  }

  const frontendFetchUserInfo = async () => {
    try {
      const response = await userApi.getCurrentUser()
      user.value = response.data.user
    } catch (err) {
      // 未登录或 token 过期
      clearTokens()
    }
  }

  const frontendLogout = async () => {
    await userApi.logout()
    clearTokens()
  }

  const frontendLogoutAll = async () => {
    await userApi.logoutAll()
    clearTokens()
    return { success: true }
  }

  const frontendChangePassword = async (oldPassword, newPassword) => {
    await userApi.changePassword(oldPassword, newPassword)
    return { success: true }
  }

  // ========== 统一的 Actions ==========
  const register = async (username, email, password) => {
    isLoading.value = true
    error.value = null
    try {
      if (config.isBackend) {
        return await backendRegister(username, email, password)
      } else {
        return await frontendRegister(username, email, password)
      }
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
      if (config.isBackend) {
        return await backendLogin(username, password)
      } else {
        return await frontendLogin(username, password)
      }
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
    } else {
      await frontendFetchUserInfo()
    }
  }

  const logout = async () => {
    if (config.isBackend) {
      await backendLogout()
    } else {
      await frontendLogout()
    }
  }

  const logoutAll = async () => {
    try {
      if (config.isBackend) {
        return await backendLogoutAll()
      } else {
        return await frontendLogoutAll()
      }
    } catch (err) {
      return { success: false, error: err.response?.data?.error || '登出失败' }
    }
  }

  const changePassword = async (oldPassword, newPassword) => {
    isLoading.value = true
    error.value = null
    try {
      if (config.isBackend) {
        return await backendChangePassword(oldPassword, newPassword)
      } else {
        return await frontendChangePassword(oldPassword, newPassword)
      }
    } catch (err) {
      error.value = err.response?.data?.error || '修改密码失败'
      return { success: false, error: error.value }
    } finally {
      isLoading.value = false
    }
  }

  const init = async () => {
    if (config.isFrontend) {
      // 前端模式：检查 localStorage 中的登录状态
      if (userApi.isLoggedIn()) {
        const localUser = userApi.getUser()
        if (localUser) {
          user.value = localUser
          // 设置模拟 token
          accessToken.value = 'local_mock_token'
          refreshToken.value = 'local_mock_refresh_token'
        }
      }
    } else if (accessToken.value) {
      // 后端模式：通过 token 获取用户信息
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
