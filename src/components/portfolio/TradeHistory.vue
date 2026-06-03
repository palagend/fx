<template>
  <div class="trade-history">
    <!-- 头部操作栏 -->
    <div class="section-header">
      <h2 class="section-title"><Icon icon="mdi:history" /> 交易历史</h2>
      <div class="section-actions">
        <button class="btn-export" @click="$emit('export')" :disabled="isSubmittingExport">
          <Icon icon="mdi:download" /> 导出
        </button>
        <button class="btn-import" @click="$emit('import')">
          <Icon icon="mdi:upload" /> 导入
        </button>
        <div class="protect-switch" @click="toggleProtect">
          <Icon :icon="protectHistory ? 'mdi:shield-check' : 'mdi:shield-off'" />
          <span class="switch-label">保护</span>
          <div class="switch" :class="{ 'on': protectHistory }">
            <div class="switch-handle"></div>
          </div>
        </div>
        <div class="filter-group">
          <select :value="filter" @change="$emit('update:filter', $event.target.value)" class="filter-select">
            <option value="all">全部交易</option>
            <option value="buy">买入</option>
            <option value="sell">卖出</option>
            <option value="recharge">充值</option>
          </select>
        </div>
        <button 
          class="btn-clear" 
          @click="$emit('clear')" 
          v-if="trades.length > 0 && !protectHistory" 
          :disabled="isSubmittingClear"
        >
          <Icon icon="mdi:delete-sweep" /> 清空历史
        </button>
      </div>
    </div>

    <!-- PC端表格视图 -->
    <div class="table-wrapper desktop-view">
      <table class="trades-table">
        <thead>
          <tr>
            <th>时间</th>
            <th>资产</th>
            <th>类型</th>
            <th>数量</th>
            <th class="trade-price-col">价格</th>
            <th>总金额</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="trade in tradesWithMeta" :key="trade.id" class="trade-row">
            <td class="trade-time">{{ formatDateTime(trade.created_at || trade.timestamp) }}</td>
            <td>
              <div class="trade-asset">
                <Icon :icon="trade.icon" :style="{ color: trade.color }" />
                <span>{{ trade.symbol }}</span>
              </div>
            </td>
            <td>
              <span class="trade-type" :class="trade.type">
                {{ getTradeTypeText(trade.type) }}
              </span>
            </td>
            <td>{{ trade.formattedAmount }}</td>
            <td class="trade-price-col">${{ trade.formattedPrice }}</td>
            <td>${{ formatAmount(trade.total) }}</td>
            <td>
              <button 
                class="btn-delete" 
                @click="$emit('delete', trade.id)" 
                :disabled="protectHistory || isSubmittingDelete" 
                title="删除"
              >
                <Icon icon="mdi:trash-can" />
              </button>
            </td>
          </tr>
          <tr v-if="tradesWithMeta.length === 0">
            <td colspan="7" class="empty-state">
              <Icon icon="mdi:inbox" />
              <p>暂无交易记录</p>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 移动端卡片视图 -->
    <div class="mobile-trade-list mobile-view">
      <MobileTradeCard
        v-for="trade in tradesWithMeta"
        :key="trade.id"
        :id="trade.id"
        :symbol="trade.symbol"
        :type="trade.type"
        :amount="trade.amount"
        :price="trade.price"
        :total="trade.total"
        :timestamp="trade.created_at || trade.timestamp"
        :icon="trade.icon"
        :color="trade.color"
        :can-delete="!protectHistory"
        :disabled="isSubmittingDelete"
        @delete="$emit('delete', $event)"
      />
      <div v-if="tradesWithMeta.length === 0" class="empty-state mobile-empty">
        <Icon icon="mdi:inbox" />
        <p>暂无交易记录</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Icon } from '@iconify/vue'
import { formatAmount, formatDateTime } from '../../utils/format'
import { getAssetColor, getAssetIcon } from '../../config/assets'
import MobileTradeCard from '../MobileTradeCard.vue'

