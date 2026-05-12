<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="show" class="modal-overlay" @click.self="close">
        <div class="modal-container">
          <div class="modal-header">
            <h3>修改密码</h3>
            <button class="btn-close" @click="close">
              <Icon icon="mdi:close" />
            </button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="handleSubmit">
              <div class="form-group">
                <label>
                  <Icon icon="mdi:lock" />
                  当前密码
                </label>
                <input
                  v-model="form.oldPassword"
                  type="password"
                  placeholder="请输入当前密码"
                  required
                />
              </div>
              <div class="form-group">
                <label>
                  <Icon icon="mdi:lock-plus" />
                  新密码
                </label>
                <input
                  v-model="form.newPassword"
                  type="password"
                  placeholder="请输入新密码（至少6位）"
                  required
                  minlength="6"
                />
              </div>
              <div class="form-group">
                <label>
                  <Icon icon="mdi:lock-check" />
                  确认新密码
                </label>
                <input
                  v-model="form.confirmPassword"
                  type="password"
                  placeholder="请再次输入新密码"
                  required
                />
              </div>
              <div v-if="error" class="auth-error">
                <Icon icon="mdi:alert-circle" />
                <span>{{ error }}</span>
              </div>
              <div v-if="success" class="auth-success">
                <Icon icon="mdi:check-circle" />
                <span>{{ success }}</span>
              </div>
              <button type="submit" class="btn-submit" :disabled="isSubmitting">
                <Icon v-if="isSubmitting" icon="mdi:loading" class="spin" />
                <span v-else>确认修改</span>
              </button>
            </form>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, watch } from 'vue'
import { Icon } from '@iconify/vue'
import { useUserStore } from '../../stores/user'

const props = defineProps({
  show: Boolean
})

const emit = defineEmits(['close'])

const userStore = useUserStore()

const form = ref({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const error = ref('')
const success = ref('')
const isSubmitting = ref(false)

watch(() => props.show, (newVal) => {
  if (!newVal) {
    resetForm()
  }
})

const resetForm = () => {
  form.value = {
    oldPassword: '',
    newPassword: '',
    confirmPassword: ''
  }
  error.value = ''
  success.value = ''
}

const close = () => {
  emit('close')
}

const handleSubmit = async () => {
  if (isSubmitting.value) return
  isSubmitting.value = true
  error.value = ''
  success.value = ''

  try {
    if (form.value.newPassword !== form.value.confirmPassword) {
      error.value = '两次输入的新密码不一致'
      return
    }

    const result = await userStore.changePassword(form.value.oldPassword, form.value.newPassword)
    if (result.success) {
      success.value = '密码修改成功'
      setTimeout(() => {
        close()
      }, 2000)
    } else {
      error.value = result.error
    }
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
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
  z-index: 2000;
  padding: 1rem;
}

.modal-container {
  background: var(--card-bg, white);
  border-radius: 16px;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  overflow: hidden;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.2rem 1.5rem;
  border-bottom: 1px solid var(--border-color, rgba(0, 0, 0, 0.08));
}

.modal-header h3 {
  margin: 0;
  font-size: 1.2rem;
  color: var(--text-primary, #212529);
}

.btn-close {
  background: none;
  border: none;
  color: var(--text-secondary, #6c757d);
  font-size: 1.5rem;
  cursor: pointer;
  padding: 0.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  transition: all 0.2s ease;
}

.btn-close:hover {
  background: var(--btn-secondary-bg, #f8f9fa);
  color: var(--text-primary, #212529);
}

.modal-body {
  padding: 1.5rem;
}

.form-group {
  margin-bottom: 1.2rem;
}

.form-group label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
  color: var(--text-secondary, #6c757d);
  font-size: 0.85rem;
  font-weight: 500;
}

.form-group input {
  width: 100%;
  padding: 0.8rem 1rem;
  border: 1px solid var(--border-color, rgba(0, 0, 0, 0.15));
  border-radius: 10px;
  background: var(--input-bg, white);
  color: var(--text-primary, #212529);
  font-size: 0.95rem;
  transition: all 0.3s ease;
}

.form-group input:focus {
  outline: none;
  border-color: #4361ee;
  box-shadow: 0 0 0 3px rgba(67, 97, 238, 0.1);
}

.auth-error {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.8rem 1rem;
  border-radius: 8px;
  margin-bottom: 1rem;
  font-size: 0.9rem;
  background: rgba(220, 53, 69, 0.1);
  color: #dc3545;
}

.auth-success {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.8rem 1rem;
  border-radius: 8px;
  margin-bottom: 1rem;
  font-size: 0.9rem;
  background: rgba(25, 135, 84, 0.1);
  color: #198754;
}

.btn-submit {
  width: 100%;
  padding: 0.9rem;
  background: linear-gradient(135deg, #4361ee 0%, #7209b7 100%);
  border: none;
  border-radius: 10px;
  color: white;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.btn-submit:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(67, 97, 238, 0.3);
}

.btn-submit:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}
</style>
