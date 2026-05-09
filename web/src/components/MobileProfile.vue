<template>
  <div class="mobile-profile">
    <header class="profile-header" v-if="showProfileHeader">
      <div class="header-bg"></div>
      <div class="profile-info">
        <div class="avatar">
          <Icon icon="mdi:user" />
        </div>
        <div class="user-details">
          <h2>{{ profileName }}</h2>
          <p>{{ profileDesc }}</p>
        </div>
        <button v-if="showLoginBtn" class="login-btn" @click="userStore.openLoginModal">
          <Icon icon="mdi:login" />
          <span>登录</span>
        </button>
      </div>
    </header>

    <section class="menu-section">
      <div class="menu-group">
        <button class="menu-item" @click="goToPortfolio">
          <div class="menu-icon blue">
            <Icon icon="mdi:wallet" />
          </div>
          <span class="menu-text">资产组合</span>
          <Icon icon="mdi:chevron-right" class="menu-arrow" />
        </button>

        <button class="menu-item" @click="goToExchange">
          <div class="menu-icon green">
            <Icon icon="mdi:swap-horizontal" />
          </div>
          <span class="menu-text">汇率换算</span>
          <Icon icon="mdi:chevron-right" class="menu-arrow" />
        </button>

        <button class="menu-item" @click="goToCalculator">
          <div class="menu-icon orange">
            <Icon icon="mdi:calculator" />
          </div>
          <span class="menu-text">实用计算器</span>
          <Icon icon="mdi:chevron-right" class="menu-arrow" />
        </button>

        <button class="menu-item" @click="goToQRCode">
          <div class="menu-icon purple">
            <Icon icon="mdi:qrcode" />
          </div>
          <span class="menu-text">二维码生成</span>
          <Icon icon="mdi:chevron-right" class="menu-arrow" />
        </button>

        <button class="menu-item" @click="goToPasswordGenerator">
          <div class="menu-icon red">
            <Icon icon="mdi:key" />
          </div>
          <span class="menu-text">密码生成器</span>
          <Icon icon="mdi:chevron-right" class="menu-arrow" />
        </button>

        <button class="menu-item" @click="goToPasswordManager">
          <div class="menu-icon cyan">
            <Icon icon="mdi:lock" />
          </div>
          <span class="menu-text">密码管理器</span>
          <Icon icon="mdi:chevron-right" class="menu-arrow" />
        </button>
      </div>
    </section>

    <section class="settings-section" v-if="!config.isFrontend">
      <div class="menu-group">
        <button v-if="userStore.isLoggedIn" class="menu-item" @click="handleOpenProfileModal">
          <div class="menu-icon gray">
            <Icon icon="mdi:account-circle" />
          </div>
          <span class="menu-text">个人资料</span>
          <Icon icon="mdi:chevron-right" class="menu-arrow" />
        </button>

        <button v-if="userStore.isLoggedIn" class="menu-item" @click="handleOpenPasswordModal">
          <div class="menu-icon gray">
            <Icon icon="mdi:lock" />
          </div>
          <span class="menu-text">修改密码</span>
          <Icon icon="mdi:chevron-right" class="menu-arrow" />
        </button>

        <button v-if="userStore.isLoggedIn" class="menu-item logout" @click="handleLogout">
          <div class="menu-icon logout-icon">
            <Icon icon="mdi:logout" />
          </div>
          <span class="menu-text">退出登录</span>
          <Icon icon="mdi:chevron-right" class="menu-arrow" />
        </button>

        <button v-if="!userStore.isLoggedIn" class="menu-item" @click="userStore.openLoginModal">
          <div class="menu-icon gray">
            <Icon icon="mdi:login" />
          </div>
          <span class="menu-text">登录</span>
          <Icon icon="mdi:chevron-right" class="menu-arrow" />
        </button>
      </div>
    </section>

    <section class="theme-section">
      <div class="menu-group">
        <button class="menu-item" @click="toggleTheme">
          <div class="menu-icon dark-icon">
            <Icon icon="mdi:theme-light-dark" />
          </div>
          <span class="menu-text">{{ isDarkRef ? '暗色模式' : '浅色模式' }}</span>
          <div class="theme-switch" :class="{ active: isDarkRef }">
            <div class="switch-circle"></div>
          </div>
        </button>
      </div>
    </section>

    <section class="cache-section">
      <div class="menu-group">
        <button class="menu-item cache" @click="showClearCacheConfirm">
          <div class="menu-icon cache-icon">
            <Icon icon="mdi:trash-can" />
          </div>
          <span class="menu-text">清理缓存</span>
          <Icon icon="mdi:chevron-right" class="menu-arrow" />
        </button>
      </div>
    </section>

    <section class="about-section">
      <div class="menu-group">
        <button class="menu-item" @click="showAbout">
          <div class="menu-icon gray">
            <Icon icon="mdi:information-circle" />
          </div>
          <span class="menu-text">关于我们</span>
          <Icon icon="mdi:chevron-right" class="menu-arrow" />
        </button>
      </div>
    </section>

    <div class="bottom-space"></div>

    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showCacheConfirm" class="modal-overlay" @click.self="showCacheConfirm = false">
          <div class="modal-container">
            <div class="modal-header">
              <h3>清理缓存</h3>
              <button class="btn-close" @click="showCacheConfirm = false">
                <Icon icon="mdi:close" />
              </button>
            </div>
            <div class="modal-body">
              <div class="confirm-content">
                <div class="confirm-icon">
                  <Icon icon="mdi:alert-triangle" />
                </div>
                <p class="confirm-text">确定要清理所有缓存数据吗？</p>
                <p class="confirm-hint">清理后将删除本地存储的所有数据，包括资产组合、汇率记录等。</p>
              </div>
              <div class="confirm-actions">
                <button class="btn-cancel" @click="showCacheConfirm = false">取消</button>
                <button class="btn-confirm" @click="clearCache">确定清理</button>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showAboutModal" class="modal-overlay" @click.self="showAboutModal = false">
          <div class="modal-container">
            <div class="modal-header">
              <h3>关于工具集合</h3>
              <button class="btn-close" @click="showAboutModal = false">
                <Icon icon="mdi:close" />
              </button>
            </div>
            <div class="modal-body">
              <div class="about-content">
                <div class="logo-icon">
                  <Icon icon="mdi:wrench" />
                </div>
                <h2>工具集合</h2>
                <p>高效、便捷的实用工具平台</p>
                <div class="version-info">
                  <span>版本 {{ packageJson.version }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, computed, inject, watch } from 'vue'
