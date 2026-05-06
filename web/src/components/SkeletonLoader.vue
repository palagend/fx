<template>
  <div class="skeleton-container" aria-busy="true" aria-label="加载中">
    <!-- 概览卡片骨架 -->
    <div class="overview-skeleton">
      <div v-for="i in 5" :key="'overview-' + i" class="skeleton-card">
        <div class="skeleton-line short"></div>
        <div class="skeleton-line long"></div>
        <div class="skeleton-line medium"></div>
      </div>
    </div>

    <!-- 图表区域骨架 -->
    <div class="chart-skeleton">
      <div class="skeleton-header">
        <div class="skeleton-line medium"></div>
        <div class="skeleton-btn"></div>
      </div>
      <div class="skeleton-chart-wrapper">
        <div class="skeleton-chart"></div>
        <div class="skeleton-legend">
          <div v-for="i in 4" :key="'legend-' + i" class="skeleton-legend-item">
            <div class="skeleton-color"></div>
            <div class="skeleton-line short"></div>
            <div class="skeleton-line tiny"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- 交易区域骨架 -->
    <div class="trading-skeleton">
      <div class="skeleton-header">
        <div class="skeleton-line medium"></div>
        <div class="skeleton-tabs">
          <div class="skeleton-tab"></div>
          <div class="skeleton-tab"></div>
        </div>
      </div>
      <div class="skeleton-asset-selector">
        <div class="skeleton-line short"></div>
        <div class="skeleton-grid">
          <div v-for="i in 6" :key="'asset-' + i" class="skeleton-asset-btn"></div>
        </div>
      </div>
      <div class="skeleton-inputs">
        <div class="skeleton-input-group">
          <div class="skeleton-line short"></div>
          <div class="skeleton-input"></div>
        </div>
        <div class="skeleton-input-group">
          <div class="skeleton-line short"></div>
          <div class="skeleton-input"></div>
        </div>
      </div>
      <div class="skeleton-actions">
        <div class="skeleton-btn long"></div>
        <div class="skeleton-btn small"></div>
      </div>
    </div>

    <!-- 资产列表骨架 -->
    <div class="portfolio-skeleton">
      <div class="skeleton-header">
        <div class="skeleton-line medium"></div>
        <div class="skeleton-select"></div>
      </div>
      <div v-for="i in 3" :key="'asset-card-' + i" class="skeleton-asset-card">
        <div class="skeleton-card-header">
          <div class="skeleton-icon"></div>
          <div class="skeleton-info">
            <div class="skeleton-line short"></div>
            <div class="skeleton-line tiny"></div>
          </div>
          <div class="skeleton-actions">
            <div class="skeleton-btn tiny"></div>
            <div class="skeleton-btn tiny"></div>
          </div>
        </div>
        <div class="skeleton-card-body">
          <div v-for="j in 4" :key="'row-' + j" class="skeleton-row">
            <div class="skeleton-line tiny"></div>
            <div class="skeleton-line short"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- 交易历史骨架 -->
    <div class="trades-skeleton">
      <div class="skeleton-header">
        <div class="skeleton-line medium"></div>
        <div class="skeleton-actions-row">
          <div class="skeleton-btn small"></div>
          <div class="skeleton-btn small"></div>
          <div class="skeleton-select small"></div>
        </div>
      </div>
      <div v-for="i in 3" :key="'trade-' + i" class="skeleton-trade-card">
        <div class="skeleton-trade-header">
          <div class="skeleton-badge"></div>
          <div class="skeleton-line tiny"></div>
        </div>
        <div class="skeleton-trade-main">
          <div class="skeleton-icon small"></div>
          <div class="skeleton-line short"></div>
          <div class="skeleton-line short"></div>
        </div>
        <div class="skeleton-trade-details">
          <div class="skeleton-line tiny"></div>
          <div class="skeleton-line tiny"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
</script>

<style scoped>
/* CSS 变量定义 */
:root {
  --skeleton-bg: #e5e7eb;
  --skeleton-highlight: #f3f4f6;
  --skeleton-card-bg: var(--card-bg, #f9fafb);
}

/* 基础骨架动画 - 使用 transform 代替 background-position 提升性能 */
@keyframes skeleton-pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.6;
  }
}

@keyframes skeleton-shimmer {
  0% {
    transform: translateX(-100%);
  }
  100% {
    transform: translateX(100%);
  }
}

/* 骨架元素基础样式 */
.skeleton-base {
  background: var(--skeleton-bg);
  position: relative;
  overflow: hidden;
}

.skeleton-base::after {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(
    90deg,
    transparent 0%,
    var(--skeleton-highlight) 50%,
    transparent 100%
  );
  animation: skeleton-shimmer 1.5s ease-in-out infinite;
}

/* 容器 */
.skeleton-container {
  padding: 16px;
}

