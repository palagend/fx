<template>
  <div class="skeleton-container" aria-busy="true" aria-label="加载中">
    <!-- 移动端顶部导航骨架 -->
    <div class="mobile-top-nav-skeleton">
      <div v-for="i in 4" :key="'tab-' + i" class="skeleton-tab-item">
        <div class="skeleton-icon tiny"></div>
        <div class="skeleton-line tiny"></div>
      </div>
    </div>

    <!-- 概览标签页骨架 -->
    <div class="overview-skeleton">
      <!-- 总资产卡片 -->
      <div class="skeleton-card total-card">
        <div class="skeleton-card-header-row">
          <div class="skeleton-line short"></div>
          <div class="skeleton-toggle"></div>
        </div>
        <div class="skeleton-line large"></div>
        <div class="skeleton-line medium"></div>
      </div>
      <!-- 其他概览卡片 -->
      <div class="overview-grid">
        <div v-for="i in 3" :key="'overview-card-' + i" class="skeleton-card">
          <div class="skeleton-line short"></div>
          <div class="skeleton-line long"></div>
          <div class="skeleton-line medium"></div>
        </div>
      </div>
    </div>

    <!-- 资产分布标签页骨架 -->
    <div class="chart-skeleton">
      <div class="skeleton-header">
        <div class="skeleton-line medium"></div>
        <div class="skeleton-view-toggle">
          <div v-for="i in 4" :key="'view-' + i" class="skeleton-view-btn"></div>
        </div>
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

    <!-- 资产详情标签页骨架 -->
    <div class="portfolio-skeleton">
      <div class="skeleton-header">
        <div class="skeleton-line medium"></div>
        <div class="skeleton-select"></div>
      </div>
      <div class="mobile-asset-cards">
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
            <div v-for="j in 3" :key="'row-' + j" class="skeleton-row">
              <div class="skeleton-line tiny"></div>
              <div class="skeleton-line short"></div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 历史标签页骨架 -->
    <div class="trades-skeleton">
      <div class="skeleton-header">
        <div class="skeleton-line medium"></div>
        <div class="skeleton-actions-row">
          <div class="skeleton-btn small"></div>
          <div class="skeleton-btn small"></div>
          <div class="skeleton-select small"></div>
        </div>
      </div>
      <div class="mobile-trade-cards">
        <div v-for="i in 3" :key="'trade-' + i" class="skeleton-trade-card">
          <div class="skeleton-trade-header">
            <div class="skeleton-badge"></div>
            <div class="skeleton-line tiny"></div>
          </div>
          <div class="skeleton-trade-main">
            <div class="skeleton-icon small"></div>
            <div class="skeleton-trade-info">
              <div class="skeleton-line short"></div>
              <div class="skeleton-line tiny"></div>
            </div>
            <div class="skeleton-line short"></div>
          </div>
          <div class="skeleton-trade-details">
            <div class="skeleton-line tiny"></div>
            <div class="skeleton-line tiny"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- 移动端浮动按钮骨架 -->
    <div class="mobile-fab-skeleton">
      <div class="skeleton-fab"></div>
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

/* 基础骨架动画 */
@keyframes skeleton-pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.6;
  }
}

/* 容器 */
.skeleton-container {
  padding: 16px;
}

/* 线条变体 */
.skeleton-line {
  background: var(--skeleton-bg);
  border-radius: 4px;
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
  height: 20px;
  width: 100%;
  max-width: 180px;
}

.skeleton-line.large {
  height: 32px;
  width: 100%;
  max-width: 200px;
}

