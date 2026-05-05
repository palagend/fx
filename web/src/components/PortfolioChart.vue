<template>
  <div class="portfolio-chart">
    <div class="chart-wrapper">
      <Doughnut
        v-if="allocation.length > 0 && hasLoaded"
        :data="chartData"
        :options="chartOptions"
        class="pie-chart"
        ref="chartRef"
      />
      <div class="pie-center">
        <span class="total-value">{{ formatValue(totalValue) }}</span>
        <span class="total-label">总资产</span>
      </div>
    </div>

    <div class="chart-legend">
      <div
        v-for="(item, index) in allocation"
        :key="index"
        class="legend-item"
        :class="{ active: activeIndex === index }"
        @mouseenter="handleLegendHover(index)"
        @mouseleave="handleLegendLeave"
      >
        <span class="legend-color" :style="{ backgroundColor: item.color }"></span>
        <span class="legend-name">{{ item.name }}</span>
        <div class="legend-right">
          <span class="legend-value">{{ formatValue(item.value) }}</span>
          <span class="legend-percent">{{ item.percentage }}%</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Doughnut } from 'vue-chartjs'
import { Chart as ChartJS, ArcElement } from 'chart.js'

ChartJS.register(ArcElement)

const props = defineProps({
  allocation: { type: Array, required: true },
  totalValue: { type: Number, required: true },
  hasLoaded: { type: Boolean, default: false },
  formatValue: { type: Function, required: true }
})

const chartRef = ref(null)
const activeIndex = ref(null)

const chartData = computed(() => ({
  labels: props.allocation.map(item => item.name),
  datasets: [{
    data: props.allocation.map(item => item.value),
    backgroundColor: props.allocation.map(item => item.color),
    borderWidth: 0,
    hoverOffset: 8,
    hoverBorderWidth: 0
  }]
}))

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  cutout: '65%',
  animation: { duration: 400, easing: 'easeOutQuad' },
  layout: {
    padding: 12
  },
  plugins: {
    legend: { display: false },
    tooltip: { enabled: false }
  },
  onHover: (event, elements) => {
    if (elements && elements.length > 0) {
      activeIndex.value = elements[0].index
    } else {
      activeIndex.value = null
    }
  }
}))

const handleLegendHover = (index) => {
  activeIndex.value = index
  if (chartRef.value && chartRef.value.chart) {
    chartRef.value.chart.setActiveElements([{ datasetIndex: 0, index }])
    chartRef.value.chart.update('none')
  }
}

const handleLegendLeave = () => {
  activeIndex.value = null
  if (chartRef.value && chartRef.value.chart) {
    chartRef.value.chart.setActiveElements([])
    chartRef.value.chart.update('none')
  }
}
</script>

<style scoped>
.portfolio-chart {
  display: flex;
  gap: 32px;
}

.chart-wrapper {
  position: relative;
  width: 220px;
  height: 220px;
  flex-shrink: 0;
}

.pie-chart {
  width: 100% !important;
  height: 100% !important;
}

.pie-center {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
  pointer-events: none;
}

.total-value {
  display: block;
  font-size: 22px;
  font-weight: 600;
  color: #1f2937;
  line-height: 1.3;
  letter-spacing: -0.02em;
}

.total-label {
  display: block;
  font-size: 11px;
  color: #9ca3af;
  margin-top: 2px;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.chart-legend {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
  overflow-y: auto;
  max-height: 240px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 8px;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.legend-item:hover,
.legend-item.active {
  background: rgba(67, 97, 238, 0.08);
}

.legend-color {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.legend-name {
  flex: 1;
  font-size: 13px;
  font-weight: 500;
  color: #4b5563;
}

.legend-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.legend-value {
  font-size: 13px;
  font-weight: 600;
  color: #1f2937;
}

.legend-percent {
  font-size: 12px;
  font-weight: 500;
  color: #6b7280;
  min-width: 40px;
  text-align: right;
}

.dark .total-value { color: #f9fafb; }
.dark .total-label { color: #6b7280; }
.dark .legend-name { color: #9ca3af; }
.dark .legend-value { color: #f3f4f6; }
.dark .legend-item:hover,
.dark .legend-item.active { background: rgba(99, 102, 241, 0.15); }

@media (max-width: 768px) {
  .portfolio-chart { flex-direction: column; align-items: center; gap: 20px; }
  .chart-wrapper { width: 200px; height: 200px; }
  .total-value { font-size: 20px; }
  .chart-legend { 
    width: 100%; 
    max-height: none;
    overflow-y: visible;
  }
}
</style>