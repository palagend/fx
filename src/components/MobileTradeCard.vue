<template>
  <div class="mobile-trade-card">
    <div class="card-content">
      <div class="trade-header">
        <div class="trade-type-badge" :class="type">
          {{ typeText }}
        </div>
        <div class="trade-time">{{ formatDate(timestamp) }}</div>
      </div>
      
      <div class="trade-main">
        <div class="asset-info">
          <Icon :icon="icon" :style="{ color: color }" />
          <span class="asset-symbol">{{ symbol }}</span>
        </div>
        <div class="trade-amount" :class="type === 'sell' ? 'negative' : 'positive'">
          {{ type === 'sell' ? '-' : '+' }}{{ formatAmount(amount) }} {{ symbol }}
        </div>
      </div>
      
      <div class="trade-details">
        <div class="detail-row">
          <span class="label">价格</span>
          <span class="value">${{ formatAmount(price) }}</span>
        </div>
        <div class="detail-row highlight">
          <span class="label">总金额</span>
          <span class="value">${{ formatAmount(total) }}</span>
        </div>
      </div>
    </div>
    
    <button 
      v-if="canDelete" 
      class="btn-delete" 
      @click="handleDelete" 
      title="删除交易"
      :disabled="disabled"
    >
      <Icon icon="mdi:trash-can" />
    </button>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Icon } from '@iconify/vue'
import { formatAmount, formatDateTime } from '../utils/format'

const props = defineProps({
  id: String,
  symbol: String,
  type: String,
  amount: Number,
  price: Number,
  total: Number,
  timestamp: String,
  icon: String,
  color: String,
  canDelete: Boolean,
  disabled: Boolean
})

const emit = defineEmits(['delete'])

const typeText = computed(() => {
  const map = {
    buy: '买入',
    sell: '卖出',
    recharge: '充值'
  }
  return map[props.type] || props.type
})

const formatDate = formatDateTime

const handleDelete = () => {
  emit('delete', props.id)
}
</script>

<style scoped>
.mobile-trade-card {
  background: var(--card-bg, white);
  border-radius: 12px;
  padding: 14px;
  margin-bottom: 10px;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.04);
  display: flex;
  align-items: center;
  gap: 12px;
}

.card-content {
  flex: 1;
  min-width: 0;
}

.trade-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.trade-type-badge {
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
}

.trade-type-badge.buy {
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
}

.trade-type-badge.sell {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.trade-type-badge.recharge {
  background: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
}

.trade-time {
  font-size: 11px;
  color: var(--text-secondary, #9ca3af);
  font-family: 'Courier New', monospace;
}

.trade-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.asset-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.asset-info svg {
  width: 20px;
  height: 20px;
}

.asset-symbol {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary, #1f2937);
}

.trade-amount {
  font-size: 15px;
  font-weight: 600;
  font-family: 'Courier New', monospace;
}

.trade-amount.positive {
  color: #10b981;
}

.trade-amount.negative {
  color: #ef4444;
}

.trade-details {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.detail-row {
  display: flex;
  justify-content: space-between;
}

.detail-row .label {
  font-size: 12px;
  color: var(--text-secondary, #6b7280);
}

.detail-row .value {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-primary, #4b5563);
  font-family: 'Courier New', monospace;
}

.detail-row.highlight .value {
  color: var(--primary-color, #4361ee);
  font-weight: 600;
}

.btn-delete {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  border: none;
  background: rgba(239, 68, 68, 0.08);
  color: #9ca3af;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
  flex-shrink: 0;
}

.btn-delete:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}

.btn-delete:active:not(:disabled) {
  transform: scale(0.95);
}

.btn-delete:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-delete svg {
  width: 16px;
  height: 16px;
}

.dark .mobile-trade-card {
  background: rgba(30, 30, 30, 0.98);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.15);
}

.dark .asset-symbol {
  color: var(--text-primary, #f3f4f6);
}

.dark .trade-time {
  color: var(--text-secondary, #6b7280);
}

.dark .detail-row .label {
  color: var(--text-secondary, #9ca3af);
}

.dark .detail-row .value {
  color: var(--text-primary, #d1d5db);
}

.dark .btn-delete {
  background: rgba(239, 68, 68, 0.1);
}
</style>