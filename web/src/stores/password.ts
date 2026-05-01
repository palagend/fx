import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { passwordApi } from '../api'

export interface PasswordEntry {
  id: string
  title: string
  username: string
  password: string
  url?: string
  tags?: string[]
  notes?: string
  createdAt?: number
  useCount?: number
  lastUsedAt?: number
}

export interface PasswordSettings {
  defaultLength: number
  autoSave: boolean
  defaultTags: string[]
}

export interface PasswordResult {
  success: boolean
  error?: string
  data?: PasswordEntry | PasswordEntry[] | unknown
  count?: number
}

export const usePasswordStore = defineStore('password', () => {
  const passwords = ref<PasswordEntry[]>([])
  const tags = ref<string[]>([])
  const settings = ref<PasswordSettings>({
    defaultLength: 16,
    autoSave: false,
    defaultTags: []
  })
  const isLoading = ref<boolean>(false)
  const error = ref<string | null>(null)
  const searchQuery = ref<string>('')
  const selectedTag = ref<string>('')

  const cachedFilteredPasswords = ref<PasswordEntry[]>([])
  const cachedRecentPasswords = ref<PasswordEntry[]>([])
  const cachedMostUsedPasswords = ref<PasswordEntry[]>([])

  const filteredPasswords = computed(() => {
    if (!searchQuery.value && !selectedTag.value) {
      return passwords.value
    }

    if (!searchQuery.value && selectedTag.value) {
      return passwords.value.filter(p => p.tags?.includes(selectedTag.value))
    }

    return cachedFilteredPasswords.value
  })

  const totalPasswords = computed(() => passwords.value.length)

  const recentPasswords = computed(() => cachedRecentPasswords.value)
  const mostUsedPasswords = computed(() => cachedMostUsedPasswords.value)

  function updateCachedValues() {
    const query = searchQuery.value.toLowerCase()
    
    if (searchQuery.value || selectedTag.value) {
      cachedFilteredPasswords.value = passwords.value.filter(p => {
        const matchesSearch = !query || 
          p.title?.toLowerCase().includes(query) ||
          p.username?.toLowerCase().includes(query) ||
          p.url?.toLowerCase().includes(query) ||
          p.tags?.some(t => t.toLowerCase().includes(query))
        
        const matchesTag = !selectedTag.value || p.tags?.includes(selectedTag.value)
        
        return matchesSearch && matchesTag
      })
    }

    cachedRecentPasswords.value = [...passwords.value]
      .sort((a, b) => (b.createdAt || 0) - (a.createdAt || 0))
      .slice(0, 5)

    cachedMostUsedPasswords.value = [...passwords.value]
      .sort((a, b) => (b.useCount || 0) - (a.useCount || 0))
      .slice(0, 5)
  }

  function reconcilePasswords(newPasswords: PasswordEntry[]) {
    const existingIds = new Set(passwords.value.map(p => p.id))
    const newIds = new Set(newPasswords.map(p => p.id))

    const toAdd = newPasswords.filter(p => !existingIds.has(p.id))
    const toRemove = passwords.value.filter(p => !newIds.has(p.id))

    toRemove.forEach(p => removePassword(p.id))
    
    for (const newPwd of toAdd) {
      passwords.value.push(newPwd)
    }

    for (const newPwd of newPasswords) {
      const existingIndex = passwords.value.findIndex(p => p.id === newPwd.id)
      if (existingIndex !== -1) {
        Object.assign(passwords.value[existingIndex], newPwd)
      }
    }
  }

  function updatePasswordInPlace(id: string, updates: Partial<PasswordEntry>) {
    const index = passwords.value.findIndex(p => p.id === id)
    if (index !== -1) {
      Object.assign(passwords.value[index], updates)
      updateCachedValues()
    }
  }

  function removePassword(id: string) {
    const index = passwords.value.findIndex(p => p.id === id)
    if (index !== -1) {
      passwords.value.splice(index, 1)
    }
  }

  async function fetchPasswords(params: Record<string, unknown> = {}): Promise<PasswordResult> {
    isLoading.value = true
    error.value = null

    try {
      const response = passwordApi.getPasswords({
        search: searchQuery.value,
        tag: selectedTag.value,
        ...params
      })

      if (response.error) {
        error.value = response.error
        return { success: false, error: response.error }
      }

      const newPasswords = response.data.passwords || []
      reconcilePasswords(newPasswords)
      updateCachedValues()

      return { success: true, data: response.data }
    } catch (err) {
      error.value = err.message
      return { success: false, error: err.message }
    } finally {
      isLoading.value = false
    }
  }

  async function getPassword(id: string): Promise<PasswordResult> {
    try {
      const response = passwordApi.getPassword(id)

      if (response.error) {
        return { success: false, error: response.error }
      }

      return { success: true, data: response.data }
    } catch (err) {
      return { success: false, error: err.message }
    }
  }

  async function createPassword(data: Omit<PasswordEntry, 'id'>): Promise<PasswordResult> {
    isLoading.value = true
    error.value = null

    try {
      const response = passwordApi.createPassword(data)

      if (response.error) {
        error.value = response.error
        return { success: false, error: response.error }
      }

      passwords.value.unshift(response.data)
      updateCachedValues()
      await fetchTags()

      return { success: true, data: response.data }
    } catch (err) {
      error.value = err.message
      return { success: false, error: err.message }
    } finally {
      isLoading.value = false
    }
  }

  async function updatePassword(id: string, data: Partial<PasswordEntry>): Promise<PasswordResult> {
    isLoading.value = true
    error.value = null

    try {
      const response = passwordApi.updatePassword(id, data)

      if (response.error) {
        error.value = response.error
        return { success: false, error: response.error }
      }

      updatePasswordInPlace(id, response.data)
      await fetchTags()

      return { success: true, data: response.data }
    } catch (err) {
      error.value = err.message
      return { success: false, error: err.message }
    } finally {
      isLoading.value = false
    }
  }

  async function deletePassword(id: string): Promise<PasswordResult> {
    isLoading.value = true
    error.value = null

    try {
      const response = passwordApi.deletePassword(id)

      if (response.error) {
        error.value = response.error
        return { success: false, error: response.error }
      }

      removePassword(id)
      updateCachedValues()

      return { success: true }
    } catch (err) {
      error.value = err.message
      return { success: false, error: err.message }
    } finally {
      isLoading.value = false
    }
  }

  async function recordUse(id: string): Promise<PasswordResult> {
    try {
      const response = passwordApi.recordPasswordUse(id)

      if (response.error) {
        return { success: false, error: response.error }
      }

      updatePasswordInPlace(id, {
        useCount: ((passwords.value.find(p => p.id === id)?.useCount || 0) + 1),
        lastUsedAt: Date.now()
      })

      return { success: true }
    } catch (err) {
      return { success: false, error: err.message }
    }
  }

  async function fetchTags(): Promise<PasswordResult> {
    try {
      const response = passwordApi.getTags()

      if (response.error) {
        return { success: false, error: response.error }
      }

      const newTags = response.data.tags || []
      const existingTagsSet = new Set(tags.value)
      
      newTags.forEach(tag => existingTagsSet.add(tag))
      tags.value = Array.from(existingTagsSet).sort()
      
      return { success: true, data: response.data }
    } catch (err) {
      return { success: false, error: err.message }
    }
  }

  function setSearchQuery(query: string): void {
    searchQuery.value = query
    updateCachedValues()
  }

  function setSelectedTag(tag: string): void {
    selectedTag.value = tag
    updateCachedValues()
  }

  async function fetchSettings(): Promise<PasswordResult> {
    try {
      const response = passwordApi.getSettings()

      if (response.error) {
        return { success: false, error: response.error }
      }

      Object.assign(settings.value, response.data)
      return { success: true, data: settings.value }
    } catch (err) {
      return { success: false, error: err.message }
    }
  }

  async function saveSettings(newSettings: Partial<PasswordSettings>): Promise<PasswordResult> {
    try {
      const response = passwordApi.saveSettings(newSettings)

      if (response.error) {
        return { success: false, error: response.error }
      }

      Object.assign(settings.value, response.data)
      return { success: true, data: settings.value }
    } catch (err) {
      return { success: false, error: err.message }
    }
  }

  async function exportPasswords(): Promise<PasswordResult> {
    try {
      const response = passwordApi.exportPasswords()

      if (response.error) {
        return { success: false, error: response.error }
      }

      const dataStr = JSON.stringify(response.data, null, 2)
      const blob = new Blob([dataStr], { type: 'application/json' })
      const url = URL.createObjectURL(blob)

      const link = document.createElement('a')
      link.href = url
      link.download = `passwords_backup_${new Date().toISOString().split('T')[0]}.json`
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      URL.revokeObjectURL(url)

      return { success: true }
    } catch (err) {
      return { success: false, error: err.message }
    }
  }

  async function importPasswords(file: File, overwrite = false): Promise<PasswordResult> {
    isLoading.value = true
    error.value = null

    try {
      const text = await file.text()
      const data = JSON.parse(text)

      if (!data.passwords || !Array.isArray(data.passwords)) {
        throw new Error('无效的文件格式')
      }

      const response = passwordApi.importPasswords(data.passwords, overwrite)

      if (response.error) {
        error.value = response.error
        return { success: false, error: response.error }
      }

      await fetchPasswords()
      await fetchTags()

      return { success: true, count: data.passwords.length }
    } catch (err) {
      error.value = err.message
      return { success: false, error: err.message }
    } finally {
      isLoading.value = false
    }
  }

  async function clearAll(): Promise<PasswordResult> {
    isLoading.value = true
    error.value = null

    try {
      const response = passwordApi.clearAllPasswords()

      if (response.error) {
        error.value = response.error
        return { success: false, error: response.error }
      }

      passwords.value = []
      tags.value = []
      updateCachedValues()

      return { success: true }
    } catch (err) {
      error.value = err.message
      return { success: false, error: err.message }
    } finally {
      isLoading.value = false
    }
  }

  async function init(): Promise<void> {
    await Promise.all([
      fetchPasswords(),
      fetchTags(),
      fetchSettings()
    ])
  }

  return {
    passwords,
    tags,
    settings,
    isLoading,
    error,
    searchQuery,
    selectedTag,
    filteredPasswords,
    totalPasswords,
    recentPasswords,
    mostUsedPasswords,
    fetchPasswords,
    getPassword,
    createPassword,
    updatePassword,
    deletePassword,
    recordUse,
    fetchTags,
    setSearchQuery,
    setSelectedTag,
    fetchSettings,
    saveSettings,
    exportPasswords,
    importPasswords,
    clearAll,
    init,
    updatePasswordInPlace,
    removePassword
  }
})