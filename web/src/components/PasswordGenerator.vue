<template>
  <div class="password-generator-container">
    <div class="card">
      <h2 class="page-title">
        <Icon icon="fa7-solid:key" />
        <span>密码生成器</span>
      </h2>

      <div class="password-display">
        <input
          type="text"
          :value="password"
          readonly
          class="password-input"
          placeholder="点击生成按钮创建密码"
        />
        <button class="copy-btn" @click="copyPassword" :disabled="!password">
          <Icon :icon="copied ? 'fa7-solid:check' : 'fa7-solid:copy'" />
          <span>{{ copied ? '已复制' : '复制' }}</span>
        </button>
      </div>

      <div class="strength-indicator" v-if="password">
        <div class="strength-bar">
          <div
            class="strength-fill"
            :style="{ width: strengthPercentage + '%', backgroundColor: strengthColor }"
          ></div>
        </div>
        <span class="strength-text" :style="{ color: strengthColor }">
          {{ strengthText }}
        </span>
      </div>

      <!-- 保存密码表单 -->
      <div class="save-section" v-if="password">
        <div class="save-form">
          <input
            type="text"
            v-model="saveForm.title"
            placeholder="密码用途/标题（如：GitHub账号）"
            class="save-input"
          />
          <input
            type="text"
            v-model="saveForm.username"
            placeholder="用户名（可选）"
            class="save-input"
          />
          <input
            type="text"
            v-model="saveForm.url"
            placeholder="网站地址（可选）"
            class="save-input"
          />
          <div class="tags-input">
            <input
              type="text"
              v-model="tagInput"
              placeholder="添加标签，按回车确认"
              class="save-input"
              @keydown.enter.prevent="addTag"
            />
            <div class="tags-list" v-if="saveForm.tags.length > 0">
              <span v-for="(tag, index) in saveForm.tags" :key="index" class="tag">
                {{ tag }}
                <button @click="removeTag(index)" class="tag-remove">
                  <Icon icon="fa7-solid:times" />
                </button>
              </span>
            </div>
          </div>
          <button class="save-btn" @click="savePassword" :disabled="!saveForm.title || isSaving">
            <Icon :icon="isSaving ? 'fa7-solid:spinner' : 'fa7-solid:save'" :class="{ spinning: isSaving }" />
            <span>{{ isSaving ? '保存中...' : '保存到密码库' }}</span>
          </button>
        </div>
      </div>

      <div class="options-section">
        <h3>
          <Icon icon="fa7-solid:cog" />
          <span>选项设置</span>
        </h3>

        <div class="length-option">
          <label>密码长度: <strong>{{ length }}</strong></label>
          <input
            type="range"
            v-model.number="length"
            min="4"
            max="64"
            class="length-slider"
          />
          <div class="length-presets">
            <button
              v-for="preset in lengthPresets"
              :key="preset"
              class="preset-btn"
              :class="{ active: length === preset }"
              @click="length = preset"
            >
              {{ preset }}
            </button>
          </div>
        </div>

        <div class="checkboxes">
          <label class="checkbox-item">
            <input type="checkbox" v-model="options.uppercase" />
            <span class="checkmark"></span>
            <span class="label-text">
              <Icon icon="fa7-solid:font" />
              大写字母 (A-Z)
            </span>
          </label>

          <label class="checkbox-item">
            <input type="checkbox" v-model="options.lowercase" />
            <span class="checkmark"></span>
            <span class="label-text">
              <Icon icon="fa7-solid:font" />
              小写字母 (a-z)
            </span>
          </label>

          <label class="checkbox-item">
            <input type="checkbox" v-model="options.numbers" />
            <span class="checkmark"></span>
            <span class="label-text">
              <Icon icon="fa7-solid:hashtag" />
              数字 (0-9)
            </span>
          </label>

          <label class="checkbox-item">
            <input type="checkbox" v-model="options.symbols" />
            <span class="checkmark"></span>
            <span class="label-text">
              <Icon icon="fa7-solid:icons" />
              特殊符号 (!@#$%^&*)
            </span>
          </label>

          <label class="checkbox-item">
            <input type="checkbox" v-model="options.excludeSimilar" />
            <span class="checkmark"></span>
            <span class="label-text">
              <Icon icon="fa7-solid:ban" />
              排除相似字符 (0, O, l, 1)
            </span>
          </label>
        </div>
      </div>

      <button class="generate-btn" @click="generatePassword">
        <Icon icon="fa7-solid:sync" :class="{ spinning: isGenerating }" />
        <span>生成密码</span>
      </button>

      <!-- 最近保存的密码 -->
      <div class="saved-passwords-section" v-if="recentSavedPasswords.length > 0">
        <h3>
          <Icon icon="fa7-solid:lock" />
          <span>最近保存</span>
          <router-link to="/password-manager" class="view-all-link">
            查看全部 <Icon icon="fa7-solid:arrow-right" />
          </router-link>
        </h3>
        <div class="saved-list">
          <div
            v-for="item in recentSavedPasswords"
            :key="item.id"
            class="saved-item"
          >
            <div class="saved-info">
              <div class="saved-title">{{ item.title }}</div>
              <div class="saved-meta" v-if="item.username">
                <Icon icon="fa7-solid:user" /> {{ item.username }}
              </div>
            </div>
            <div class="saved-actions">
              <button class="action-btn" @click="copySavedPassword(item.password)" title="复制密码">
                <Icon icon="fa7-solid:copy" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <div class="history-section" v-if="history.length > 0">
        <h3>
          <Icon icon="fa7-solid:history" />
          <span>生成历史</span>
        </h3>
        <div class="history-list">
          <div
            v-for="(item, index) in history"
            :key="index"
            class="history-item"
          >
            <span class="history-password">{{ item }}</span>
            <button class="history-copy-btn" @click="copyToClipboard(item)">
              <Icon icon="fa7-solid:copy" />
            </button>
          </div>
        </div>
        <button class="clear-history-btn" @click="clearHistory">
          <Icon icon="fa7-solid:trash-alt" />
          <span>清空历史</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { usePasswordStore } from '../stores/password'

const passwordStore = usePasswordStore()

const password = ref('')
const length = ref(16)
const copied = ref(false)
const isGenerating = ref(false)
const isSaving = ref(false)
const history = ref([])
const tagInput = ref('')

const saveForm = reactive({
  title: '',
  username: '',
  url: '',
  tags: []
})

const lengthPresets = [8, 12, 16, 20, 32]

const options = reactive({
  uppercase: true,
  lowercase: true,
  numbers: true,
  symbols: true,
  excludeSimilar: false
})

const charSets = {
  uppercase: 'ABCDEFGHIJKLMNOPQRSTUVWXYZ',
  lowercase: 'abcdefghijklmnopqrstuvwxyz',
  numbers: '0123456789',
  symbols: '!@#$%^&*()_+-=[]{}|;:,.<>?'
}

const similarChars = /[0O1l]/g

const recentSavedPasswords = computed(() => passwordStore.recentPasswords)

const strength = computed(() => {
  if (!password.value) return 0

  let score = 0
  const pwd = password.value

  // 长度评分
  if (pwd.length >= 8) score += 10
  if (pwd.length >= 12) score += 10
  if (pwd.length >= 16) score += 10
  if (pwd.length >= 20) score += 10

  // 字符类型评分
  if (/[a-z]/.test(pwd)) score += 15
  if (/[A-Z]/.test(pwd)) score += 15
  if (/[0-9]/.test(pwd)) score += 15
  if (/[^a-zA-Z0-9]/.test(pwd)) score += 15

  return Math.min(score, 100)
})

const strengthPercentage = computed(() => strength.value)

const strengthText = computed(() => {
  const s = strength.value
  if (s < 40) return '弱'
  if (s < 60) return '一般'
  if (s < 80) return '强'
  return '非常强'
})

const strengthColor = computed(() => {
  const s = strength.value
  if (s < 40) return '#ef4444'
  if (s < 60) return '#f59e0b'
  if (s < 80) return '#3b82f6'
  return '#10b981'
})

const generatePassword = () => {
  isGenerating.value = true

  setTimeout(() => {
    let chars = ''
    let result = ''

    if (options.uppercase) chars += charSets.uppercase
    if (options.lowercase) chars += charSets.lowercase
    if (options.numbers) chars += charSets.numbers
    if (options.symbols) chars += charSets.symbols

    if (chars === '') {
      chars = charSets.lowercase
      options.lowercase = true
    }

    if (options.excludeSimilar) {
      chars = chars.replace(similarChars, '')
    }

    const array = new Uint32Array(length.value)
    crypto.getRandomValues(array)

    for (let i = 0; i < length.value; i++) {
      result += chars[array[i] % chars.length]
    }

    password.value = result

    // 添加到历史记录
    if (!history.value.includes(result)) {
      history.value.unshift(result)
      if (history.value.length > 10) {
        history.value.pop()
      }
    }

    // 重置保存表单
    saveForm.title = ''
    saveForm.username = ''
    saveForm.url = ''
    saveForm.tags = []

    isGenerating.value = false
  }, 200)
}

const addTag = () => {
  const tag = tagInput.value.trim()
  if (tag && !saveForm.tags.includes(tag)) {
    saveForm.tags.push(tag)
  }
  tagInput.value = ''
}

const removeTag = (index) => {
  saveForm.tags.splice(index, 1)
}

const savePassword = async () => {
  if (!password.value || !saveForm.title) return

  isSaving.value = true

  const result = await passwordStore.createPassword({
    password: password.value,
    title: saveForm.title,
    username: saveForm.username,
    url: saveForm.url,
    tags: saveForm.tags,
    strength: strength.value,
    length: password.value.length
  })

  if (result.success) {
    alert('密码已保存到密码库！')
    // 重置表单
    saveForm.title = ''
    saveForm.username = ''
    saveForm.url = ''
    saveForm.tags = []
  } else {
    alert('保存失败：' + result.error)
  }

  isSaving.value = false
}

const copyToClipboard = async (text) => {
  try {
    await navigator.clipboard.writeText(text)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 2000)
  } catch (err) {
    console.error('复制失败:', err)
  }
}