import { Icon } from '@iconify/vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { config } from '../config'
import packageJson from '../../package.json'

const router = useRouter()
const userStore = useUserStore()

const showAboutModal = ref(false)
const showCacheConfirm = ref(false)

const isDarkRef = ref(false)
const isDark = inject('isDark', ref(false))
const toggleTheme = inject('toggleTheme', () => {})
const openProfileModal = inject('openProfileModal', () => {})
const openPasswordModal = inject('openPasswordModal', () => {})

if (isDark) {
  isDarkRef.value = isDark.value
  watch(isDark, (newVal) => {
    isDarkRef.value = newVal
  })
}

const showProfileHeader = computed(() => {
  return config.isFrontend || config.isBackend
})

const showLoginBtn = computed(() => {
  return config.isBackend && !userStore.isLoggedIn
})

const profileName = computed(() => {
  if (config.isFrontend) {
    return '本地用户'
  }
  return userStore.user?.username || '未登录'
})

const profileDesc = computed(() => {
  if (config.isFrontend) {
    return '数据本地存储'
  }
  return userStore.user?.email || '请登录以使用完整功能'
})

const goToPortfolio = () => {
  router.push('/portfolio')
}

const goToExchange = () => {
  router.push('/exchange-rate')
}

const goToCalculator = () => {
  router.push('/calculator')
}

const goToQRCode = () => {
  router.push('/qrcode-generator')
}

const goToPasswordGenerator = () => {
  router.push('/password-generator')
}

const goToPasswordManager = () => {
  router.push('/password-manager')
}

const handleOpenProfileModal = () => {
  openProfileModal()
}

const handleOpenPasswordModal = () => {
  openPasswordModal()
}

const handleLogout = async () => {
  await userStore.logout()
}

const showAbout = () => {
  showAboutModal.value = true
}

const showClearCacheConfirm = () => {
  showCacheConfirm.value = true
}

const clearCache = () => {
  localStorage.clear()
  showCacheConfirm.value = false
  router.push('/')
}
</script>

<style scoped>
.mobile-profile {
  min-height: 100vh;
  background: #f0f2f5;
}

.profile-header {
  position: relative;
  padding: 40px 20px 20px;
}

