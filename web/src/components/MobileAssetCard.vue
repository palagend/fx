<template>
  <div class="mobile-asset-card" :class="{ selected: selected }" @click="handleClick">
    <div class="card-header">
      <div class="asset-icon-wrapper">
        <Icon :icon="icon" :style="{ color: color }" />
      </div>
      <div class="asset-info">
        <div class="asset-name">{{ name }}</div>
        <div class="asset-symbol">{{ symbol }}</div>
      </div>
      <div class="asset-value-badge">
        {{ formatValueFn(marketValue) }}
      </div>
    </div>

    <div class="card-body">
      <div class="row">
        <div class="label">持有量</div>
        <div class="value">{{ formatAmount(amount) }} {{ symbol }}</div>
      </div>
      <div class="row">
        <div class="label">成本价</div>
        <div class="value">${{ formatAmount(avgCost) }}</div>
      </div>
      <div class="row">
        <div class="label">当前价</div>
        <div class="value price">${{ formatAmount(currentPrice) }}</div>
      </div>
      <div class="row profit-row" :class="profitClass">
        <div class="label">浮动盈亏</div>
        <div class="value">
          <span class="profit-value">{{ profitSign }}{{ formatValueFn(Math.abs(profit)) }}</span>
          <span class="profit-rate" v-if="profitRate">{{ profitRateSign }}{{ Math.abs(profitRate).toFixed(2) }}%</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Icon } from '@iconify/vue'
import { formatAmount, formatCompactAmount } from '../utils/format'

const props = defineProps({
  symbol: String,
  name: String,
  icon: String,
  color: String,
  amount: Number,
  avgCost: Number,
  currentPrice: Number,
  marketValue: Number,
  realizedPL: Number,
  selected: Boolean,
  formatValueFn: {
    type: Function,
    default: formatCompactAmount
  }
})

const emit = defineEmits(['click'])

const profit = computed(() => {
  if (props.avgCost === 0) return props.amount * props.currentPrice
  if (props.avgCost < 0) return props.amount * props.currentPrice - props.avgCost * props.amount
  return props.amount * (props.currentPrice - props.avgCost)
})

const profitRate = computed(() => {
  if (props.avgCost === 0 || props.symbol === 'USDT') return null
  if (props.avgCost < 0) return null
  return ((props.currentPrice - props.avgCost) / props.avgCost) * 100
})

const profitSign = computed(() => profit.value >= 0 ? '+' : '-')
const profitRateSign = computed(() => profitRate.value >= 0 ? '+' : '-')

const profitClass = computed(() => {
  if (props.avgCost === 0 || props.avgCost < 0) return { positive: true }
  return { positive: profit.value >= 0, negative: profit.value < 0 }
})

const handleClick = () => {
  emit('click', props.symbol)
}
</script>

<style scoped>
.mobile-asset-card {
  background: var(--card-bg, white);
  border-radius: 16px;
  padding: 16px;
  margin-bottom: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  transition: all 0.2s ease;
  border: 2px solid transparent;
}

.mobile-asset-card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
}

.mobile-asset-card.selected {
  border-color: var(--primary-color, #4361ee);
  background: rgba(67, 97, 238, 0.05);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.asset-icon-wrapper {
  width: 48px;
  height: 48px;
  border-radius: 14px;
  background: rgba(0, 0, 0, 0.05);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.asset-icon-wrapper svg {
  width: 28px;
  height: 28px;
}

.asset-info {
  flex: 1;
  min-width: 0;
}

.asset-name {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
}

.asset-symbol {
  font-size: 12px;
  color: var(--text-secondary, #6b7280);
  margin-top: 2px;
}

.asset-value-badge {
  font-size: 14px;
  font-weight: 600;
  color: #6366f1;
  background: rgba(99, 102, 241, 0.1);
  padding: 6px 12px;
  border-radius: 20px;
}

.card-body {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.row {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.row.profit-row {
  grid-column: 1 / -1;
  flex-direction: row;
  justify-content: space-between;
  align-items: center;
  padding-top: 8px;
  border-top: 1px solid rgba(0, 0, 0, 0.06);
}

.label {
  font-size: 12px;
  color: var(--text-secondary, #6b7280);
}

.value {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary, #1f2937);
}

.value.price {
  color: #6366f1;
}

.profit-value {
  font-weight: 600;
}

.profit-rate {
  font-size: 12px;
  margin-left: 8px;
  opacity: 0.8;
}

.positive {
  color: #10b981;
}

.negative {
  color: #ef4444;
}

/* 深色模式适配 */
@media (prefers-color-scheme: dark) {
  .asset-value-badge {
    background: rgba(74, 144, 226, 0.2);
    color: #4a90e2;
  }

  .row.profit-row {
    border-top-color: rgba(255, 255, 255, 0.06);
  }
}
</style>
