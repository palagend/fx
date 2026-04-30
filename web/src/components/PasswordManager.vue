<template>
  <div class="password-manager-container">
    <div class="card">
      <h2 class="page-title">
        <Icon icon="fa7-solid:lock" />
        <span>密码管理器</span>
        <span class="password-count">({{ filteredPasswords.length }})</span>
      </h2>

      <!-- 搜索和筛选 -->
      <div class="search-section">
        <div class="search-box">
          <Icon icon="fa7-solid:search" class="search-icon" />
          <input
            type="text"
            v-model="searchQuery"
            placeholder="搜索密码..."
            class="search-input"
            @input="handleSearch"
          />
        </div>
        <div class="filter-tags" v-if="tags.length > 0">
          <button
            v-for="tag in tags"
            :key="tag"
            class="filter-tag"
            :class="{ active: selectedTag === tag }"
            @click="toggleTag(tag)"
          >
            {{ tag }}
          </button>
          <button
            v-if="selectedTag"
            class="clear-filter"
            @click="clearFilter"
          >
            <Icon icon="fa7-solid:times" />
            清除筛选
          </button>
        </div>
      </div>

      <!-- 操作按钮 -->
      <div class="actions-bar">
        <button class="action-btn-primary" @click="showImport = true">
          <Icon icon="fa7-solid:file-import" />
          <span>导入</span>
        </button>
        <button class="action-btn-primary" @click="exportPasswords">
          <Icon icon="fa7-solid:file-export" />
          <span>导出</span>
        </button>
        <button class="action-btn-danger" @click="confirmClearAll">
          <Icon icon="fa7-solid:trash-alt" />
          <span>清空</span>
        </button>
      </div>

      <!-- 密码列表 -->
      <div class="passwords-list" v-if="filteredPasswords.length > 0">
        <div
          v-for="item in filteredPasswords"
          :key="item.id"
          class="password-item"
        >
          <div class="password-header">
            <div class="password-title-section">
              <h3 class="password-title">{{ item.title }}</h3>
              <div class="password-tags" v-if="item.tags?.length">
                <span v-for="tag in item.tags" :key="tag" class="password-tag">{{ tag }}</span>
              </div>
            </div>
            <div class="password-strength" :style="{ backgroundColor: getStrengthColor(item.strength) }">
              {{ getStrengthText(item.strength) }}
            </div>
          </div>

          <div class="password-details">
            <div class="detail-row" v-if="item.username">
              <span class="detail-label">
                <Icon icon="fa7-solid:user" />
                用户名
              </span>
              <span class="detail-value">{{ item.username }}</span>
            </div>
            <div class="detail-row" v-if="item.url">
              <span class="detail-label">
                <Icon icon="fa7-solid:link" />
                网址
              </span>
              <a :href="item.url" target="_blank" class="detail-value link">{{ item.url }}</a>
            </div>
            <div class="detail-row">
              <span class="detail-label">
                <Icon icon="fa7-solid:key" />
                密码
              </span>
              <div class="password-value">
                <span class="password-mask">{{ showPassword[item.id] ? item.password : '••••••••' }}</span>
                <button class="toggle-btn" @click="toggleShowPassword(item.id)">
                  <Icon :icon="showPassword[item.id] ? 'fa7-solid:eye-slash' : 'fa7-solid:eye'" />
                </button>
              </div>
            </div>
          </div>

          <div class="password-meta">
            <span class="meta-item">
              <Icon icon="fa7-solid:calendar" />
              {{ formatDate(item.createdAt) }}
            </span>
            <span class="meta-item" v-if="item.useCount">
              <Icon icon="fa7-solid:copy" />
              使用 {{ item.useCount }} 次
            </span>
          </div>

          <div class="password-actions">
            <button class="action-btn" @click="copyPassword(item)" title="复制密码">
              <Icon icon="fa7-solid:copy" />
            </button>
            <button class="action-btn" @click="editPassword(item)" title="编辑">
              <Icon icon="fa7-solid:edit" />
            </button>
            <button class="action-btn delete" @click="confirmDelete(item)" title="删除">
              <Icon icon="fa7-solid:trash-alt" />
            </button>
          </div>
        </div>
      </div>

      <!-- 空状态 -->
      <div class="empty-state" v-else>
        <Icon icon="fa7-solid:lock-open" class="empty-icon" />
        <p>暂无保存的密码</p>
        <router-link to="/password-generator" class="create-link">
          去生成密码
        </router-link>
      </div>
    </div>

    <!-- 编辑对话框 -->
    <div class="modal-overlay" v-if="showEditModal" @click.self="closeEditModal">
      <div class="modal">
        <h3>
          <Icon icon="fa7-solid:edit" />
          编辑密码
        </h3>
        <div class="modal-form">
          <div class="form-group">
            <label>标题</label>
            <input type="text" v-model="editForm.title" />
          </div>
          <div class="form-group">
            <label>用户名</label>
            <input type="text" v-model="editForm.username" />
          </div>
          <div class="form-group">
            <label>网址</label>
            <input type="text" v-model="editForm.url" />
          </div>
          <div class="form-group">
            <label>密码</label>
            <div class="password-input-group">
              <input :type="showEditPassword ? 'text' : 'password'" v-model="editForm.password" />
              <button @click="showEditPassword = !showEditPassword">
                <Icon :icon="showEditPassword ? 'fa7-solid:eye-slash' : 'fa7-solid:eye'" />
              </button>
            </div>
          </div>
          <div class="form-group">
            <label>标签</label>
            <div class="tags-edit">
              <input
                type="text"
                v-model="editTagInput"
                placeholder="按回车添加标签"
                @keydown.enter.prevent="addEditTag"
              />
              <div class="tags-list" v-if="editForm.tags.length">
                <span v-for="(tag, index) in editForm.tags" :key="index" class="tag">
                  {{ tag }}
                  <button @click="removeEditTag(index)">
                    <Icon icon="fa7-solid:times" />
                  </button>
                </span>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-actions">
          <button class="btn-secondary" @click="closeEditModal">取消</button>
          <button class="btn-primary" @click="saveEdit" :disabled="!editForm.title">
            <Icon icon="fa7-solid:save" />
            保存
          </button>
        </div>
      </div>
    </div>

    <!-- 导入对话框 -->
    <div class="modal-overlay" v-if="showImport" @click.self="showImport = false">
      <div class="modal">
        <h3>
          <Icon icon="fa7-solid:file-import" />
          导入密码
        </h3>
        <div class="modal-body">
          <div class="file-upload">
            <input
              type="file"
              ref="fileInput"
              accept=".json"
              @change="handleFileSelect"
              class="file-input"
            />
            <div class="upload-area" @click="$refs.fileInput.click()">
              <Icon icon="fa7-solid:cloud-upload-alt" />
              <p>点击选择文件或拖拽到此处</p>
              <span>支持 .json 格式</span>
            </div>
          </div>
          <label class="checkbox-label">
            <input type="checkbox" v-model="importOverwrite" />
            <span>覆盖现有数据</span>
          </label>
        </div>
        <div class="modal-actions">
          <button class="btn-secondary" @click="showImport = false">取消</button>
          <button class="btn-primary" @click="confirmImport" :disabled="!selectedFile">
            导入
          </button>
        </div>
      </div>
    </div>

    <!-- 确认对话框 -->
    <div class="modal-overlay" v-if="showConfirm" @click.self="showConfirm = false">
      <div class="modal confirm-modal">
        <div class="confirm-icon" :class="confirmType">
          <Icon :icon="confirmType === 'danger' ? 'fa7-solid:exclamation-triangle' : 'fa7-solid:question-circle'" />
        </div>
        <h3>{{ confirmTitle }}</h3>
        <p>{{ confirmMessage }}</p>
        <div class="modal-actions">
          <button class="btn-secondary" @click="showConfirm = false">取消</button>
          <button :class="confirmType === 'danger' ? 'btn-danger' : 'btn-primary'" @click="confirmAction">
            确认
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { usePasswordStore } from '../stores/password'