.header-bg {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 120px;
  background: linear-gradient(135deg, #4361ee 0%, #7209b7 100%);
  border-radius: 0;
}

.profile-info {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  gap: 16px;
}

.avatar {
  width: 56px;
  height: 56px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid white;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.avatar svg {
  width: 24px;
  height: 24px;
  color: white;
}

.user-details {
  flex: 1;
}

.user-details h2 {
  font-size: 17px;
  font-weight: 600;
  color: white;
  margin: 0 0 3px;
}

.user-details p {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.8);
  margin: 0;
}

.login-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 20px;
  color: white;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.login-btn svg {
  width: 16px;
  height: 16px;
}

.login-btn:active {
  transform: scale(0.96);
}

.menu-section,
.settings-section,
.cache-section,
.about-section {
  padding: 0 16px 16px;
  animation: slideInUp 0.4s ease-out forwards;
  opacity: 0;
}

.menu-section {
  animation-delay: 0.1s;
}

.settings-section {
  animation-delay: 0.2s;
}

.about-section {
  animation-delay: 0.3s;
}

@keyframes slideInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.menu-group {
  background: white;
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 16px;
  width: 100%;
  padding: 18px 16px;
  background: transparent;
  border: none;
  cursor: pointer;
  transition: all 0.15s ease;
  position: relative;
  overflow: hidden;
}

.menu-item::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent 0%, #e8e8e8 50%, transparent 100%);
}

.menu-item:first-child::before {
  display: none;
}

.menu-item:active {
  background: rgba(67, 97, 238, 0.08);
  transform: scale(0.99);
}

.menu-item::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  width: 0;
  height: 0;
  background: rgba(67, 97, 238, 0.1);
  border-radius: 50%;
  transform: translate(-50%, -50%);
  transition: width 0.3s ease, height 0.3s ease;
}

.menu-item:active::after {
  width: 300px;
  height: 300px;
}

.menu-icon {
  width: 44px;
  height: 44px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.menu-icon svg {
  width: 22px;
  height: 22px;
  color: white;
}

.menu-item:active .menu-icon {
  transform: scale(0.95);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.15);
}

