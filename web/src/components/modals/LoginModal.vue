<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="show" class="modal-overlay" @click.self="close">
        <div class="modal-container">
          <div class="modal-header">
            <h3>{{ isRegistering ? '注册账号' : '用户登录' }}</h3>
            <button class="btn-close" @click="close">
              <Icon icon="mdi:close" />
            </button>
          </div>
          <div class="modal-body">
            <form @submit.prevent="handleSubmit">
              <div class="form-group">
                <label>
                  <Icon icon="mdi:user" />
                  用户名
                </label>
                <input
                  v-model="form.username"
                  type="text"
                  placeholder="请输入用户名"
                  required
                  minlength="3"
                  maxlength="50"
                />
              </div>
              <div v-if="isRegistering" class="form-group">
                <label>
                  <Icon icon="mdi:email" />
                  邮箱
                </label>
                <input
                  v-model="form.email"
                  type="email"
                  placeholder="请输入邮箱"
                  required
                />
              </div>
              <div class="form-group">
                <label>
                  <Icon icon="mdi:lock" />
                  密码
                </label>
                <div class="password-input">
                  <input
                    v-model="form.password"
                    :type="showPassword ? 'text' : 'password'"
                    placeholder="请输入密码"
                    required
                    minlength="6"
                  />
                  <button type="button" class="btn-toggle-password" @click="showPassword = !showPassword">
                    <Icon :icon="showPassword ? 'mdi:eye-off' : 'mdi:eye'" />
                  </button>
                </div>
              </div>
              <div v-if="isRegistering" class="form-group">
                <label>
                  <Icon icon="mdi:lock-check" />
                  确认密码
                </label>
                <input
                  v-model="form.confirmPassword"
                  :type="showPassword ? 'text' : 'password'"
                  placeholder="请再次输入密码"
                  required
                />
              </div>
              <div v-if="error" class="auth-error">
                <Icon icon="mdi:alert-circle" />
                <span>{{ error }}</span>
              </div>
              <button type="submit" class="btn-submit" :disabled="isSubmitting">
                <Icon v-if="isSubmitting" icon="mdi:loading" class="spin" />
                <span v-else>{{ isRegistering ? '注册' : '登录' }}</span>
              </button>
            </form>
            <div class="auth-switch">
              <span>{{ isRegistering ? '已有账号？' : '还没有账号？' }}</span>
              <button type="button" class="btn-switch" @click="isRegistering = !isRegistering">
                {{ isRegistering ? '立即登录' : '立即注册' }}
              </button>
            </div>
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

const isRegistering = ref(false)
const showPassword = ref(false)
const error = ref('')
const isSubmitting = ref(false)

const form = ref({
  username: '',
  email: '',
  password: '',
  confirmPassword: ''
})

watch(() => props.show, (newVal) => {
  if (!newVal) {
    resetForm()
  }
})

const resetForm = () => {
  form.value = {
    username: '',
    email: '',
    password: '',
    confirmPassword: ''
  }
  isRegistering.value = false
  error.value = ''
}

const close = () => {
  emit('close')
}

const handleSubmit = async () => {
  if (isSubmitting.value) return
  isSubmitting.value = true
  error.value = ''

  try {
    if (isRegistering.value) {
      if (form.value.password !== form.value.confirmPassword) {
        error.value = '两次输入的密码不一致'
        return
      }
      const result = await userStore.register(form.value.username, form.value.email, form.value.password)
      if (result.success) {
        close()
      } else {
        error.value = result.error
      }
    } else {
      const result = await userStore.login(form.value.username, form.value.password)
      if (result.success) {
        close()
      } else {
        error.value = result.error
      }
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

.password-input {
  position: relative;
}

.password-input input {
  padding-right: 2.5rem;
}

.btn-toggle-password {
  position: absolute;
  right: 0.8rem;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  color: var(--text-secondary, #6c757d);
  cursor: pointer;
  font-size: 1.2rem;
  display: flex;
  align-items: center;
  justify-content: center;
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

.auth-switch {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  margin-top: 1.2rem;
  color: var(--text-secondary, #6c757d);
  font-size: 0.9rem;
}

.btn-switch {
  background: none;
  border: none;
  color: #4361ee;
  font-weight: 600;
  cursor: pointer;
  padding: 0;
}

.btn-switch:hover {
  text-decoration: underline;
}
</style>
