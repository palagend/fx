<template>
  <div class="mobile-tools">
    <header class="tools-header">
      <h1>实用工具</h1>
      <p>便捷高效的工具集合</p>
    </header>

    <section class="tools-grid">
      <router-link to="/calculator" class="tool-card">
        <div class="tool-icon blue">
          <Icon icon="fa7-solid:calculator" />
        </div>
        <div class="tool-info">
          <h3>计算器</h3>
          <p>支持基础运算和科学计算</p>
        </div>
        <Icon icon="fa7-solid:chevron-right" class="tool-arrow" />
      </router-link>

      <router-link to="/exchange-rate" class="tool-card">
        <div class="tool-icon green">
          <Icon icon="fa7-solid:exchange-alt" />
        </div>
        <div class="tool-info">
          <h3>汇率换算</h3>
          <p>实时汇率查询与换算</p>
        </div>
        <Icon icon="fa7-solid:chevron-right" class="tool-arrow" />
      </router-link>

      <router-link to="/qrcode-generator" class="tool-card">
        <div class="tool-icon purple">
          <Icon icon="fa7-solid:qrcode" />
        </div>
        <div class="tool-info">
          <h3>二维码生成</h3>
          <p>快速生成各种二维码</p>
        </div>
        <Icon icon="fa7-solid:chevron-right" class="tool-arrow" />
      </router-link>

      <router-link to="/password-generator" class="tool-card">
        <div class="tool-icon orange">
          <Icon icon="fa7-solid:key" />
        </div>
        <div class="tool-info">
          <h3>密码生成</h3>
          <p>生成高强度安全密码</p>
        </div>
        <Icon icon="fa7-solid:chevron-right" class="tool-arrow" />
      </router-link>

      <router-link to="/password-manager" class="tool-card">
        <div class="tool-icon red">
          <Icon icon="fa7-solid:lock" />
        </div>
        <div class="tool-info">
          <h3>密码管理</h3>
          <p>安全管理您的密码</p>
        </div>
        <Icon icon="fa7-solid:chevron-right" class="tool-arrow" />
      </router-link>

      <router-link to="/portfolio" class="tool-card">
        <div class="tool-icon cyan">
          <Icon icon="fa7-solid:wallet" />
        </div>
        <div class="tool-info">
          <h3>资产组合</h3>
          <p>管理您的投资组合</p>
        </div>
        <Icon icon="fa7-solid:chevron-right" class="tool-arrow" />
      </router-link>
    </section>

    <section class="tips-section">
      <div class="tip-card">
        <Icon icon="fa7-solid:lightbulb" class="tip-icon" />
        <div class="tip-content">
          <h3>使用提示</h3>
          <p>所有工具均为本地运行，数据不会上传到服务器，请放心使用。</p>
        </div>
      </div>
    </section>

    <div class="bottom-space"></div>
  </div>
</template>

<script setup>
import { Icon } from '@iconify/vue'
</script>

<style scoped>
.mobile-tools {
  min-height: 100vh;
  background: #f0f2f5;
  padding-bottom: 80px;
}

.tools-header {
  background: linear-gradient(135deg, #4361ee 0%, #7209b7 100%);
  padding: 40px 20px 30px;
  border-radius: 0 0 30px 30px;
}

.tools-header h1 {
  font-size: 28px;
  font-weight: 700;
  color: white;
  margin: 0 0 8px;
  position: relative;
  z-index: 1;
}

.tools-header p {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.8);
  margin: 0;
  position: relative;
  z-index: 1;
}

.tools-grid {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  animation: slideInUp 0.4s ease-out forwards;
  opacity: 0;
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

.tool-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 18px 16px;
  background: white;
  border-radius: 16px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  text-decoration: none;
  transition: all 0.15s ease;
  position: relative;
  overflow: hidden;
}

.tool-card:active {
  background: rgba(67, 97, 238, 0.08);
  transform: scale(0.99);
}

.tool-card::after {
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

.tool-card:active::after {
  width: 300px;
  height: 300px;
}

.tool-icon {
  width: 44px;
  height: 44px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.tool-card:active .tool-icon {
  transform: scale(0.95);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.15);
}

.tool-icon svg {
  width: 22px;
  height: 22px;
  color: white;
}

.tool-icon.blue {
  background: linear-gradient(135deg, #4facfe, #00f2fe);
}

.tool-icon.green {
  background: linear-gradient(135deg, #11998e, #38ef7d);
}

.tool-icon.purple {
  background: linear-gradient(135deg, #667eea, #764ba2);
}

.tool-icon.orange {
  background: linear-gradient(135deg, #f093fb, #f5576c);
}

.tool-icon.red {
  background: linear-gradient(135deg, #fa709a, #fee140);
}

.tool-icon.cyan {
  background: linear-gradient(135deg, #43e97b, #38f9d7);
}

.tool-info {
  flex: 1;
  min-width: 0;
}

.tool-info h3 {
  font-size: 16px;
  font-weight: 500;
  color: #1a1a1a;
  margin: 0 0 4px;
  letter-spacing: 0.3px;
}

.tool-info p {
  font-size: 13px;
  color: #888;
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tool-arrow {
  width: 18px;
  height: 18px;
  color: #b0b0b0;
  flex-shrink: 0;
  transition: transform 0.15s ease;
}

.tool-card:active .tool-arrow {
  transform: translateX(2px);
}

.tips-section {
  padding: 0 20px;
  animation: slideInUp 0.4s ease-out 0.2s forwards;
  opacity: 0;
}

.tip-card {
  display: flex;
  gap: 16px;
  padding: 18px;
  background: linear-gradient(135deg, rgba(67, 97, 238, 0.08), rgba(114, 9, 183, 0.08));
  border-radius: 16px;
  border: 1px solid rgba(67, 97, 238, 0.12);
}

.tip-icon {
  width: 44px;
  height: 44px;
  background: linear-gradient(135deg, #4361ee, #7209b7);
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(67, 97, 238, 0.25);
}

.tip-icon svg {
  width: 22px;
  height: 22px;
  color: white;
}

.tip-content h3 {
  font-size: 15px;
  font-weight: 600;
  color: #333;
  margin: 0 0 6px;
}

.tip-content p {
  font-size: 13px;
  color: #666;
  margin: 0;
  line-height: 1.5;
}

.bottom-space {
  height: 20px;
}

.dark .mobile-tools {
  background: #0a0a0a;
}

.dark .tools-header {
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
}

.dark .tools-header h1 {
  color: #ffffff;
}

.dark .tools-header p {
  color: rgba(255, 255, 255, 0.65);
}

.dark .tool-card {
  background: rgba(25, 25, 25, 0.95);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.4);
}

.dark .tool-card:active {
  background: rgba(74, 144, 226, 0.12);
}

.dark .tool-card:active::after {
  background: rgba(74, 144, 226, 0.2);
}

.dark .tool-info h3 {
  color: #ffffff;
}

.dark .tool-info p {
  color: #888;
}

.dark .tool-arrow {
  color: #6c757d;
}

.dark .tip-card {
  background: rgba(74, 144, 226, 0.1);
  border-color: rgba(74, 144, 226, 0.15);
}

.dark .tip-icon {
  background: linear-gradient(135deg, #4a90e2, #6a5acd);
}

.dark .tip-content h3 {
  color: #ffffff;
}

.dark .tip-content p {
  color: #adb5bd;
}
</style>