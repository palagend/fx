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
      <div class="quick-actions">
        <button class="action-btn btn-buy" @click.stop="handleBuy" title="买入">
          <Icon icon="mdi:arrow-down-circle" />
        </button>
        <button class="action-btn btn-sell" @click.stop="handleSell" title="卖出">
          <Icon icon="mdi:arrow-up-circle" />
        </button>
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
      <div class="row highlight">
        <div class="label">总价值</div>
        <div class="value">{{ formatCompactAmount(marketValue) }}</div>
      </div>
      <div class="row profit-row" :class="profitClass">
        <div class="label">浮动盈亏</div>
        <div class="value">
          <span class="profit-value">{{ profitSign }}{{ formatCompactAmount(Math.abs(profit)) }}</span>
          <span class="profit-rate" v-if="profitRate">{{ profitRateSign }}{{ Math.abs(profitRate).toFixed(2) }}%</span>
        </div>
      </div>
      <div v-if="realizedPL !== 0" class="row">
        <div class="label">实现盈亏</div>
        <div class="value" :class="realizedPL >= 0 ? 'positive' : 'negative'">
          {{ realizedPL >= 0 ? '+' : '-' }}{{ formatCompactAmount(Math.abs(realizedPL)) }}
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
  selected: Boolean
})

const emit = defineEmits(['click', 'buy', 'sell'])

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

const handleBuy = () => {
  emit('buy', props)
}

const handleSell = () => {
  emit('sell', props)
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

.quick-actions {
  display: flex;
  gap: 8px;
}

.action-btn {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
  touch-action: manipulation;
}

.action-btn:active {
  transform: scale(0.95);
}

.action-btn.btn-buy {
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
}

.action-btn.btn-buy:hover {
  background: rgba(16, 185, 129, 0.2);
}

.action-btn.btn-sell {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.action-btn.btn-sell:hover {
  background: rgba(239, 68, 68, 0.2);
}

.action-btn svg {
  width: 18px;
  height: 18px;
}

.card-body {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.row.highlight {
  padding-top: 10px;
  border-top: 1px dashed rgba(0, 0, 0, 0.08);
}

.row.profit-row .value {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.label {
  font-size: 13px;
  color: var(--text-secondary, #6b7280);
}

.value {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
  font-family: 'Courier New', monospace;
}

.value.price {
  color: var(--primary-color, #4361ee);
}

.value.positive {
  color: #10b981;
}

.value.negative {
  color: #ef4444;
}

.profit-value {
  font-size: 15px;
}

.profit-rate {
  font-size: 12px;
  opacity: 0.8;
}

.dark .mobile-asset-card {
  background: rgba(30, 30, 30, 0.98);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

.dark .mobile-asset-card.selected {
  background: rgba(74, 144, 226, 0.1);
}

.dark .asset-icon-wrapper {
  background: rgba(255, 255, 255, 0.05);
}

.dark .asset-name {
  color: var(--text-primary, #f3f4f6);
}

.dark .asset-symbol {
  color: var(--text-secondary, #9ca3af);
}

.dark .label {
  color: var(--text-secondary, #9ca3af);
}

.dark .value {
  color: var(--text-primary, #f3f4f6);
}

.dark .row.highlight {
  border-top-color: rgba(255, 255, 255, 0.08);
}
</style>