/* 卡片基础样式 */
.skeleton-card {
  background: var(--skeleton-card-bg);
  border-radius: 12px;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* 线条变体 */
.skeleton-line {
  composes: skeleton-base;
  height: 16px;
  border-radius: 4px;
  background: var(--skeleton-bg);
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

.skeleton-line.tiny {
  height: 12px;
  width: 60px;
}

.skeleton-line.short {
  height: 14px;
  width: 100px;
}

.skeleton-line.medium {
  height: 16px;
  width: 150px;
}

.skeleton-line.long {
  height: 24px;
  width: 100%;
  max-width: 200px;
}

/* 按钮变体 */
.skeleton-btn {
  composes: skeleton-base;
  height: 36px;
  width: 100px;
  border-radius: 8px;
  background: var(--skeleton-bg);
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

.skeleton-btn.small {
  width: 70px;
  height: 32px;
}

.skeleton-btn.tiny {
  width: 32px;
  height: 32px;
  border-radius: 8px;
}

.skeleton-btn.long {
  width: 100%;
  max-width: 200px;
}

/* 图标变体 */
.skeleton-icon {
  composes: skeleton-base;
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: var(--skeleton-bg);
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

.skeleton-icon.small {
  width: 24px;
  height: 24px;
}

/* 选择器 */
.skeleton-select {
  composes: skeleton-base;
  height: 32px;
  width: 120px;
  border-radius: 6px;
  background: var(--skeleton-bg);
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

.skeleton-select.small {
  width: 80px;
}

/* 徽章 */
.skeleton-badge {
  composes: skeleton-base;
  height: 20px;
  width: 50px;
  border-radius: 12px;
  background: var(--skeleton-bg);
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

/* 颜色块 */
.skeleton-color {
  composes: skeleton-base;
  width: 16px;
  height: 16px;
  border-radius: 4px;
  background: var(--skeleton-bg);
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

/* 标签 */
.skeleton-tab {
  composes: skeleton-base;
  height: 36px;
  width: 80px;
  border-radius: 8px;
  background: var(--skeleton-bg);
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

/* 资产按钮 */
.skeleton-asset-btn {
  composes: skeleton-base;
  height: 56px;
  border-radius: 10px;
  background: var(--skeleton-bg);
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

/* 输入框 */
.skeleton-input {
  composes: skeleton-base;
  height: 44px;
  border-radius: 10px;
  background: var(--skeleton-bg);
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

/* 图表 */
.skeleton-chart {
  composes: skeleton-base;
  width: 200px;
  height: 200px;
  border-radius: 50%;
  background: var(--skeleton-bg);
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

/* 布局区域 */
.overview-skeleton {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 20px;
}

.chart-skeleton {
  background: var(--card-bg, white);
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 20px;
}

.skeleton-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.skeleton-chart-wrapper {
  display: flex;
  align-items: center;
  gap: 40px;
}

.skeleton-legend {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.skeleton-legend-item {
  display: flex;
  align-items: center;
  gap: 10px;
}

.trading-skeleton {
  background: var(--card-bg, white);
  border-radius: 16px;
  padding: 20px;
  margin-bottom: 20px;
}

.skeleton-tabs {
  display: flex;
  gap: 8px;
}

.skeleton-asset-selector {
  margin-top: 16px;
}

.skeleton-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(80px, 1fr));
  gap: 8px;
  margin-top: 12px;
}

.skeleton-inputs {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 16px;
  margin-top: 20px;
}

.skeleton-input-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.skeleton-actions {
  display: flex;
  gap: 12px;
  margin-top: 16px;
}

.portfolio-skeleton,
.trades-skeleton {
  background: var(--card-bg, white);
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 20px;
}

.skeleton-asset-card {
  padding: 16px;
  margin-bottom: 12px;
  border-radius: 16px;
  background: rgba(0, 0, 0, 0.02);
}

.skeleton-card-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.skeleton-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.skeleton-card-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.skeleton-row {
  display: flex;
  justify-content: space-between;
}

.skeleton-trade-card {
  padding: 14px;
  margin-bottom: 10px;
  border-radius: 12px;
  background: rgba(0, 0, 0, 0.02);
}

.skeleton-trade-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.skeleton-trade-main {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 10px;
}

.skeleton-trade-details {
  display: flex;
  justify-content: space-between;
}

.skeleton-actions-row {
  display: flex;
  gap: 10px;
}

/* 暗黑模式 */
.dark .skeleton-card {
  background: rgba(30, 30, 30, 0.98);
}

.dark .skeleton-line,
.dark .skeleton-btn,
.dark .skeleton-icon,
.dark .skeleton-select,
.dark .skeleton-badge,
.dark .skeleton-color,
.dark .skeleton-tab,
.dark .skeleton-asset-btn,
.dark .skeleton-input,
.dark .skeleton-chart {
  background: #374151;
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

.dark .skeleton-asset-card,
.dark .skeleton-trade-card {
  background: rgba(255, 255, 255, 0.02);
}

/* 减少动画偏好 */
@media (prefers-reduced-motion: reduce) {
  .skeleton-line,
  .skeleton-btn,
  .skeleton-icon,
  .skeleton-select,
  .skeleton-badge,
  .skeleton-color,
  .skeleton-tab,
  .skeleton-asset-btn,
  .skeleton-input,
  .skeleton-chart {
    animation: none;
  }
}

/* 移动端适配 */
@media (max-width: 768px) {
  .skeleton-container {
    padding: 0;
  }

  .overview-skeleton {
    grid-template-columns: 1fr;
    gap: 12px;
  }

  .skeleton-chart-wrapper {
    flex-direction: column;
    gap: 20px;
  }

  .skeleton-chart {
    width: 160px;
    height: 160px;
  }

  .skeleton-legend {
    width: 100%;
  }

  .skeleton-header {
    flex-wrap: wrap;
    gap: 12px;
  }

  .skeleton-actions-row {
    flex-wrap: wrap;
  }

  .skeleton-inputs {
    grid-template-columns: 1fr;
  }

  .skeleton-actions {
    flex-wrap: wrap;
  }

  .skeleton-btn.long {
    flex: 1;
    min-width: 150px;
  }
}
</style>
