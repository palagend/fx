const STORAGE_KEY = 'local_user'

export interface LocalUser {
  id: string
  username: string
  email: string
  created_at: string
}

const DEFAULT_USER: LocalUser = {
  id: 'local_user',
  username: '本地用户',
  email: 'local@example.com',
  created_at: new Date().toISOString()
}

function getOrCreateUser(): LocalUser {
  const stored = localStorage.getItem(STORAGE_KEY)
  if (stored) {
    return JSON.parse(stored)
  }
  localStorage.setItem(STORAGE_KEY, JSON.stringify(DEFAULT_USER))
  return DEFAULT_USER
}

export const localUserApi = {
  getCurrentUser(): LocalUser {
    return getOrCreateUser()
  },

  isLoggedIn(): boolean {
    return true
  },

  getUser(): LocalUser {
    return getOrCreateUser()
  }
}