const props = defineProps({
  trades: {
    type: Array,
    default: () => []
  },
  filter: {
    type: String,
    default: 'all'
  },
  protectHistory: {
    type: Boolean,
    default: true
  },
  isSubmittingExport: {
    type: Boolean,
    default: false
  },
  isSubmittingClear: {
    type: Boolean,
    default: false
  },
  isSubmittingDelete: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits([
  'update:filter',
  'update:protectHistory',
  'delete',
  'clear',
  'export',
  'import'
])

// 过滤后的交易列表
const filteredTrades = computed(() => {
  const filter = props.filter
  return filter === 'all' 
    ? props.trades 
    : props.trades?.filter(t => t.type === filter) || []
})

// 交易类型文本
const getTradeTypeText = (type) => {
  const map = {
    buy: '买入',
    sell: '卖出',
    recharge: '充值'
  }
  return map[type] || type
}

// 切换保护状态
const toggleProtect = () => {
  emit('update:protectHistory', !props.protectHistory)
}

// 计算属性缓存工具
function createComputedCache() {
  let cache = null
  let lastKey = ''

  return {
    get: (key, compute) => {
      if (key !== lastKey || cache === null) {
        cache = compute()
        lastKey = key
      }
      return cache
    },
    clear: () => {
      cache = null
      lastKey = ''
    }
  }
}

// 缓存交易元数据
const tradesWithMetaCache = createComputedCache()

function computeTradesWithMeta() {
  return filteredTrades.value.map(item => ({
    ...item,
    icon: getAssetIcon(item.asset_type, item.symbol),
    color: getAssetColor(item.asset_type, item.symbol),
    formattedAmount: formatAmount(item.amount),
    formattedPrice: formatAmount(item.price),
    formattedTotal: formatAmount(item.amount * item.price),
    formattedFee: formatAmount(item.fee || 0)
  }))
}

const tradesWithMeta = computed(() => {
  const cacheKey = `${props.filter}-${props.trades?.length || 0}`
  return tradesWithMetaCache.get(cacheKey, computeTradesWithMeta)
})
</script>

<style scoped>
.trade-history {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0;
  font-size: 1.25rem;
  color: var(--text-primary);
}

.section-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.btn-export,
.btn-import {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.5rem 1rem;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--bg-secondary);
  color: var(--text-secondary);
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-export:hover,
.btn-import:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.btn-export:disabled,
.btn-import:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.protect-switch {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.5rem 0.75rem;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.protect-switch:hover {
  border-color: var(--primary-color);
}

.switch-label {
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.switch {
  width: 36px;
  height: 20px;
  background: #e5e7eb;
  border-radius: 10px;
  position: relative;
  transition: background 0.2s;
}

.switch.on {
  background: #10b981;
}

.switch-handle {
  width: 16px;
  height: 16px;
  background: white;
  border-radius: 50%;
  position: absolute;
  top: 2px;
  left: 2px;
  transition: transform 0.2s;
}

.switch.on .switch-handle {
  transform: translateX(16px);
}

.filter-group {
  display: flex;
  align-items: center;
}

.filter-select {
  padding: 0.5rem 2rem 0.5rem 0.75rem;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--bg-secondary);
  color: var(--text-primary);
  font-size: 0.9rem;
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%236b7280' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 0.5rem center;
}

.btn-clear {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.5rem 1rem;
  border: 1px solid #ef4444;
  border-radius: 8px;
  background: transparent;
  color: #ef4444;
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-clear:hover:not(:disabled) {
  background: #ef4444;
  color: white;
}

.btn-clear:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.table-wrapper {
  overflow-x: auto;
  border-radius: 12px;
  border: 1px solid var(--border-color);
}

.trades-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}

.trades-table th {
  padding: 0.875rem 1rem;
  text-align: left;
  font-weight: 600;
  color: var(--text-secondary);
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
}

.trades-table td {
  padding: 0.875rem 1rem;
  border-bottom: 1px solid var(--border-color);
  color: var(--text-primary);
}

.trades-table tbody tr:last-child td {
  border-bottom: none;
}

.trades-table tbody tr:hover {
  background: var(--bg-secondary);
}

.trade-time {
  font-size: 0.85rem;
  color: var(--text-secondary);
  white-space: nowrap;
}

.trade-asset {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 500;
}

.trade-type {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.75rem;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 500;
}

.trade-type.buy {
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
}

.trade-type.sell {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.trade-type.recharge {
  background: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
}

.trade-price-col {
  text-align: right;
}

.btn-delete {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: #ef4444;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-delete:hover:not(:disabled) {
  background: rgba(239, 68, 68, 0.1);
}

.btn-delete:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.empty-state {
  text-align: center;
  padding: 3rem 1rem;
  color: var(--text-secondary);
}

.empty-state svg {
  font-size: 3rem;
  margin-bottom: 1rem;
  opacity: 0.5;
}

.empty-state p {
  margin: 0;
  font-size: 0.95rem;
}

.mobile-trade-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.mobile-empty {
  padding: 3rem 1rem;
  background: var(--bg-secondary);
  border-radius: 12px;
  border: 1px dashed var(--border-color);
}

/* 响应式 */
@media (max-width: 768px) {
  .section-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .section-actions {
    width: 100%;
    justify-content: flex-start;
  }

  .desktop-view {
    display: none;
  }

  .mobile-view {
    display: flex;
  }
}

@media (min-width: 769px) {
  .mobile-view {
    display: none;
  }
}
</style>
