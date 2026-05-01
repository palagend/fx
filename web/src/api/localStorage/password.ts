interface PasswordData {
  id: string
  password: string
  title: string
  username: string
  url: string
  tags: string[]
  strength: number
  length: number
  createdAt: number
  updatedAt: number
  useCount: number
  lastUsedAt: number | null
}

interface GetPasswordsParams {
  search?: string
  tag?: string
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
}

interface GetPasswordsResponse {
  data: {
    passwords: PasswordData[]
    total: number
  }
  error?: string
  status?: number
}

interface GetPasswordResponse {
  data?: PasswordData
  error?: string
  status?: number
}

interface CreatePasswordRequest {
  title?: string
  username?: string
  password: string
  url?: string
  tags?: string[]
  strength?: number
  length?: number
}

interface CreatePasswordResponse {
  data?: PasswordData
  error?: string
  status?: number
}

interface UpdatePasswordRequest {
  title?: string
  username?: string
  password?: string
  url?: string
  tags?: string[]
}

interface UpdatePasswordResponse {
  data?: PasswordData
  error?: string
  status?: number
}

interface DeletePasswordResponse {
  data?: { success: boolean }
  error?: string
  status?: number
}

interface RecordPasswordUseResponse {
  data?: { success: boolean }
  error?: string
  status?: number
}

interface GetTagsResponse {
  data: {
    tags: string[]
  }
  error?: string
  status?: number
}

interface ExportPasswordsResponse {
  data: {
    passwords: PasswordData[]
    exportAt: number
  }
  error?: string
  status?: number
}

interface ImportPasswordsResponse {
  data: { success: boolean }
  error?: string
  status?: number
}

interface GetSettingsResponse {
  data: {
    defaultLength: number
    autoSave: boolean
    defaultTags: string[]
  }
  error?: string
  status?: number
}

interface SaveSettingsRequest {
  defaultLength?: number
  autoSave?: boolean
  defaultTags?: string[]
}

interface SaveSettingsResponse {
  data: {
    defaultLength: number
    autoSave: boolean
    defaultTags: string[]
  }
  error?: string
  status?: number
}

const STORAGE_KEYS = {
  PASSWORDS: 'password_manager_passwords',
  SETTINGS: 'password_manager_settings'
}

function getStorageData<T>(key: string, defaultValue: T): T {
  try {
    const data = localStorage.getItem(key)
    return data ? JSON.parse(data) : defaultValue
  } catch (error) {
    console.error(`读取 ${key} 失败:`, error)
    return defaultValue
  }
}

function setStorageData(key: string, value: unknown): boolean {
  try {
    localStorage.setItem(key, JSON.stringify(value))
    return true
  } catch (error) {
    console.error(`保存 ${key} 失败:`, error)
    return false
  }
}

function generateUUID(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0
    const v = c === 'x' ? r : (r & 0x3 | 0x8)
    return v.toString(16)
  })
}

function encryptPassword(password: string): string {
  return btoa(password)
}

function decryptPassword(encrypted: string): string {
  try {
    return atob(encrypted)
  } catch {
    return encrypted
  }
}

export function getPasswords(params: GetPasswordsParams = {}): GetPasswordsResponse {
  const { search, tag, sortBy = 'createdAt', sortOrder = 'desc' } = params

  let passwords = getStorageData<PasswordData[]>(STORAGE_KEYS.PASSWORDS, [])

  passwords = passwords.map(p => ({
    ...p,
    password: decryptPassword(p.password)
  }))

  if (search) {
    const searchLower = search.toLowerCase()
    passwords = passwords.filter(p =>
      p.title?.toLowerCase().includes(searchLower) ||
      p.username?.toLowerCase().includes(searchLower) ||
      p.url?.toLowerCase().includes(searchLower) ||
      p.tags?.some(t => t.toLowerCase().includes(searchLower))
    )
  }

  if (tag) {
    passwords = passwords.filter(p => p.tags?.includes(tag))
  }

  passwords.sort((a, b) => {
    const aVal = a[sortBy as keyof PasswordData] || 0
    const bVal = b[sortBy as keyof PasswordData] || 0
    return sortOrder === 'asc' ? (aVal > bVal ? 1 : -1) : (aVal < bVal ? 1 : -1)
  })

  return {
    data: {
      passwords,
      total: passwords.length
    }
  }
}

export function getPassword(id: string): GetPasswordResponse {
  const passwords = getStorageData<PasswordData[]>(STORAGE_KEYS.PASSWORDS, [])
  const password = passwords.find(p => p.id === id)

  if (!password) {
    return { error: '密码不存在', status: 404 }
  }

  return {
    data: {
      ...password,
      password: decryptPassword(password.password)
    }
  }
}

