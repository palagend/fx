<template>
  <div id="app" :class="{ dark: isDark, mobile: isMobile }">
    <nav class="navbar" v-if="!isMobile">
      <div class="nav-container">
        <div class="nav-left">
          <UserProfile v-if="config.isBackend" />
          <router-link to="/" class="nav-logo">
            <Icon icon="mdi:wrench" />
            <span>工具集合</span>
          </router-link>
        </div>
        <ul class="nav-menu">
          <li class="nav-item">
            <router-link to="/exchange-rate" class="nav-link" @click="closeMenu">
              <Icon icon="mdi:swap-horizontal" />
              <span>汇率查询</span>
            </router-link>
          </li>
          <li class="nav-item">
            <router-link to="/calculator" class="nav-link" @click="closeMenu">
              <Icon icon="mdi:calculator" />
              <span>计算器</span>
            </router-link>
          </li>
          <li class="nav-item">
            <router-link to="/portfolio" class="nav-link" @click="closeMenu">
              <Icon icon="mdi:wallet" />
              <span>资产组合</span>
            </router-link>
          </li>
          <li class="nav-item">
            <router-link to="/qrcode-generator" class="nav-link" @click="closeMenu">
              <Icon icon="mdi:qrcode" />
              <span>二维码生成器</span>
            </router-link>
          </li>
          <li class="nav-item">
            <router-link to="/password-generator" class="nav-link" @click="closeMenu">
              <Icon icon="mdi:key" />
              <span>密码生成器</span>
            </router-link>
          </li>
          <li class="nav-item">
            <router-link to="/password-manager" class="nav-link" @click="closeMenu">
              <Icon icon="mdi:lock" />
              <span>密码管理器</span>
            </router-link>
          </li>
        </ul>
        <div class="nav-toggle" @click="toggleMenu">
          <span class="hamburger"></span>
        </div>
        <div class="theme-toggle" @click="toggleTheme">
          <Icon icon="solar:sun-bold" />
          <Icon icon="solar:moon-bold" />
          <div class="toggle-circle"></div>
        </div>
      </div>
    </nav>

    <div class="container" :class="{ 'mobile-container': isMobile }">
      <router-view />
    </div>

    <MobileNav v-if="isMobile" />

    <Teleport to="body">
      <div v-if="isMobile" class="mobile-overlay" :class="{ show: showMobileOverlay }" @click="closeMobileOverlay"></div>
    </Teleport>

    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showLoginModal" class="modal-overlay" @click.self="closeLoginModal">
          <div class="modal-container">
            <div class="modal-header">
              <h3>{{ isRegistering ? '注册账号' : '用户登录' }}</h3>
              <button class="btn-close" @click="closeLoginModal">
                <Icon icon="mdi:close" />
              </button>
            </div>
            <div class="modal-body">
              <form @submit.prevent="handleAuthSubmit">
                <div class="form-group">
                  <label>
                    <Icon icon="mdi:user" />
                    用户名
                  </label>
                  <input
                    v-model="authForm.username"
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
                    v-model="authForm.email"
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
                      v-model="authForm.password"
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
                    v-model="authForm.confirmPassword"
                    :type="showPassword ? 'text' : 'password'"
                    placeholder="请再次输入密码"
                    required
                  />
                </div>
                <div v-if="authError" class="auth-error">
                  <Icon icon="mdi:alert-circle" />
                  <span>{{ authError }}</span>
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

    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showProfileModal" class="modal-overlay" @click.self="closeProfileModal">
          <div class="modal-container">
            <div class="modal-header">
              <h3>个人资料</h3>
              <button class="btn-close" @click="closeProfileModal">
                <Icon icon="mdi:close" />
              </button>
            </div>
            <div class="modal-body">
              <div class="profile-info">
                <div class="avatar-huge">
                  <Icon icon="mdi:user-circle" />
                </div>
                <div class="info-list">
                  <div class="info-item">
                    <span class="info-label">用户名</span>
                    <span class="info-value">{{ userStore.user?.username }}</span>
                  </div>
                  <div class="info-item">
                    <span class="info-label">邮箱</span>
                    <span class="info-value">{{ userStore.user?.email }}</span>
                  </div>
                  <div class="info-item">
                    <span class="info-label">注册时间</span>
                    <span class="info-value">{{ formatDate(userStore.user?.created_at) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showPasswordModal" class="modal-overlay" @click.self="closePasswordModal">
          <div class="modal-container">
            <div class="modal-header">
              <h3>修改密码</h3>
              <button class="btn-close" @click="closePasswordModal">
                <Icon icon="mdi:close" />
              </button>
            </div>
            <div class="modal-body">
              <form @submit.prevent="handlePasswordChange">
                <div class="form-group">
                  <label>
                    <Icon icon="mdi:lock" />
                    当前密码
                  </label>
                  <input
                    v-model="passwordForm.oldPassword"
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
                    v-model="passwordForm.newPassword"
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
                    v-model="passwordForm.confirmPassword"
                    type="password"
                    placeholder="请再次输入新密码"
                    required
                  />
                </div>
                <div v-if="passwordError" class="auth-error">
                  <Icon icon="mdi:alert-circle" />
                  <span>{{ passwordError }}</span>
                </div>
                <div v-if="passwordSuccess" class="auth-success">
                  <Icon icon="mdi:check-circle" />
                  <span>{{ passwordSuccess }}</span>
                </div>
                <button type="submit" class="btn-submit" :disabled="isChangingPassword">
                  <Icon v-if="isChangingPassword" icon="mdi:loading" class="spin" />
                  <span v-else>确认修改</span>
                </button>
              </form>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, provide, watch } from 'vue'
import { Icon } from '@iconify/vue'
import UserProfile from './components/UserProfile.vue'
import MobileNav from './components/MobileNav.vue'
import { useUserStore } from './stores/user'
import { config } from './config'

const userStore = useUserStore()

const isDark = ref(false)
const isMobile = ref(false)
const showMobileOverlay = ref(false)

const showLoginModal = ref(false)
const isRegistering = ref(false)
const showPassword = ref(false)
const authError = ref('')
const isSubmitting = ref(false)

const showProfileModal = ref(false)
const showPasswordModal = ref(false)
const passwordError = ref('')
const passwordSuccess = ref('')
const isChangingPassword = ref(false)

const passwordForm = ref({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const authForm = ref({
  username: '',
  email: '',
  password: '',
  confirmPassword: ''
})

watch(() => userStore.showLoginModal, (newVal) => {
  showLoginModal.value = newVal
})

const loadTheme = () => {
  const savedTheme = localStorage.getItem('theme')
  if (savedTheme) {
    isDark.value = savedTheme === 'dark'
    document.documentElement.classList.toggle('dark', isDark.value)
  }
}

const toggleTheme = () => {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

provide('isDark', isDark)
provide('toggleTheme', toggleTheme)

const isMenuOpen = ref(false)

const closeLoginModal = () => {
  userStore.closeLoginModal()
  authError.value = ''
  resetAuthForm()
}

const resetAuthForm = () => {
  authForm.value = {
    username: '',
    email: '',
    password: '',
    confirmPassword: ''
  }
  isRegistering.value = false
}

const handleAuthSubmit = async () => {
  if (isSubmitting.value) return
  isSubmitting.value = true
  authError.value = ''

  try {
    if (isRegistering.value) {
      if (authForm.value.password !== authForm.value.confirmPassword) {
        authError.value = '两次输入的密码不一致'
        return
      }
      const result = await userStore.register(authForm.value.username, authForm.value.email, authForm.value.password)
      if (result.success) {
        closeLoginModal()
      } else {
        authError.value = result.error
      }
    } else {
      const result = await userStore.login(authForm.value.username, authForm.value.password)
      if (result.success) {
        closeLoginModal()
      } else {
        authError.value = result.error
      }
    }
  } finally {
    isSubmitting.value = false
  }
}

const openProfileModal = () => {
  showProfileModal.value = true
}

const closeProfileModal = () => {
  showProfileModal.value = false
}

const openPasswordModal = () => {
  showPasswordModal.value = true
  passwordError.value = ''
  passwordSuccess.value = ''
}

const closePasswordModal = () => {
  showPasswordModal.value = false
  passwordError.value = ''
  passwordSuccess.value = ''
  passwordForm.value = {
    oldPassword: '',
    newPassword: '',
    confirmPassword: ''
  }
}

provide('openProfileModal', openProfileModal)
provide('openPasswordModal', openPasswordModal)

const handlePasswordChange = async () => {
  if (isChangingPassword.value) return
  isChangingPassword.value = true
  passwordError.value = ''
  passwordSuccess.value = ''

  try {
    if (passwordForm.value.newPassword !== passwordForm.value.confirmPassword) {
      passwordError.value = '两次输入的新密码不一致'
      return
    }

    const result = await userStore.changePassword(passwordForm.value.oldPassword, passwordForm.value.newPassword)
    if (result.success) {
      passwordSuccess.value = '密码修改成功'
      setTimeout(() => {
        closePasswordModal()
      }, 2000)
    } else {
      passwordError.value = result.error
    }
  } finally {
    isChangingPassword.value = false
  }
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  })
}

const toggleMenu = () => {
  isMenuOpen.value = !isMenuOpen.value
  const navMenu = document.querySelector('.nav-menu')
  const navToggle = document.querySelector('.nav-toggle')
  if (navMenu) {
    navMenu.classList.toggle('open', isMenuOpen.value)
  }
  if (navToggle) {
    navToggle.classList.toggle('active', isMenuOpen.value)
  }
}

const closeMenu = () => {
  if (isMenuOpen.value) {
    isMenuOpen.value = false
    const navMenu = document.querySelector('.nav-menu')
    const navToggle = document.querySelector('.nav-toggle')
    if (navMenu) {
      navMenu.classList.remove('open')
    }
    if (navToggle) {
      navToggle.classList.remove('active')
    }
  }
}

const checkMobile = () => {
  const width = window.innerWidth
  const wasMobile = isMobile.value
  isMobile.value = width < 768
  
  if (isMobile.value && !wasMobile) {
    closeMobileOverlay()
  }
}

const closeMobileOverlay = () => {
  showMobileOverlay.value = false
}

onMounted(() => {
  loadTheme()
  checkMobile()
  window.addEventListener('resize', checkMobile)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})
</script>

<style>
:root {
  --card-bg: #ffffff;
  --shadow: 0 4px 20px rgba(0, 0, 0, 0.12), 0 2px 8px rgba(0, 0, 0, 0.06);
  --border-color: #e8eaed;
  --text-primary: #1a1d21;
  --text-secondary: #5f6368;
  --text-muted: #9aa0a6;
  --primary-color: #4361ee;
  --primary-dark: #3651d4;
  --secondary-color: #7209b7;
  --input-bg: #f8f9fa;
  --btn-secondary-bg: #f1f3f4;
  --result-bg: #f8f9fa;
  --success-color: #28a745;
  --loss-bg: rgba(245, 87, 108, 0.08);
  --loss-border: rgba(245, 87, 108, 0.2);
  --bg-secondary: #f1f3f4;
  --bg-primary: #f8f9fa;
}

.dark {
  --card-bg: rgba(30, 30, 30, 0.98);
  --shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
  --border-color: rgba(255, 255, 255, 0.12);
  --text-primary: #e9ecef;
  --text-secondary: #adb5bd;
  --text-muted: #6c757d;
  --primary-color: #4a90e2;
  --primary-dark: #3d7bc9;
  --secondary-color: #7b68ee;
  --input-bg: #2d2d2d;
  --btn-secondary-bg: #3d3d3d;
  --result-bg: #2d2d2d;
  --success-color: #3ddc84;
  --loss-bg: rgba(245, 87, 108, 0.15);
  --loss-border: rgba(245, 87, 108, 0.3);
  --bg-secondary: #2d2d2d;
  --bg-primary: #1e1e1e;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  background: #f0f2f5;
  color: var(--text-primary);
  transition: background 0.3s ease, color 0.3s ease;
}

#app {
  min-height: 100vh;
}

.navbar {
  background: #4361ee;
  padding: 1rem 0;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  transition: background 0.3s ease;
}

.nav-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 2rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.nav-left {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.nav-logo {
  color: white;
  font-size: 1.5rem;
  font-weight: 700;
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.nav-logo svg {
  width: 1.3rem;
  height: 1.3rem;
}

.nav-logo span {
  color: white;
}

.nav-menu {
  list-style: none;
  display: flex;
  gap: 2rem;
}

.nav-link {
  color: white;
  text-decoration: none;
  font-size: 1rem;
  font-weight: 500;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border-radius: 8px;
}

.nav-link svg {
  width: 1rem;
  height: 1rem;
}

.nav-link:hover {
  background: rgba(255, 255, 255, 0.15);
  color: white;
}

.nav-link.router-link-active {
  background: rgba(255, 255, 255, 0.2);
  color: white;
}

.theme-toggle {
  background-color: rgba(255, 255, 255, 0.2);
  border: none;
  width: 50px;
  height: 28px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 4px;
  cursor: pointer;
  position: relative;
  transition: all 0.3s ease;
}

.theme-toggle svg {
  width: 12px;
  height: 12px;
  z-index: 1;
  color: white;
}

.theme-toggle .toggle-circle {
  position: absolute;
  left: 3px;
  width: 22px;
  height: 22px;
  background-color: white;
  border-radius: 50%;
  transition: transform 0.3s ease;
}

.dark .theme-toggle .toggle-circle {
  transform: translateX(22px);
}

.dark .theme-toggle i {
  color: white;
}

.container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 2rem;
}

.mobile-container {
  max-width: 480px;
  margin: 0 auto;
  padding: 0;
  padding-bottom: calc(80px + env(safe-area-inset-bottom));
  width: 100%;
  min-height: 100vh;
  box-sizing: border-box;
}

.dark body {
  background: #121212;
  color: #e9ecef;
}

.dark .navbar {
  background: #1e1e1e;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.dark .nav-link {
  color: rgba(255, 255, 255, 0.85);
}

.dark .nav-link:hover,
.dark .nav-link.router-link-active {
  color: white;
}

.nav-toggle {
  display: none;
  flex-direction: column;
  cursor: pointer;
  padding: 8px;
  border-radius: 8px;
  transition: background 0.3s ease;
}

.nav-toggle:hover {
  background: rgba(255, 255, 255, 0.15);
}

.nav-toggle.active .hamburger {
  background-color: transparent;
}

.nav-toggle.active .hamburger::before {
  top: 0;
  transform: rotate(45deg);
}

.nav-toggle.active .hamburger::after {
  top: 0;
  transform: rotate(-45deg);
}

.hamburger {
  width: 24px;
  height: 2px;
  background-color: white;
  position: relative;
  transition: all 0.3s ease;
}

.hamburger::before,
.hamburger::after {
  content: '';
  position: absolute;
  width: 24px;
  height: 2px;
  background-color: white;
  transition: all 0.3s ease;
}

.hamburger::before {
  top: -8px;
}

.hamburger::after {
  top: 8px;
}

.nav-menu.open {
  display: flex;
  flex-direction: column;
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: rgba(255, 255, 255, 0.98);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  padding: 1rem 0;
  z-index: 100;
  animation: slideDown 0.3s ease-out;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.dark .nav-menu.open {
  background: rgba(30, 30, 30, 0.98);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
}

.nav-menu.open .nav-item {
  width: 100%;
  text-align: center;
}

.nav-menu.open .nav-link {
  width: 100%;
  justify-content: center;
  padding: 0.75rem 1rem;
  transition: all 0.2s ease;
}

.nav-menu.open .nav-link:hover {
  background: rgba(67, 97, 238, 0.1);
  color: white;
}

.nav-menu.open .nav-link.router-link-active {
  background: rgba(67, 97, 238, 0.2);
  color: white;
}

@media (min-width: 993px) {
  .nav-menu {
    display: flex;
  }

  .nav-menu.open {
    display: flex;
    position: static;
    box-shadow: none;
    padding: 0;
  }

  .nav-menu.open .nav-item {
    width: auto;
  }

  .nav-menu.open .nav-link {
    width: auto;
  }

  .nav-toggle {
    display: none;
  }
}

@media (max-width: 992px) {
  .nav-container {
    flex-direction: row;
    flex-wrap: wrap;
    gap: 1rem;
    padding: 0.75rem 1rem;
  }

  .nav-left {
    flex: 1;
    gap: 0.75rem;
  }

  .nav-logo {
    font-size: 1.2rem;
  }

  .nav-logo span {
    display: none;
  }

  .nav-menu {
    display: none;
  }

  .nav-menu.open {
    display: flex;
    flex-direction: column;
    position: fixed;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: 90%;
    max-width: 400px;
    background: rgba(255, 255, 255, 0.98);
    box-shadow: 0 12px 48px rgba(0, 0, 0, 0.15);
    padding: 2rem 1rem;
    z-index: 1000;
    animation: slideDownCenter 0.3s ease-out;
    backdrop-filter: blur(10px);
  }

  @keyframes slideDownCenter {
    from {
      opacity: 0;
      transform: translate(-50%, -40%);
    }
    to {
      opacity: 1;
      transform: translate(-50%, -50%);
    }
  }

  .dark .nav-menu.open {
    background: rgba(30, 30, 30, 0.98);
    box-shadow: 0 12px 48px rgba(0, 0, 0, 0.5);
  }

  .nav-menu.open .nav-item {
    width: 100%;
    text-align: center;
  }

  .nav-menu.open .nav-link {
    width: 100%;
    justify-content: center;
    padding: 0.75rem 1rem;
    transition: all 0.2s ease;
    color: var(--text-primary);
  }

  .nav-menu.open .nav-link:hover {
    background: rgba(67, 97, 238, 0.1);
    color: white;
  }

  .nav-menu.open .nav-link.router-link-active {
    background: rgba(67, 97, 238, 0.2);
    color: white;
  }

  .nav-menu.open .nav-link svg {
    color: var(--text-primary);
  }

  .nav-menu.open .nav-link:hover svg,
  .nav-menu.open .nav-link.router-link-active svg {
    color: white;
  }

  .nav-toggle {
    display: flex;
    flex-direction: column;
    cursor: pointer;
    padding: 8px;
    border-radius: 8px;
    transition: background 0.3s ease;
  }

  .nav-toggle:hover {
    background: rgba(255, 255, 255, 0.15);
  }

  .nav-toggle.active .hamburger {
    background-color: transparent;
  }

  .nav-toggle.active .hamburger::before {
    top: 0;
    transform: rotate(45deg);
  }

  .nav-toggle.active .hamburger::after {
    top: 0;
    transform: rotate(-45deg);
  }

  .hamburger {
    width: 24px;
    height: 2px;
    background-color: white;
    position: relative;
    transition: all 0.3s ease;
  }

  .hamburger::before,
  .hamburger::after {
    content: '';
    position: absolute;
    width: 24px;
    height: 2px;
    background-color: white;
    transition: all 0.3s ease;
  }

  .hamburger::before {
    top: -8px;
  }

  .hamburger::after {
    top: 8px;
  }

  .nav-link {
    padding: 0.5rem 0.8rem;
    font-size: 0.9rem;
  }

  .nav-link span {
    display: none;
  }

  .container {
    padding: 1rem;
  }
}

@media (min-width: 993px) {
  .nav-toggle {
    display: none;
  }
}

@media (max-width: 768px) {
  .navbar {
    display: none;
  }

  .container {
    padding: 0;
    padding-bottom: calc(80px + env(safe-area-inset-bottom));
    max-width: 480px;
    margin: 0 auto;
    width: 100%;
    min-height: 100vh;
    box-sizing: border-box;
  }

  body {
    background: #f0f2f5;
  }

  .dark body {
    background: #121212;
  }
}

.mobile-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 998;
  opacity: 0;
  visibility: hidden;
  transition: all 0.3s ease;
}

.mobile-overlay.show {
  opacity: 1;
  visibility: visible;
}

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

.modal-enter-active,
.modal-leave-active {
  transition: all 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .modal-container,
.modal-leave-to .modal-container {
  transform: scale(0.95);
}

.spin {
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

.profile-info {
  text-align: center;
  padding: 1rem 0;
}

.avatar-huge {
  width: 100px;
  height: 100px;
  margin: 0 auto 1.5rem;
  background: linear-gradient(135deg, #4361ee 0%, #7209b7 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 3rem;
  color: white;
}

.info-list {
  text-align: left;
}

.info-item {
  display: flex;
  justify-content: space-between;
  padding: 0.8rem 0;
  border-bottom: 1px solid var(--border-color, rgba(0, 0, 0, 0.08));
}

.info-item:last-child {
  border-bottom: none;
}

.info-label {
  color: var(--text-secondary, #6c757d);
  font-size: 0.9rem;
}

.info-value {
  color: var(--text-primary, #212529);
  font-weight: 500;
  font-size: 0.9rem;
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
</style>