const passwordStore = usePasswordStore()

// 状态
const searchQuery = ref('')
const selectedTag = ref('')
const showPassword = ref({})
const showEditModal = ref(false)
const showEditPassword = ref(false)
const showImport = ref(false)
const showConfirm = ref(false)
const selectedFile = ref(null)
const importOverwrite = ref(false)

// 确认对话框状态
const confirmTitle = ref('')
const confirmMessage = ref('')
const confirmType = ref('warning')
const confirmCallback = ref(null)

// 编辑表单
const editForm = ref({
  id: '',
  title: '',
  username: '',
  url: '',
  password: '',
  tags: []
})
const editTagInput = ref('')

// 计算属性
const filteredPasswords = computed(() => passwordStore.filteredPasswords)
const tags = computed(() => passwordStore.tags)

// 方法
const handleSearch = () => {
  passwordStore.setSearchQuery(searchQuery.value)
}

const toggleTag = (tag) => {
  selectedTag.value = selectedTag.value === tag ? '' : tag
  passwordStore.setSelectedTag(selectedTag.value)
}

const clearFilter = () => {
  selectedTag.value = ''
  searchQuery.value = ''
  passwordStore.setSelectedTag('')
  passwordStore.setSearchQuery('')
}

const toggleShowPassword = (id) => {
  showPassword.value[id] = !showPassword.value[id]
}

