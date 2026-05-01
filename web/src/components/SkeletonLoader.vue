<template>
  <div class="skeleton-container">
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
.skeleton-container {
  padding: 16px;
}

.skeleton-card {
  background: var(--card-bg, #f9fafb);
  border-radius: 12px;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.skeleton-line {
  background: linear-gradient(90deg, #e5e7eb 25%, #f3f4f6 50%, #e5e7eb 75%);
  background-size: 200% 100%;
  animation: skeleton-loading 1.5s ease-in-out infinite;
  border-radius: 4px;
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

.skeleton-btn {
  background: linear-gradient(90deg, #e5e7eb 25%, #f3f4f6 50%, #e5e7eb 75%);
  background-size: 200% 100%;
  animation: skeleton-loading 1.5s ease-in-out infinite;
  border-radius: 8px;
  height: 36px;
  width: 100px;
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

.skeleton-icon {
  background: linear-gradient(90deg, #e5e7eb 25%, #f3f4f6 50%, #e5e7eb 75%);
  background-size: 200% 100%;
  animation: skeleton-loading 1.5s ease-in-out infinite;
  border-radius: 12px;
  width: 48px;
  height: 48px;
}

.skeleton-icon.small {
  width: 24px;
  height: 24px;
}

.skeleton-select {
  background: linear-gradient(90deg, #e5e7eb 25%, #f3f4f6 50%, #e5e7eb 75%);
  background-size: 200% 100%;
  animation: skeleton-loading 1.5s ease-in-out infinite;
  border-radius: 6px;
  height: 32px;
  width: 120px;
}

.skeleton-select.small {
  width: 80px;
}

.skeleton-badge {
  background: linear-gradient(90deg, #e5e7eb 25%, #f3f4f6 50%, #e5e7eb 75%);
  background-size: 200% 100%;
  animation: skeleton-loading 1.5s ease-in-out infinite;
  border-radius: 12px;
  height: 20px;
  width: 50px;
}

.skeleton-color {
  background: linear-gradient(90deg, #e5e7eb 25%, #f3f4f6 50%, #e5e7eb 75%);
  background-size: 200% 100%;
  animation: skeleton-loading 1.5s ease-in-out infinite;
  border-radius: 4px;
  width: 16px;
  height: 16px;
}

@keyframes skeleton-loading {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

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

.skeleton-chart {
  background: linear-gradient(90deg, #e5e7eb 25%, #f3f4f6 50%, #e5e7eb 75%);
  background-size: 200% 100%;
  animation: skeleton-loading 1.5s ease-in-out infinite;
  border-radius: 50%;
  width: 200px;
  height: 200px;
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

.skeleton-tab {
  background: linear-gradient(90deg, #e5e7eb 25%, #f3f4f6 50%, #e5e7eb 75%);
  background-size: 200% 100%;
  animation: skeleton-loading 1.5s ease-in-out infinite;
  border-radius: 8px;
  height: 36px;
  width: 80px;
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

.skeleton-asset-btn {
  background: linear-gradient(90deg, #e5e7eb 25%, #f3f4f6 50%, #e5e7eb 75%);
  background-size: 200% 100%;
  animation: skeleton-loading 1.5s ease-in-out infinite;
  border-radius: 10px;
  height: 56px;
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

.skeleton-input {
  background: linear-gradient(90deg, #e5e7eb 25%, #f3f4f6 50%, #e5e7eb 75%);
  background-size: 200% 100%;
  animation: skeleton-loading 1.5s ease-in-out infinite;
  border-radius: 10px;
  height: 44px;
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

.skeleton-actions {
  display: flex;
  gap: 8px;
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
.dark .skeleton-input {
  background: linear-gradient(90deg, #374151 25%, #4b5563 50%, #374151 75%);
}

.dark .skeleton-asset-card,
.dark .skeleton-trade-card {
  background: rgba(255, 255, 255, 0.02);
}

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