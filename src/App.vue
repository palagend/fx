<template>
  <div id="app" :class="{ dark: isDark, mobile: isMobile }">
    <Navbar :is-mobile="isMobile" :is-dark="isDark" @toggle-theme="toggleTheme" />

    <div class="container" :class="{ 'mobile-container': isMobile }">
      <router-view />
    </div>

    <MobileNav v-if="isMobile" />

    <Teleport to="body">
      <div v-if="isMobile" class="mobile-overlay" :class="{ show: showMobileOverlay }" @click="closeMobileOverlay"></div>
    </Teleport>

    <LoginModal :show="showLoginModal" @close="closeLoginModal" />
    <ProfileModal :show="showProfileModal" @close="closeProfileModal" />
    <PasswordModal :show="showPasswordModal" @close="closePasswordModal" />
  </div>
</template>

<script setup>
import { ref, watch, provide } from 'vue'
import Navbar from './components/Navbar.vue'
import MobileNav from './components/MobileNav.vue'
import LoginModal from './components/modals/LoginModal.vue'
import ProfileModal from './components/modals/ProfileModal.vue'
import PasswordModal from './components/modals/PasswordModal.vue'
import { useUserStore } from './stores/user'
import { useTheme } from './composables/useTheme'
import { useMobile } from './composables/useMobile'

const userStore = useUserStore()

const { isDark, toggleTheme, loadTheme, provideTheme } = useTheme()
const { isMobile, showMobileOverlay, closeMobileOverlay, provideMobile } = useMobile()

provideTheme()
provideMobile()

const showLoginModal = ref(false)
const showProfileModal = ref(false)
const showPasswordModal = ref(false)

watch(() => userStore.showLoginModal, (newVal) => {
  showLoginModal.value = newVal
})

const closeLoginModal = () => {
  userStore.closeLoginModal()
}

const closeProfileModal = () => {
  showProfileModal.value = false
}

const closePasswordModal = () => {
  showPasswordModal.value = false
}

const openProfileModal = () => {
  showProfileModal.value = true
}

const openPasswordModal = () => {
  showPasswordModal.value = true
}

provide('openProfileModal', openProfileModal)
provide('openPasswordModal', openPasswordModal)

loadTheme()
</script>

<style>
@import './styles/variables.css';

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
</style>