const getStrengthColor = (strength) => {
  if (!strength) return '#9ca3af'
  if (strength < 40) return '#ef4444'
  if (strength < 60) return '#f59e0b'
  if (strength < 80) return '#3b82f6'
  return '#10b981'
}

const getStrengthText = (strength) => {
  if (!strength) return '未知'
  if (strength < 40) return '弱'
  if (strength < 60) return '一般'
  if (strength < 80) return '强'
  return '非常强'
}

const formatDate = (timestamp) => {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  return date.toLocaleDateString('zh-CN')
}

const copyPassword = async (item) => {
  try {
    await navigator.clipboard.writeText(item.password)
    await passwordStore.recordUse(item.id)
    alert('密码已复制到剪贴板')
  } catch (err) {
    console.error('复制失败:', err)
  }
}

const editPassword = (item) => {
  editForm.value = {
    id: item.id,
    title: item.title,
    username: item.username || '',
    url: item.url || '',
    password: item.password,
    tags: [...(item.tags || [])]
  }
  showEditModal.value = true
}

const closeEditModal = () => {
  showEditModal.value = false
  showEditPassword.value = false
  editForm.value = { id: '', title: '', username: '', url: '', password: '', tags: [] }
}

const addEditTag = () => {
  const tag = editTagInput.value.trim()
  if (tag && !editForm.value.tags.includes(tag)) {
    editForm.value.tags.push(tag)
  }
  editTagInput.value = ''
}

const removeEditTag = (index) => {
  editForm.value.tags.splice(index, 1)
}

const saveEdit = async () => {
  const result = await passwordStore.updatePassword(editForm.value.id, {
    title: editForm.value.title,
    username: editForm.value.username,
    url: editForm.value.url,
    password: editForm.value.password,
    tags: editForm.value.tags
  })

  if (result.success) {
    closeEditModal()
  } else {
    alert('保存失败：' + result.error)
  }
}

const confirmDelete = (item) => {
  confirmTitle.value = '删除密码'
  confirmMessage.value = `确定要删除 "${item.title}" 吗？此操作不可恢复。`
  confirmType.value = 'danger'
  confirmCallback.value = () => deletePassword(item.id)
  showConfirm.value = true
}

const deletePassword = async (id) => {
  const result = await passwordStore.deletePassword(id)
  if (result.success) {
    showConfirm.value = false
  } else {
    alert('删除失败：' + result.error)
  }
}

const handleFileSelect = (event) => {
  selectedFile.value = event.target.files[0]
}

const confirmImport = async () => {
  if (!selectedFile.value) return

  const result = await passwordStore.importPasswords(selectedFile.value, importOverwrite.value)

  if (result.success) {
    alert(`成功导入 ${result.count} 条密码`)
    showImport.value = false
    selectedFile.value = null
  } else {
    alert('导入失败：' + result.error)
  }
}

const exportPasswords = () => {
  passwordStore.exportPasswords()
}

const confirmClearAll = () => {
  confirmTitle.value = '清空所有密码'
  confirmMessage.value = '确定要删除所有保存的密码吗？此操作不可恢复。'
  confirmType.value = 'danger'
  confirmCallback.value = clearAll
  showConfirm.value = true
}

const clearAll = async () => {
  const result = await passwordStore.clearAll()
  if (result.success) {
    showConfirm.value = false
  } else {
    alert('清空失败：' + result.error)
  }
}

const confirmAction = () => {
  if (confirmCallback.value) {
    confirmCallback.value()
  }
}

onMounted(() => {
  passwordStore.init()
})
</script>

<style scoped>
.password-manager-container {
  max-width: 900px;
  margin: 0 auto;
  padding: 20px;
}

