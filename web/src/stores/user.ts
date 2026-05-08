import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { userApi } from '../api'
import type { User, UserSettings } from '../types'

const STORAGE_KEY_ACCESS_TOKEN = 'fx_access_token'
const STORAGE_KEY_REFRESH_TOKEN = 'fx_refresh_token'
const STORAGE_KEY_USER = 'fx_user'

export interface LoginResult {
  success: boolean
  error?: string
}

export const useUserStore = defineStore('user', () => {
  const user = ref<User | null>(null)
  const isLoggedIn = computed(() => !!user.value && !!accessToken.value)
  const accessToken = ref<string | null>(null)
  const refreshToken = ref<string | null>(null)
  const isLoading = ref<boolean>(false)
  const error = ref<string | null>(null)
  const showLoginModal = ref(false)
  const showRegisterModal = ref(false)

  const settings = computed<UserSettings>(() => {
    return user.value?.settings || { theme: 'system', currency: 'USD' }
  })

  function setTokens(tokens: { access_token: string; refresh_token: string }): void {
    accessToken.value = tokens.access_token
    refreshToken.value = tokens.refresh_token
    localStorage.setItem(STORAGE_KEY_ACCESS_TOKEN, tokens.access_token)
    localStorage.setItem(STORAGE_KEY_REFRESH_TOKEN, tokens.refresh_token)
  }

  function clearTokens(): void {
    accessToken.value = null
    refreshToken.value = null
    localStorage.removeItem(STORAGE_KEY_ACCESS_TOKEN)
    localStorage.removeItem(STORAGE_KEY_REFRESH_TOKEN)
  }

  function openLoginModal(): void {
    showLoginModal.value = true
  }

  function closeLoginModal(): void {
    showLoginModal.value = false
  }

  function openRegisterModal(): void {
    showRegisterModal.value = true
  }

  function closeRegisterModal(): void {
    showRegisterModal.value = false
  }

  function setUser(newUser: User): void {
    user.value = newUser
    localStorage.setItem(STORAGE_KEY_USER, JSON.stringify(newUser))
  }

  function clearUser(): void {
    user.value = null
    clearTokens()
    localStorage.removeItem(STORAGE_KEY_USER)
  }

  async function register(email: string, username: string, password: string): Promise<LoginResult> {
    isLoading.value = true
    error.value = null

    try {
      const response = await fetch('/api/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, username, password })
      })
      const data = await response.json()
      if (!response.ok) {
        throw new Error(data.error || '注册失败')
      }
      return { success: true }
    } catch (err) {
      const message = (err as Error).message || '注册失败'
      error.value = message
      return { success: false, error: message }
    } finally {
      isLoading.value = false
    }
  }

  async function login(username: string, password: string): Promise<LoginResult> {
    isLoading.value = true
    error.value = null

    try {
      const response = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ Username: username, Password: password })
      })
      const data = await response.json()
      if (!response.ok) {
        throw new Error(data.error || '登录失败')
      }
      setTokens({
        access_token: data.tokens.access_token,
        refresh_token: data.tokens.refresh_token
      })
      setUser(data.user)
      showLoginModal.value = false
      return { success: true }
    } catch (err) {
      const message = (err as Error).message || '登录失败'
      error.value = message
      return { success: false, error: message }
    } finally {
      isLoading.value = false
    }
  }

  async function logout(): Promise<LoginResult> {
    try {
      await fetch('/api/auth/logout', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${accessToken.value}`
        }
      })
    } catch (err) {
      console.error('登出失败:', err)
    } finally {
      clearUser()
      showLoginModal.value = false
    }
    return { success: true }
  }

  async function changePassword(oldPassword: string, newPassword: string): Promise<LoginResult> {
    isLoading.value = true
    error.value = null

    try {
      const response = await fetch('/api/auth/change-password', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${accessToken.value}`
        },
        body: JSON.stringify({ oldPassword, newPassword })
      })
      const data = await response.json()
      if (!response.ok) {
        throw new Error(data.error || '修改密码失败')
      }
      return { success: true }
    } catch (err) {
      const message = (err as Error).message || '修改密码失败'
      error.value = message
      return { success: false, error: message }
    } finally {
      isLoading.value = false
    }
  }

  async function init(): Promise<void> {
    const savedAccessToken = localStorage.getItem(STORAGE_KEY_ACCESS_TOKEN)
    const savedRefreshToken = localStorage.getItem(STORAGE_KEY_REFRESH_TOKEN)
    const savedUser = localStorage.getItem(STORAGE_KEY_USER)

    if (savedAccessToken && savedRefreshToken) {
      accessToken.value = savedAccessToken
      refreshToken.value = savedRefreshToken
      
      if (savedUser) {
        try {
          user.value = JSON.parse(savedUser)
        } catch (err) {
          console.error('解析保存的用户信息失败:', err)
        }
      }
    } else if (userApi) {
      try {
        const localUser = userApi.getUser()
        user.value = {
          id: localUser.id,
          username: localUser.username,
          email: localUser.email,
          created_at: localUser.created_at
        }
      } catch (err) {
        console.error('初始化用户失败:', err)
      }
    }
  }

  async function updateSettings(newSettings: Partial<UserSettings>): Promise<LoginResult> {
    try {
      if (user.value) {
        const currentSettings = user.value.settings || { theme: 'system', currency: 'USD' }
        user.value.settings = {
          ...currentSettings,
          ...newSettings
        }
      }
      return { success: true }
    } catch (err) {
      return { success: false, error: (err as Error).message }
    }
  }

  return {
    user,
    isLoggedIn,
    accessToken,
    refreshToken,
    isLoading,
    error,
    showLoginModal,
    showRegisterModal,
    settings,
    setTokens,
    clearTokens,
    openLoginModal,
    closeLoginModal,
    openRegisterModal,
    closeRegisterModal,
    setUser,
    clearUser,
    register,
    login,
    logout,
    changePassword,
    init,
    updateSettings
  }
})