const copyPassword = () => {
  if (password.value) {
    copyToClipboard(password.value)
  }
}

const copySavedPassword = async (pwd) => {
  await copyToClipboard(pwd)
}

const clearHistory = () => {
  history.value = []
}

// 监听选项变化，确保至少选中一项
watch(
  () => [options.uppercase, options.lowercase, options.numbers, options.symbols],
  ([upper, lower, num, sym]) => {
    if (!upper && !lower && !num && !sym) {
      options.lowercase = true
    }
  }
)

onMounted(() => {
  passwordStore.fetchPasswords()
})
</script>

<style scoped>
.password-generator-container {
  max-width: 600px;
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

.password-display {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.password-input {
  flex: 1;
  padding: 16px 20px;
  font-size: 1.25rem;
  font-family: 'Courier New', monospace;
  border: 2px solid var(--border-color, #e5e7eb);
  border-radius: 12px;
  background: var(--input-bg, #f9fafb);
  color: var(--text-primary, #1f2937);
  outline: none;
  transition: all 0.3s ease;
}

.password-input:focus {
  border-color: #6366f1;
}

.copy-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 20px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
  border: none;
  border-radius: 12px;
  font-size: 1rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.copy-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(99, 102, 241, 0.3);
}

.copy-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.strength-indicator {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.strength-bar {
  flex: 1;
  height: 8px;
  background: var(--border-color, #e5e7eb);
  border-radius: 4px;
  overflow: hidden;
}

.strength-fill {
  height: 100%;
  border-radius: 4px;
  transition: all 0.3s ease;
}

.strength-text {
  font-size: 0.875rem;
  font-weight: 600;
  min-width: 60px;
  text-align: right;
}

/* 保存密码区域 */
.save-section {
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.05), rgba(139, 92, 246, 0.05));
  border: 1px dashed #6366f1;
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 24px;
}

.save-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.save-input {
  padding: 12px 16px;
  border: 2px solid var(--border-color, #e5e7eb);
  border-radius: 8px;
  background: var(--input-bg, #ffffff);
  color: var(--text-primary, #1f2937);
  font-size: 0.9375rem;
  outline: none;
  transition: all 0.3s ease;
}

.save-input:focus {
  border-color: #6366f1;
}

.tags-input {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.tags-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
  border-radius: 16px;
  font-size: 0.8125rem;
}

.tag-remove {
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
  transition: all 0.2s ease;
}

.tag-remove:hover {
  background: rgba(255, 255, 255, 0.4);
}

.save-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px 20px;
  background: linear-gradient(135deg, #10b981, #059669);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 0.9375rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.save-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(16, 185, 129, 0.3);
}

.save-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.options-section {
  margin-bottom: 24px;
}

.options-section h3 {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 1.125rem;
  color: var(--text-primary, #1f2937);
  margin-bottom: 16px;
}

.options-section h3 :deep(svg) {
  color: #6366f1;
}

.length-option {
  margin-bottom: 20px;
}

.length-option label {
  display: block;
  margin-bottom: 12px;
  color: var(--text-secondary, #6b7280);
}

.length-option label strong {
  color: #6366f1;
  font-size: 1.25rem;
}

.length-slider {
  width: 100%;
  height: 8px;
  border-radius: 4px;
  background: var(--border-color, #e5e7eb);
  outline: none;
  -webkit-appearance: none;
  margin-bottom: 12px;
}

.length-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  cursor: pointer;
  box-shadow: 0 2px 8px rgba(99, 102, 241, 0.3);
}

.length-presets {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.preset-btn {
  padding: 6px 14px;
  border: 2px solid var(--border-color, #e5e7eb);
  border-radius: 8px;
  background: transparent;
  color: var(--text-secondary, #6b7280);
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.3s ease;
}

.preset-btn:hover,
.preset-btn.active {
  border-color: #6366f1;
  color: #6366f1;
  background: rgba(99, 102, 241, 0.1);
}

.checkboxes {
  display: grid;
  gap: 12px;
}

.checkbox-item {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  padding: 8px;
  border-radius: 8px;
  transition: background 0.3s ease;
}

.checkbox-item:hover {
  background: var(--hover-bg, #f3f4f6);
}

.checkbox-item input {
  display: none;
}

.checkmark {
  width: 22px;
  height: 22px;
  border: 2px solid var(--border-color, #d1d5db);
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
  flex-shrink: 0;
}

.checkmark::after {
  content: '';
  width: 6px;
  height: 10px;
  border: solid white;
  border-width: 0 2px 2px 0;
  transform: rotate(45deg);
  opacity: 0;
  transition: opacity 0.3s ease;
}

.checkbox-item input:checked + .checkmark {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  border-color: #6366f1;
}

.checkbox-item input:checked + .checkmark::after {
  opacity: 1;
}

.label-text {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-primary, #1f2937);
  font-size: 0.9375rem;
}

.label-text :deep(svg) {
  color: #6366f1;
  font-size: 1rem;
}

.generate-btn {
  width: 100%;
  padding: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: white;
  border: none;
  border-radius: 12px;
  font-size: 1.125rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
}

.generate-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 12px 24px rgba(99, 102, 241, 0.3);
}

.generate-btn :deep(svg) {
  font-size: 1.25rem;
}

.generate-btn :deep(svg.spinning) {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* 最近保存的密码区域 */
.saved-passwords-section {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid var(--border-color, #e5e7eb);
}

.saved-passwords-section h3 {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 1.125rem;
  color: var(--text-primary, #1f2937);
  margin-bottom: 16px;
}

.saved-passwords-section h3 :deep(svg) {
  color: #6366f1;
}

.view-all-link {
  margin-left: auto;
  font-size: 0.875rem;
  color: #6366f1;
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 4px;
}

.view-all-link:hover {
  text-decoration: underline;
}

.saved-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.saved-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--input-bg, #f9fafb);
  border-radius: 8px;
  transition: all 0.3s ease;
}

.saved-item:hover {
  background: var(--hover-bg, #f3f4f6);
}

.saved-info {
  flex: 1;
  min-width: 0;
}

.saved-title {
  font-weight: 500;
  color: var(--text-primary, #1f2937);
  margin-bottom: 4px;
}

.saved-meta {
  font-size: 0.8125rem;
  color: var(--text-secondary, #6b7280);
  display: flex;
  align-items: center;
  gap: 4px;
}

.saved-actions {
  display: flex;
  gap: 8px;
}

.action-btn {
  padding: 8px;
  background: transparent;
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 6px;
  color: var(--text-secondary, #6b7280);
  cursor: pointer;
  transition: all 0.3s ease;
}

.action-btn:hover {
  background: #6366f1;
  border-color: #6366f1;
  color: white;
}

.history-section {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid var(--border-color, #e5e7eb);
}

.history-section h3 {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 1.125rem;
  color: var(--text-primary, #1f2937);
  margin-bottom: 16px;
}

.history-section h3 :deep(svg) {
  color: #6366f1;
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 12px;
}

.history-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  background: var(--input-bg, #f9fafb);
  border-radius: 8px;
  font-family: 'Courier New', monospace;
  font-size: 0.875rem;
}

.history-password {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-primary, #1f2937);
}

.history-copy-btn {
  padding: 6px 10px;
  background: transparent;
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 6px;
  color: var(--text-secondary, #6b7280);
  cursor: pointer;
  transition: all 0.3s ease;
}

.history-copy-btn:hover {
  background: #6366f1;
  border-color: #6366f1;
  color: white;
}

.clear-history-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 100%;
  padding: 10px;
  background: transparent;
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 8px;
  color: var(--text-secondary, #6b7280);
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.3s ease;
}

.clear-history-btn:hover {
  background: #ef4444;
  border-color: #ef4444;
  color: white;
}

/* Dark mode support */
:global(.dark) .card {
  --card-bg: #1f2937;
  --text-primary: #f9fafb;
  --text-secondary: #9ca3af;
  --border-color: #374151;
  --input-bg: #111827;
  --hover-bg: #374151;
}

@media (max-width: 640px) {
  .password-generator-container {
    padding: 12px;
  }

  .card {
    padding: 20px;
  }

  .password-display {
    flex-direction: column;
  }

  .copy-btn {
    width: 100%;
    justify-content: center;
  }

  .save-section {
    padding: 16px;
  }
}
</style>