.card {
  background: var(--card-bg, #ffffff);
  border-radius: 16px;
  padding: 30px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

.page-title {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 1.5rem;
  color: var(--text-primary, #1f2937);
  margin-bottom: 24px;
}

.page-title :deep(svg) {
  font-size: 1.75rem;
  color: #6366f1;
}

.password-count {
  font-size: 1rem;
  color: var(--text-secondary, #6b7280);
  font-weight: 400;
}

/* 搜索区域 */
.search-section {
  margin-bottom: 20px;
}

.search-box {
  position: relative;
  margin-bottom: 12px;
}

.search-icon {
  position: absolute;
  left: 16px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-secondary, #6b7280);
  font-size: 1.125rem;
}

.search-input {
  width: 100%;
  padding: 14px 16px 14px 48px;
  border: 2px solid var(--border-color, #e5e7eb);
  border-radius: 12px;
  background: var(--input-bg, #f9fafb);
  color: var(--text-primary, #1f2937);
  font-size: 1rem;
  outline: none;
  transition: all 0.3s ease;
}

.search-input:focus {
  border-color: #6366f1;
}

.filter-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.filter-tag {
  padding: 6px 14px;
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 20px;
  background: transparent;
  color: var(--text-secondary, #6b7280);
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.3s ease;
}

.filter-tag:hover,
.filter-tag.active {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  border-color: #6366f1;
  color: white;
}

.clear-filter {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 14px;
  border: 1px solid #ef4444;
  border-radius: 20px;
  background: transparent;
  color: #ef4444;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.3s ease;
}

.clear-filter:hover {
  background: #ef4444;
  color: white;
}

/* 操作栏 */
.actions-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 24px;
  flex-wrap: wrap;
}

.action-btn-primary,
.action-btn-danger {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 18px;
  border: none;
  border-radius: 8px;
  font-size: 0.9375rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.action-btn-primary {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
}

.action-btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(99, 102, 241, 0.3);
}

.action-btn-danger {
  background: transparent;
  border: 1px solid #ef4444;
  color: #ef4444;
}

.action-btn-danger:hover {
  background: #ef4444;
  color: white;
}

/* 密码列表 */
.passwords-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.password-item {
  background: var(--input-bg, #f9fafb);
  border-radius: 12px;
  padding: 20px;
  transition: all 0.3s ease;
}

.password-item:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.password-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;
}

.password-title-section {
  flex: 1;
}

.password-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
  margin-bottom: 8px;
}

.password-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.password-tag {
  padding: 3px 10px;
  background: rgba(99, 102, 241, 0.1);
  border-radius: 12px;
  font-size: 0.75rem;
  color: #6366f1;
}

.password-strength {
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 0.75rem;
  font-weight: 500;
  color: white;
}

.password-details {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 16px;
}

.detail-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.detail-label {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 80px;
  font-size: 0.875rem;
  color: var(--text-secondary, #6b7280);
}

.detail-value {
  flex: 1;
  font-size: 0.9375rem;
  color: var(--text-primary, #1f2937);
}

.detail-value.link {
  color: #6366f1;
  text-decoration: none;
}

.detail-value.link:hover {
  text-decoration: underline;
}

.password-value {
  display: flex;
  align-items: center;
  gap: 12px;
}

.password-mask {
  font-family: 'Courier New', monospace;
  font-size: 1rem;
  color: var(--text-primary, #1f2937);
}

.toggle-btn {
  padding: 6px;
  background: transparent;
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 6px;
  color: var(--text-secondary, #6b7280);
  cursor: pointer;
  transition: all 0.3s ease;
}

.toggle-btn:hover {
  background: #6366f1;
  border-color: #6366f1;
  color: white;
}

.password-meta {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--border-color, #e5e7eb);
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.8125rem;
  color: var(--text-secondary, #6b7280);
}

.password-actions {
  display: flex;
  gap: 8px;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  background: transparent;
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 8px;
  color: var(--text-secondary, #6b7280);
  cursor: pointer;
  transition: all 0.3s ease;
}

.action-btn:hover {
  background: #6366f1;
  border-color: #6366f1;
  color: white;
}

.action-btn.delete:hover {
  background: #ef4444;
  border-color: #ef4444;
}

/* 空状态 */
.empty-state {
  text-align: center;
  padding: 60px 20px;
}

.empty-icon {
  font-size: 4rem;
  color: var(--border-color, #e5e7eb);
  margin-bottom: 16px;
}

.empty-state p {
  color: var(--text-secondary, #6b7280);
  margin-bottom: 16px;
}

.create-link {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
  border-radius: 8px;
  text-decoration: none;
  font-weight: 500;
  transition: all 0.3s ease;
}

.create-link:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(99, 102, 241, 0.3);
}

/* 对话框 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}

.modal {
  background: var(--card-bg, #ffffff);
  border-radius: 16px;
  padding: 24px;
  width: 100%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal h3 {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 1.25rem;
  color: var(--text-primary, #1f2937);
  margin-bottom: 20px;
}

.modal h3 :deep(svg) {
  color: #6366f1;
}

.modal-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-bottom: 24px;
}

.form-group label {
  display: block;
  margin-bottom: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-secondary, #6b7280);
}

.form-group input {
  width: 100%;
  padding: 12px 16px;
  border: 2px solid var(--border-color, #e5e7eb);
  border-radius: 8px;
  background: var(--input-bg, #f9fafb);
  color: var(--text-primary, #1f2937);
  font-size: 0.9375rem;
  outline: none;
  transition: all 0.3s ease;
}

.form-group input:focus {
  border-color: #6366f1;
}

.password-input-group {
  display: flex;
  gap: 8px;
}

.password-input-group input {
  flex: 1;
}

.password-input-group button {
  padding: 12px;
  background: transparent;
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 8px;
  color: var(--text-secondary, #6b7280);
  cursor: pointer;
  transition: all 0.3s ease;
}

.password-input-group button:hover {
  background: #6366f1;
  border-color: #6366f1;
  color: white;
}

.tags-edit input {
  margin-bottom: 8px;
}

.tags-edit .tags-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tags-edit .tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
  border-radius: 16px;
  font-size: 0.8125rem;
}

.tags-edit .tag button {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  padding: 0;
  background: rgba(255, 255, 255, 0.2);
  border: none;
  border-radius: 50%;
  color: white;
  cursor: pointer;
}

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.btn-secondary,
.btn-primary,
.btn-danger {
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  font-size: 0.9375rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-secondary {
  background: var(--input-bg, #f3f4f6);
  color: var(--text-secondary, #6b7280);
}

.btn-secondary:hover {
  background: var(--border-color, #e5e7eb);
}

.btn-primary {
  display: flex;
  align-items: center;
  gap: 6px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(99, 102, 241, 0.3);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-danger {
  background: #ef4444;
  color: white;
}

.btn-danger:hover {
  background: #dc2626;
}

/* 确认对话框 */
.confirm-modal {
  text-align: center;
}

.confirm-icon {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 16px;
  font-size: 2rem;
}

.confirm-icon.warning {
  background: rgba(245, 158, 11, 0.1);
  color: #f59e0b;
}

.confirm-icon.danger {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.confirm-modal h3 {
  justify-content: center;
}

.confirm-modal p {
  color: var(--text-secondary, #6b7280);
  margin-bottom: 24px;
}

/* 文件上传 */
.file-upload {
  margin-bottom: 16px;
}

.file-input {
  display: none;
}

.upload-area {
  border: 2px dashed var(--border-color, #e5e7eb);
  border-radius: 12px;
  padding: 40px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s ease;
}

.upload-area:hover {
  border-color: #6366f1;
  background: rgba(99, 102, 241, 0.05);
}

.upload-area :deep(svg) {
  font-size: 3rem;
  color: #6366f1;
  margin-bottom: 12px;
}

.upload-area p {
  color: var(--text-primary, #1f2937);
  margin-bottom: 4px;
}

.upload-area span {
  font-size: 0.875rem;
  color: var(--text-secondary, #6b7280);
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.checkbox-label input {
  width: 18px;
  height: 18px;
  accent-color: #6366f1;
}

/* Dark mode support */
:global(.dark) .card,
:global(.dark) .modal {
  --card-bg: #1f2937;
  --text-primary: #f9fafb;
  --text-secondary: #9ca3af;
  --border-color: #374151;
  --input-bg: #111827;
}

@media (max-width: 640px) {
  .password-manager-container {
    padding: 12px;
  }

  .card {
    padding: 20px;
  }

  .password-header {
    flex-direction: column;
    gap: 12px;
  }

  .password-strength {
    align-self: flex-start;
  }

  .detail-row {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }

  .detail-label {
    width: auto;
  }

  .actions-bar {
    justify-content: center;
  }

  .modal-actions {
    flex-direction: column;
  }

  .btn-secondary,
  .btn-primary,
  .btn-danger {
    width: 100%;
    justify-content: center;
  }
}
</style>
