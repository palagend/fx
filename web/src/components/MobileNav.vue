<template>
  <nav class="mobile-nav">
    <div class="nav-items">
      <router-link 
        to="/portfolio" 
        class="nav-item"
        :class="{ active: $route.path.startsWith('/portfolio') }"
      >
        <div class="nav-icon">
          <Icon icon="mdi:wallet" />
        </div>
        <span class="nav-text">资产</span>
      </router-link>
      
      <router-link 
        to="/exchange-rate" 
        class="nav-item"
        :class="{ active: $route.path.startsWith('/exchange-rate') }"
      >
        <div class="nav-icon">
          <Icon icon="mdi:swap-horizontal" />
        </div>
        <span class="nav-text">汇率</span>
      </router-link>
      
      <router-link 
        to="/tools" 
        class="nav-item"
        :class="{ active: $route.path === '/tools' }"
      >
        <div class="nav-icon">
          <Icon icon="mdi:compass" />
        </div>
        <span class="nav-text">发现</span>
      </router-link>
      
      <router-link 
        to="/profile" 
        class="nav-item"
        :class="{ active: $route.path.startsWith('/profile') }"
      >
        <div class="nav-icon">
          <Icon icon="mdi:user" />
        </div>
        <span class="nav-text">我</span>
      </router-link>
    </div>
  </nav>
</template>

<script setup>
import { Icon } from '@iconify/vue'
</script>

<style scoped>
.mobile-nav {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  min-height: 56px;
  height: fit-content;
  background: rgba(255, 255, 255, 0.98);
  backdrop-filter: blur(20px);
  border-top: 1px solid rgba(0, 0, 0, 0.05);
  padding-bottom: env(safe-area-inset-bottom);
  z-index: 999;
  display: flex;
  align-items: center;
  justify-content: center;
}

.nav-items {
  display: flex;
  width: 100%;
  max-width: 480px;
  justify-content: space-around;
}

.nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1;
  min-height: 56px;
  padding: 8px 0;
  text-decoration: none;
  color: #999999;
  transition: all 0.2s ease;
  position: relative;
}

.nav-item.active {
  color: #4361ee;
}

.nav-icon {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 4px;
  transition: transform 0.2s ease;
}

.nav-icon svg {
  width: 24px;
  height: 24px;
}

.nav-item.active .nav-icon {
  transform: scale(1.1);
}

.nav-text {
  font-size: 10px;
  font-weight: 500;
}

.nav-item.active::after {
  content: '';
  position: absolute;
  bottom: 8px;
  width: 4px;
  height: 4px;
  background: #4361ee;
  border-radius: 50%;
}

.dark .mobile-nav {
  background: rgba(30, 30, 30, 0.98);
  border-top-color: rgba(255, 255, 255, 0.08);
}

.dark .nav-item {
  color: rgba(255, 255, 255, 0.6);
}

.dark .nav-item.active {
  color: #4a90e2;
}

.dark .nav-item.active::after {
  background: #4a90e2;
}
</style>