export function createPassword(data: CreatePasswordRequest): CreatePasswordResponse {
  const passwords = getStorageData<PasswordData[]>(STORAGE_KEYS.PASSWORDS, [])

  const newPassword: PasswordData = {
    id: generateUUID(),
    password: encryptPassword(data.password),
    title: data.title || '',
    username: data.username || '',
    url: data.url || '',
    tags: data.tags || [],
    strength: data.strength || 0,
    length: data.length || data.password?.length || 0,
    createdAt: Date.now(),
    updatedAt: Date.now(),
    useCount: 0,
    lastUsedAt: null
  }

  passwords.push(newPassword)

  if (!setStorageData(STORAGE_KEYS.PASSWORDS, passwords)) {
    return { error: '保存失败', status: 500 }
  }

  return {
    data: {
      ...newPassword,
      password: data.password
    }
  }
}

export function updatePassword(id: string, data: UpdatePasswordRequest): UpdatePasswordResponse {
  const passwords = getStorageData<PasswordData[]>(STORAGE_KEYS.PASSWORDS, [])
  const index = passwords.findIndex(p => p.id === id)

  if (index === -1) {
    return { error: '密码不存在', status: 404 }
  }

  const updatedPassword: PasswordData = {
    ...passwords[index],
    ...data,
    password: data.password ? encryptPassword(data.password) : passwords[index].password,
    updatedAt: Date.now()
  }

  passwords[index] = updatedPassword

  if (!setStorageData(STORAGE_KEYS.PASSWORDS, passwords)) {
    return { error: '保存失败', status: 500 }
  }

  return {
    data: {
      ...updatedPassword,
      password: data.password || decryptPassword(passwords[index].password)
    }
  }
}

export function deletePassword(id: string): DeletePasswordResponse {
  let passwords = getStorageData<PasswordData[]>(STORAGE_KEYS.PASSWORDS, [])
  const index = passwords.findIndex(p => p.id === id)

  if (index === -1) {
    return { error: '密码不存在', status: 404 }
  }

  passwords.splice(index, 1)

  if (!setStorageData(STORAGE_KEYS.PASSWORDS, passwords)) {
    return { error: '删除失败', status: 500 }
  }

  return { data: { success: true } }
}

export function recordPasswordUse(id: string): RecordPasswordUseResponse {
  const passwords = getStorageData<PasswordData[]>(STORAGE_KEYS.PASSWORDS, [])
  const index = passwords.findIndex(p => p.id === id)

  if (index === -1) {
    return { error: '密码不存在', status: 404 }
  }

  passwords[index].useCount = (passwords[index].useCount || 0) + 1
  passwords[index].lastUsedAt = Date.now()

  setStorageData(STORAGE_KEYS.PASSWORDS, passwords)

  return { data: { success: true } }
}

export function getTags(): GetTagsResponse {
  const passwords = getStorageData<PasswordData[]>(STORAGE_KEYS.PASSWORDS, [])
  const tagsSet = new Set<string>()

  passwords.forEach(p => {
    p.tags?.forEach(tag => tagsSet.add(tag))
  })

  return {
    data: {
      tags: Array.from(tagsSet).sort()
    }
  }
}

export function exportPasswords(): ExportPasswordsResponse {
  const passwords = getStorageData<PasswordData[]>(STORAGE_KEYS.PASSWORDS, [])
  const decrypted = passwords.map(p => ({
    ...p,
    password: decryptPassword(p.password)
  }))

  return {
    data: {
      passwords: decrypted,
      exportAt: Date.now()
    }
  }
}

export function importPasswords(passwordsData: PasswordData[], overwrite = false): ImportPasswordsResponse {
  if (overwrite) {
    const encrypted = passwordsData.map(p => ({
      ...p,
      password: encryptPassword(p.password),
      updatedAt: Date.now()
    }))
    setStorageData(STORAGE_KEYS.PASSWORDS, encrypted)
  } else {
    const existing = getStorageData<PasswordData[]>(STORAGE_KEYS.PASSWORDS, [])
    const encrypted = passwordsData.map(p => ({
      ...p,
      id: generateUUID(),
      password: encryptPassword(p.password),
      createdAt: Date.now(),
      updatedAt: Date.now()
    }))
    setStorageData(STORAGE_KEYS.PASSWORDS, [...existing, ...encrypted])
  }

  return { data: { success: true } }
}

export function clearAllPasswords(): ImportPasswordsResponse {
  setStorageData(STORAGE_KEYS.PASSWORDS, [])
  return { data: { success: true } }
}

export function getSettings(): GetSettingsResponse {
  const settings = getStorageData(STORAGE_KEYS.SETTINGS, {
    defaultLength: 16,
    autoSave: false,
    defaultTags: []
  })

  return { data: settings }
}

export function saveSettings(settings: SaveSettingsRequest): SaveSettingsResponse {
  const current = getStorageData(STORAGE_KEYS.SETTINGS, {
    defaultLength: 16,
    autoSave: false,
    defaultTags: []
  })
  const updated = { ...current, ...settings }
  setStorageData(STORAGE_KEYS.SETTINGS, updated)
  return { data: updated }
}