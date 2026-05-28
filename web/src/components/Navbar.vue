<template>
  <nav class="navbar" v-if="!isMobile">
    <div class="nav-container">
      <div class="nav-left">
        <UserProfile v-if="config.isBackend" />
        <router-link to="/" class="nav-logo">
          <Icon icon="mdi:wrench" />
          <span>工具集合</span>
        </router-link>
      </div>
      <ul class="nav-menu" :class="{ open: isMenuOpen }">
        <li v-for="item in navItems" :key="item.path" class="nav-item">
          <router-link :to="item.path" class="nav-link" @click="closeMenu">
            <Icon :icon="item.icon" />
            <span>{{ item.label }}</span>
          </router-link>
        </li>
      </ul>
      <div class="nav-toggle" :class="{ active: isMenuOpen }" @click="toggleMenu">
        <span class="hamburger"></span>
      </div>
      <div class="theme-toggle" @click="toggleTheme">
        <Icon icon="solar:sun-bold" />
        <Icon icon="solar:moon-bold" />
        <div class="toggle-circle"></div>
      </div>
    </div>
  </nav>
</template>

<script setup>
import { ref } from 'vue'
import { Icon } from '@iconify/vue'
import UserProfile from './UserProfile.vue'
import { config } from '../config'

const props = defineProps({
  isMobile: Boolean,
  isDark: Boolean
})

const emit = defineEmits(['toggle-theme'])

const isMenuOpen = ref(false)

const navItems = [
  { path: '/exchange-rate', icon: 'mdi:swap-horizontal', label: '汇率查询' },
  { path: '/calculator', icon: 'mdi:calculator', label: '计算器' },
  { path: '/portfolio', icon: 'mdi:wallet', label: '资产组合' },
  { path: '/qrcode-generator', icon: 'mdi:qrcode', label: '二维码生成器' },
  { path: '/password-generator', icon: 'mdi:key', label: '密码生成器' },
  { path: '/password-manager', icon: 'mdi:lock', label: '密码管理器' }
]

const toggleMenu = () => {
  isMenuOpen.value = !isMenuOpen.value
}

const closeMenu = () => {
  isMenuOpen.value = false
}

const toggleTheme = () => {
  emit('toggle-theme')
}
</script>

<style scoped>
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
  font-size: 2rem;
  font-weight: 700;
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.nav-logo svg {
  width: 1.8rem;
  height: 1.8rem;
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
  }

  .nav-link {
    padding: 0.5rem 0.8rem;
    font-size: 0.9rem;
  }

  .nav-link span {
    display: none;
  }
}

@media (max-width: 768px) {
  .navbar {
    display: none;
  }
}
</style>