.menu-icon.blue {
  background: linear-gradient(135deg, #4facfe, #00f2fe);
}

.menu-icon.green {
  background: linear-gradient(135deg, #11998e, #38ef7d);
}

.menu-icon.orange {
  background: linear-gradient(135deg, #f093fb, #f5576c);
}

.menu-icon.purple {
  background: linear-gradient(135deg, #667eea, #764ba2);
}

.menu-icon.red {
  background: linear-gradient(135deg, #fa709a, #fee140);
}

.menu-icon.cyan {
  background: linear-gradient(135deg, #43e97b, #38f9d7);
}

.menu-icon.gray {
  background: #f0f2f5;
}

.menu-icon.gray svg {
  color: #666;
}

.menu-icon.logout-icon {
  background: rgba(220, 53, 69, 0.1);
}

.menu-icon.logout-icon svg {
  color: #dc3545;
}

.menu-icon.cache-icon {
  background: rgba(245, 158, 11, 0.1);
}

.menu-icon.cache-icon svg {
  color: #f59e0b;
}

.menu-text {
  flex: 1;
  text-align: left;
  font-size: 16px;
  font-weight: 500;
  color: #1a1a1a;
  letter-spacing: 0.3px;
}

.menu-arrow {
  width: 18px;
  height: 18px;
  color: #b0b0b0;
  transition: transform 0.15s ease;
}

.menu-item:active .menu-arrow {
  transform: translateX(2px);
}

.menu-item.logout .menu-text {
  color: #dc3545;
}

.menu-icon.dark-icon {
  background: linear-gradient(135deg, #ffa726, #ff7043);
}

.menu-icon.dark-icon svg {
  color: white;
}

.theme-section {
  padding: 0 16px 16px;
  animation: slideInUp 0.4s ease-out 0.35s forwards;
  opacity: 0;
}

.theme-switch {
  width: 48px;
  height: 26px;
  background: #e8e8e8;
  border-radius: 13px;
  position: relative;
  transition: background 0.3s ease;
}

.theme-switch.active {
  background: linear-gradient(135deg, #4361ee, #7209b7);
}

.switch-circle {
  position: absolute;
  top: 3px;
  left: 3px;
  width: 20px;
  height: 20px;
  background: white;
  border-radius: 50%;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.15);
  transition: transform 0.3s ease;
}

.theme-switch.active .switch-circle {
  transform: translateX(22px);
}

.bottom-space {
  height: 16px;
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
  padding: 20px;
}

.modal-container {
  background: white;
  border-radius: 20px;
  width: 100%;
  max-width: 320px;
  overflow: hidden;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid #f0f0f0;
}

.modal-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: #333;
  margin: 0;
}

.btn-close {
  background: none;
  border: none;
  color: #999;
  font-size: 20px;
  cursor: pointer;
  padding: 4px;
  border-radius: 8px;
  transition: all 0.2s;
}

.btn-close:hover {
  background: #f0f0f0;
}

.modal-body {
  padding: 30px 20px;
}

.about-content {
  text-align: center;
}

.logo-icon {
  width: 64px;
  height: 64px;
  background: linear-gradient(135deg, #4361ee, #7209b7);
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 16px;
}

.logo-icon svg {
  width: 32px;
  height: 32px;
  color: white;
}

.about-content h2 {
  font-size: 20px;
  font-weight: 600;
  color: #333;
  margin: 0 0 8px;
}

.about-content p {
  font-size: 14px;
  color: #666;
  margin: 0 0 16px;
}

.version-info {
  font-size: 12px;
  color: #999;
}

.dark .mobile-profile {
  background: #0a0a0a;
}

.dark .header-bg {
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
}

.dark .avatar {
  background: rgba(74, 144, 226, 0.25);
  border-color: rgba(255, 255, 255, 0.2);
}

.dark .user-details h2 {
  color: #ffffff;
}

.dark .user-details p {
  color: rgba(255, 255, 255, 0.65);
}

.dark .login-btn {
  background: rgba(74, 144, 226, 0.25);
  border-color: rgba(74, 144, 226, 0.4);
}

.dark .login-btn svg {
  color: #4a90e2;
}

.dark .menu-group {
  background: rgba(25, 25, 25, 0.95);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.4);
}

.dark .menu-item::before {
  background: linear-gradient(90deg, transparent 0%, rgba(255, 255, 255, 0.06) 50%, transparent 100%);
}

.dark .menu-item:active {
  background: rgba(74, 144, 226, 0.12);
}

.dark .menu-item:active::after {
  background: rgba(74, 144, 226, 0.2);
}

.dark .menu-text {
  color: #ffffff;
}

.dark .menu-arrow {
  color: #6c757d;
}

.dark .theme-switch {
  background: #3d3d3d;
}

.dark .theme-switch.active {
  background: linear-gradient(135deg, #4a90e2, #6a5acd);
}

.dark .switch-circle {
  background: #e9ecef;
}

.dark .modal-container {
  background: rgba(30, 30, 30, 0.98);
}

.dark .modal-header {
  border-bottom-color: rgba(255, 255, 255, 0.08);
}

.dark .modal-header h3 {
  color: #e9ecef;
}

.dark .btn-close {
  color: #6c757d;
}

.dark .btn-close:hover {
  background: rgba(255, 255, 255, 0.05);
}

.dark .about-content h2 {
  color: #e9ecef;
}

.dark .about-content p {
  color: #adb5bd;
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

.confirm-content {
  text-align: center;
  padding: 1rem 0;
}

.confirm-icon {
  width: 64px;
  height: 64px;
  margin: 0 auto 1rem;
  background: rgba(245, 158, 11, 0.1);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 2rem;
  color: #f59e0b;
}

.confirm-text {
  font-size: 1.1rem;
  font-weight: 600;
  color: #333;
  margin: 0 0 0.5rem;
}

.confirm-hint {
  font-size: 0.85rem;
  color: #666;
  margin: 0;
  line-height: 1.5;
}

.confirm-actions {
  display: flex;
  gap: 1rem;
  margin-top: 1.5rem;
}

.btn-cancel,
.btn-confirm {
  flex: 1;
  padding: 0.8rem;
  border-radius: 10px;
  font-size: 0.95rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-cancel {
  background: #f0f2f5;
  border: none;
  color: #666;
}

.btn-cancel:hover {
  background: #e9ecef;
}

.btn-confirm {
  background: linear-gradient(135deg, #dc3545, #c0392b);
  border: none;
  color: white;
}

.btn-confirm:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(220, 53, 69, 0.3);
}

.dark .confirm-text {
  color: #e9ecef;
}

.dark .confirm-hint {
  color: #adb5bd;
}

.dark .btn-cancel {
  background: rgba(255, 255, 255, 0.08);
  color: #adb5bd;
}

.dark .btn-cancel:hover {
  background: rgba(255, 255, 255, 0.1);
}
</style>