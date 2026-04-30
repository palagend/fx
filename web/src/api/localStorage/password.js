// LocalStorage 版本的 Password API（纯前端模式）
// 预留后端接口结构，方便后续切换

const STORAGE_KEYS = {
  PASSWORDS: 'password_manager_passwords',
  SETTINGS: 'password_manager_settings'
}

// ========== 辅助函数 ==========

// 获取存储的数据
function getStorageData(key, defaultValue = []) {
  try {
    const data = localStorage.getItem(key)
    return data ? JSON.parse(data) : defaultValue
  } catch (error) {
    console.error(`读取 ${key} 失败:`, error)
    return defaultValue
  }
}

// 设置存储的数据
function setStorageData(key, value) {
  try {
    localStorage.setItem(key, JSON.stringify(value))
    return true
  } catch (error) {
    console.error(`保存 ${key} 失败:`, error)
    return false
  }
}

// 生成 UUID
function generateUUID() {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0
    const v = c === 'x' ? r : (r & 0x3 | 0x8)
    return v.toString(16)
  })
}

// 加密密码（简单 Base64，生产环境应使用 AES）
function encryptPassword(password) {
  // TODO: 生产环境使用 CryptoJS.AES 加密
  return btoa(password)
}

// 解密密码
function decryptPassword(encrypted) {
  // TODO: 生产环境使用 CryptoJS.AES 解密
  try {
    return atob(encrypted)
  } catch {
    return encrypted
  }
}

// ========== 密码管理 API ==========

/**
 * 获取所有密码
 * @param {Object} params - 查询参数
 * @param {string} params.search - 搜索关键词
 * @param {string} params.tag - 标签筛选
 * @param {string} params.sortBy - 排序字段
 * @param {string} params.sortOrder - 排序方向 (asc/desc)
 */
export function getPasswords(params = {}) {
  const { search, tag, sortBy = 'createdAt', sortOrder = 'desc' } = params

  let passwords = getStorageData(STORAGE_KEYS.PASSWORDS, [])

  // 解密密码
  passwords = passwords.map(p => ({
    ...p,
    password: decryptPassword(p.password)
  }))

  // 搜索筛选
  if (search) {
    const searchLower = search.toLowerCase()
    passwords = passwords.filter(p =>
      p.title?.toLowerCase().includes(searchLower) ||
      p.username?.toLowerCase().includes(searchLower) ||
      p.url?.toLowerCase().includes(searchLower) ||
      p.tags?.some(t => t.toLowerCase().includes(searchLower))
    )
  }

  // 标签筛选
  if (tag) {
    passwords = passwords.filter(p => p.tags?.includes(tag))
  }

  // 排序
  passwords.sort((a, b) => {
    const aVal = a[sortBy] || 0
    const bVal = b[sortBy] || 0
    return sortOrder === 'asc' ? (aVal > bVal ? 1 : -1) : (aVal < bVal ? 1 : -1)
  })

  return {
    data: {
      passwords,
      total: passwords.length
    }
  }
}

/**
 * 获取单个密码
 * @param {string} id - 密码 ID
 */
export function getPassword(id) {
  const passwords = getStorageData(STORAGE_KEYS.PASSWORDS, [])
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

/**
 * 创建密码
 * @param {Object} data - 密码数据
 */
export function createPassword(data) {
  const passwords = getStorageData(STORAGE_KEYS.PASSWORDS, [])

  const newPassword = {
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

/**
 * 更新密码
 * @param {string} id - 密码 ID
 * @param {Object} data - 更新数据
 */
export function updatePassword(id, data) {
  const passwords = getStorageData(STORAGE_KEYS.PASSWORDS, [])
  const index = passwords.findIndex(p => p.id === id)

  if (index === -1) {
    return { error: '密码不存在', status: 404 }
  }

  const updatedPassword = {
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

/**
 * 删除密码
 * @param {string} id - 密码 ID
 */
export function deletePassword(id) {
  let passwords = getStorageData(STORAGE_KEYS.PASSWORDS, [])
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

/**
 * 记录密码使用
 * @param {string} id - 密码 ID
 */
export function recordPasswordUse(id) {
  const passwords = getStorageData(STORAGE_KEYS.PASSWORDS, [])
  const index = passwords.findIndex(p => p.id === id)

  if (index === -1) {
    return { error: '密码不存在', status: 404 }
  }

  passwords[index].useCount = (passwords[index].useCount || 0) + 1
  passwords[index].lastUsedAt = Date.now()

  setStorageData(STORAGE_KEYS.PASSWORDS, passwords)

  return { data: { success: true } }
}

/**
 * 获取所有标签
 */
export function getTags() {
  const passwords = getStorageData(STORAGE_KEYS.PASSWORDS, [])
  const tagsSet = new Set()

  passwords.forEach(p => {
    p.tags?.forEach(tag => tagsSet.add(tag))
  })

  return {
    data: {
      tags: Array.from(tagsSet).sort()
    }
  }
}

/**
 * 导出所有密码
 */
export function exportPasswords() {
  const passwords = getStorageData(STORAGE_KEYS.PASSWORDS, [])
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

/**
 * 导入密码
 * @param {Array} passwords - 密码数组
 * @param {boolean} overwrite - 是否覆盖现有数据
 */
export function importPasswords(passwords, overwrite = false) {
  if (overwrite) {
    const encrypted = passwords.map(p => ({
      ...p,
      password: encryptPassword(p.password),
      updatedAt: Date.now()
    }))
    setStorageData(STORAGE_KEYS.PASSWORDS, encrypted)
  } else {
    const existing = getStorageData(STORAGE_KEYS.PASSWORDS, [])
    const encrypted = passwords.map(p => ({
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

/**
 * 清空所有密码
 */
export function clearAllPasswords() {
  setStorageData(STORAGE_KEYS.PASSWORDS, [])
  return { data: { success: true } }
}

// ========== 设置管理 ==========

/**
 * 获取设置
 */
export function getSettings() {
  const settings = getStorageData(STORAGE_KEYS.SETTINGS, {
    defaultLength: 16,
    autoSave: false,
    defaultTags: []
  })

  return { data: settings }
}

/**
 * 保存设置
 */
export function saveSettings(settings) {
  const current = getStorageData(STORAGE_KEYS.SETTINGS, {})
  const updated = { ...current, ...settings }
  setStorageData(STORAGE_KEYS.SETTINGS, updated)
  return { data: updated }
}
