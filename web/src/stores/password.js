import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { passwordApi } from '../api'

export const usePasswordStore = defineStore('password', () => {
  // ========== 状态 ==========
  const passwords = ref([])
  const tags = ref([])
  const settings = ref({
    defaultLength: 16,
    autoSave: false,
    defaultTags: []
  })
  const isLoading = ref(false)
  const error = ref(null)
  const searchQuery = ref('')
  const selectedTag = ref('')

  // ========== 计算属性 ==========
  const filteredPasswords = computed(() => {
    let result = passwords.value

    if (searchQuery.value) {
      const query = searchQuery.value.toLowerCase()
      result = result.filter(p =>
        p.title?.toLowerCase().includes(query) ||
        p.username?.toLowerCase().includes(query) ||
        p.url?.toLowerCase().includes(query) ||
        p.tags?.some(t => t.toLowerCase().includes(query))
      )
    }

    if (selectedTag.value) {
      result = result.filter(p => p.tags?.includes(selectedTag.value))
    }

    return result
  })

  const totalPasswords = computed(() => passwords.value.length)

  const recentPasswords = computed(() => {
    return [...passwords.value]
      .sort((a, b) => (b.createdAt || 0) - (a.createdAt || 0))
      .slice(0, 5)
  })

  const mostUsedPasswords = computed(() => {
    return [...passwords.value]
      .sort((a, b) => (b.useCount || 0) - (a.useCount || 0))
      .slice(0, 5)
  })

  // ========== Actions ==========

  // 获取所有密码
  async function fetchPasswords(params = {}) {
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

      passwords.value = response.data.passwords || []
      return { success: true, data: response.data }
    } catch (err) {
      error.value = err.message
      return { success: false, error: err.message }
    } finally {
      isLoading.value = false
    }
  }

  // 获取单个密码
  async function getPassword(id) {
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

  // 创建密码
  async function createPassword(data) {
    isLoading.value = true
    error.value = null

    try {
      const response = passwordApi.createPassword(data)

      if (response.error) {
        error.value = response.error
        return { success: false, error: response.error }
      }

      passwords.value.unshift(response.data)
      await fetchTags() // 刷新标签列表

      return { success: true, data: response.data }
    } catch (err) {
      error.value = err.message
      return { success: false, error: err.message }
    } finally {
      isLoading.value = false
    }
  }

  // 更新密码
  async function updatePassword(id, data) {
    isLoading.value = true
    error.value = null

    try {
      const response = passwordApi.updatePassword(id, data)

      if (response.error) {
        error.value = response.error
        return { success: false, error: response.error }
      }

      const index = passwords.value.findIndex(p => p.id === id)
      if (index !== -1) {
        passwords.value[index] = response.data
      }

      await fetchTags() // 刷新标签列表

      return { success: true, data: response.data }
    } catch (err) {
      error.value = err.message
      return { success: false, error: err.message }
    } finally {
      isLoading.value = false
    }
  }

  // 删除密码
  async function deletePassword(id) {
    isLoading.value = true
    error.value = null

    try {
      const response = passwordApi.deletePassword(id)

      if (response.error) {
        error.value = response.error
        return { success: false, error: response.error }
      }

      passwords.value = passwords.value.filter(p => p.id !== id)

      return { success: true }
    } catch (err) {
      error.value = err.message
      return { success: false, error: err.message }
    } finally {
      isLoading.value = false
    }
  }

  // 记录密码使用
  async function recordUse(id) {
    try {
      const response = passwordApi.recordPasswordUse(id)

      if (response.error) {
        return { success: false, error: response.error }
      }

      const index = passwords.value.findIndex(p => p.id === id)
      if (index !== -1) {
        passwords.value[index].useCount = (passwords.value[index].useCount || 0) + 1
        passwords.value[index].lastUsedAt = Date.now()
      }

      return { success: true }
    } catch (err) {
      return { success: false, error: err.message }
    }
  }

  // 获取所有标签
  async function fetchTags() {
    try {
      const response = passwordApi.getTags()

      if (response.error) {
        return { success: false, error: response.error }
      }

      tags.value = response.data.tags || []
      return { success: true, data: response.data }
    } catch (err) {
      return { success: false, error: err.message }
    }
  }

  // 设置搜索关键词
  function setSearchQuery(query) {
    searchQuery.value = query
  }

  // 设置选中的标签
  function setSelectedTag(tag) {
    selectedTag.value = tag
  }

  // 获取设置
  async function fetchSettings() {
    try {
      const response = passwordApi.getSettings()

      if (response.error) {
        return { success: false, error: response.error }
      }

      settings.value = { ...settings.value, ...response.data }
      return { success: true, data: settings.value }
    } catch (err) {
      return { success: false, error: err.message }
    }
  }

  // 保存设置
  async function saveSettings(newSettings) {
    try {
      const response = passwordApi.saveSettings(newSettings)

      if (response.error) {
        return { success: false, error: response.error }
      }

      settings.value = { ...settings.value, ...response.data }
      return { success: true, data: settings.value }
    } catch (err) {
      return { success: false, error: err.message }
    }
  }

  // 导出密码
  async function exportPasswords() {
    try {
      const response = passwordApi.exportPasswords()

      if (response.error) {
        return { success: false, error: response.error }
      }

      // 创建下载文件
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

  // 导入密码
  async function importPasswords(file, overwrite = false) {
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

  // 清空所有密码
  async function clearAll() {
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

      return { success: true }
    } catch (err) {
      error.value = err.message
      return { success: false, error: err.message }
    } finally {
      isLoading.value = false
    }
  }

  // 初始化
  async function init() {
    await Promise.all([
      fetchPasswords(),
      fetchTags(),
      fetchSettings()
    ])
  }

  return {
    // 状态
    passwords,
    tags,
    settings,
    isLoading,
    error,
    searchQuery,
    selectedTag,

    // 计算属性
    filteredPasswords,
    totalPasswords,
    recentPasswords,
    mostUsedPasswords,

    // Actions
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
    init
  }
})
