<template>
  <div class="trade-preview" v-if="showPreview">
    <div class="preview-header-row">
      <span class="preview-title">交易预览</span>
      <span class="preview-total-value">${{ formatAmount(total) }}</span>
    </div>
    <div class="preview-details-full">
      <!-- 买入预览 -->
      <template v-if="tradeType === 'buy'">
        <div class="preview-detail-row" v-if="currentHolding > 0">
          <span class="detail-label">当前持仓</span>
          <span class="detail-value">{{ formatAmount(currentHolding) }}</span>
        </div>
        <div class="preview-detail-row" v-if="currentHolding > 0">
          <span class="detail-label">买入后持仓</span>
          <span class="detail-value highlight">{{ formatAmount(currentHolding + amount) }}</span>
        </div>
        <div class="preview-detail-row">
          <span class="detail-label">买入后成本价</span>
          <span class="detail-value highlight">${{ formatAmount(newAvgCost) }}</span>
        </div>
        <div class="preview-detail-row impact">
          <span class="detail-label">USD支出</span>
          <span class="detail-value negative">-${{ formatAmount(total) }}</span>
        </div>
      </template>

      <!-- 卖出预览 -->
      <template v-if="tradeType === 'sell'">
        <div class="preview-detail-row">
          <span class="detail-label">当前持仓</span>
          <span class="detail-value">{{ formatAmount(currentHolding) }}</span>
        </div>
        <div class="preview-detail-row">
          <span class="detail-label">卖出后持仓</span>
          <span class="detail-value">{{ formatAmount(Math.max(0, currentHolding - amount)) }}</span>
        </div>
        <div class="preview-detail-row" v-if="estimatedPL !== 0">
          <span class="detail-label">预估盈亏</span>
          <span :class="['detail-value', estimatedPL >= 0 ? 'positive' : 'negative']">
            {{ estimatedPL >= 0 ? '+' : '-' }}${{ formatAmount(Math.abs(estimatedPL)) }}
          </span>
        </div>
        <div class="preview-detail-row impact">
          <span class="detail-label">USD收入</span>
          <span class="detail-value positive">+${{ formatAmount(total) }}</span>
        </div>
      </template>
    </div>
  </div>

  <!-- 空状态 -->
  <div class="trade-preview empty" v-else-if="showEmpty">
    <Icon icon="mdi:calculator-variant" class="empty-icon" />
    <p>填写交易信息查看预览</p>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Icon } from '@iconify/vue'
import { formatAmount } from '../../utils/format'

const props = defineProps({
  tradeType: {
    type: String,
    required: true
  },
  symbol: {
    type: String,
    default: ''
  },
  amount: {
    type: Number,
    default: 0
  },
  price: {
    type: Number,
    default: 0
  },
  currentHolding: {
    type: Number,
    default: 0
  },
  currentAvgCost: {
    type: Number,
    default: 0
  },
  showEmpty: {
    type: Boolean,
    default: false
  }
})

// 交易总额
const total = computed(() => props.amount * props.price)

// 是否显示预览
const showPreview = computed(() => {
  return props.symbol && props.amount > 0 && props.price > 0
})

// 预估实现盈亏（卖出时）
// 实现盈亏 = USD收入 - USD成本（按卖出比例计算）
const estimatedPL = computed(() => {
  if (props.tradeType !== 'sell') return 0
  if (!props.currentHolding || props.currentHolding === 0) return 0

  // 本次卖出获得的USD
  const usdOut = props.amount * props.price

  // 按卖出比例计算的USD投入成本
  const totalCost = props.currentAvgCost * props.currentHolding
  const costRatio = props.amount / props.currentHolding
  const usdIn = totalCost * costRatio

  return usdOut - usdIn
})

// 新成本价（买入时）
const newAvgCost = computed(() => {
  if (props.tradeType !== 'buy') return props.currentAvgCost

  const tradeTotal = props.amount * props.price
  const currentTotalCost = props.currentAvgCost * props.currentHolding
  const newTotalCost = currentTotalCost + tradeTotal
  const newTotalAmount = props.currentHolding + props.amount

  return newTotalAmount > 0 ? newTotalCost / newTotalAmount : 0
})
</script>

<style scoped>
.trade-preview {
  background: var(--bg-secondary);
  border-radius: 12px;
  padding: 1.25rem;
  border: 1px solid var(--border-color);
}

.trade-preview.empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 200px;
  color: var(--text-secondary);
}

.empty-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
  opacity: 0.5;
}

.preview-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--border-color);
}

.preview-title {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-primary);
}

.preview-total-value {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--primary-color);
}

.preview-details-full {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.preview-detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.9rem;
}

.preview-detail-row.impact {
  margin-top: 0.5rem;
  padding-top: 0.75rem;
  border-top: 1px dashed var(--border-color);
}

.detail-label {
  color: var(--text-secondary);
}

.detail-value {
  font-weight: 600;
  color: var(--text-primary);
}

.detail-value.highlight {
  color: var(--primary-color);
}

.detail-value.positive {
  color: #10b981;
}

.detail-value.negative {
  color: #ef4444;
}
</style>