/* 按钮变体 */
.skeleton-btn {
  background: var(--skeleton-bg);
  height: 36px;
  width: 100px;
  border-radius: 8px;
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

/* 图标变体 */
.skeleton-icon {
  background: var(--skeleton-bg);
  width: 48px;
  height: 48px;
  border-radius: 12px;
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

.skeleton-icon.small {
  width: 24px;
  height: 24px;
  border-radius: 6px;
}

.skeleton-icon.tiny {
  width: 20px;
  height: 20px;
  border-radius: 4px;
}

/* 选择器 */
.skeleton-select {
  background: var(--skeleton-bg);
  height: 32px;
  width: 120px;
  border-radius: 6px;
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

.skeleton-select.small {
  width: 80px;
}

/* 徽章 */
.skeleton-badge {
  background: var(--skeleton-bg);
  height: 20px;
  width: 50px;
  border-radius: 12px;
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

/* 颜色块 */
.skeleton-color {
  background: var(--skeleton-bg);
  width: 16px;
  height: 16px;
  border-radius: 4px;
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

/* 切换按钮 */
.skeleton-toggle {
  background: var(--skeleton-bg);
  height: 32px;
  width: 100px;
  border-radius: 16px;
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

/* 视图切换 */
.skeleton-view-toggle {
  display: flex;
  gap: 4px;
  background: var(--skeleton-bg);
  padding: 4px;
  border-radius: 10px;
}

.skeleton-view-btn {
  background: var(--skeleton-bg);
  height: 32px;
  width: 60px;
  border-radius: 8px;
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

/* 标签项 */
.skeleton-tab-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
}

/* 移动端顶部导航 */
.mobile-top-nav-skeleton {
  display: none;
  justify-content: space-around;
  background: var(--card-bg, white);
  border-radius: 12px;
  padding: 8px;
  margin-bottom: 16px;
}

/* 概览区域 */
.overview-skeleton {
  margin-bottom: 20px;
}

.overview-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-top: 16px;
}

.skeleton-card {
  background: var(--skeleton-card-bg);
  border-radius: 12px;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.skeleton-card.total-card {
  gap: 16px;
}

.skeleton-card-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

/* 图表区域 */
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
  flex-wrap: wrap;
  gap: 12px;
}

.skeleton-chart-wrapper {
  display: flex;
  align-items: center;
  gap: 40px;
}

.skeleton-chart {
  background: var(--skeleton-bg);
  width: 200px;
  height: 200px;
  border-radius: 50%;
  animation: skeleton-pulse 1.5s ease-in-out infinite;
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

/* 资产列表区域 */
.portfolio-skeleton {
  background: var(--card-bg, white);
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 20px;
}

.mobile-asset-cards {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.skeleton-asset-card {
  padding: 16px;
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

/* 交易历史区域 */
.trades-skeleton {
  background: var(--card-bg, white);
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 20px;
}

.mobile-trade-cards {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.skeleton-trade-card {
  padding: 14px;
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

.skeleton-trade-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.skeleton-trade-details {
  display: flex;
  justify-content: space-between;
}

.skeleton-actions-row {
  display: flex;
  gap: 10px;
}

/* 浮动按钮 */
.mobile-fab-skeleton {
  display: none;
  position: fixed;
  bottom: 24px;
  right: 24px;
}

.skeleton-fab {
  background: var(--skeleton-bg);
  width: 56px;
  height: 56px;
  border-radius: 50%;
  animation: skeleton-pulse 1.5s ease-in-out infinite;
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
.dark .skeleton-toggle,
.dark .skeleton-view-btn,
.dark .skeleton-chart,
.dark .skeleton-fab {
  background: #374151;
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
  .skeleton-toggle,
  .skeleton-view-btn,
  .skeleton-chart,
  .skeleton-fab {
    animation: none;
  }
}

/* 移动端适配 */
@media (max-width: 768px) {
  .skeleton-container {
    padding: 0;
  }

  /* 显示移动端导航 */
  .mobile-top-nav-skeleton {
    display: flex;
  }

  /* 概览区域调整为单列 */
  .overview-grid {
    grid-template-columns: 1fr;
    gap: 12px;
  }

  .skeleton-card.total-card {
    margin-bottom: 12px;
  }

  /* 图表区域调整 */
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

  .skeleton-view-toggle {
    width: 100%;
    justify-content: space-between;
  }

  .skeleton-view-btn {
    flex: 1;
    height: 36px;
  }

  /* 头部调整 */
  .skeleton-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .skeleton-actions-row {
    flex-wrap: wrap;
    width: 100%;
  }

  /* 显示浮动按钮 */
  .mobile-fab-skeleton {
    display: block;
  }
}
</style>
