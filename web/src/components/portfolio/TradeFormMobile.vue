<template>
  <TradeFormBase
    :trade-type="tradeType"
    :asset-type="assetType"
    :selected-symbol="selectedSymbol"
    :amount="amount"
    :price="price"
    :current-market-price="currentMarketPrice"
    :cash-balance="cashBalance"
    :is-loading="isLoading"
    :is-submitting="isSubmitting"
    @update:trade-type="$emit('update:tradeType', $event)"
    @update:asset-type="$emit('update:assetType', $event)"
    @update:selected-symbol="$emit('update:selectedSymbol', $event)"
    @update:amount="$emit('update:amount', $event)"
    @update:price="$emit('update:price', $event)"
    @submit="$emit('submit', $event)"
    @reset="$emit('reset')"
    @symbol-select="$emit('symbolSelect', $event)"
    v-slot="slotProps"
  >
    <div class="trade-form-mobile">
      <!-- 交易类型选择（无标题） -->
      <div class="trade-type-tabs">
        <button
          :class="['type-tab', { active: slotProps.tradeType === 'buy' }]"
          @click="slotProps.updateTradeType('buy')"
        >
          <Icon icon="mdi:arrow-down-circle" /> 买入
        </button>
        <button
          :class="['type-tab', { active: slotProps.tradeType === 'sell' }]"
          @click="slotProps.updateTradeType('sell')"
        >
          <Icon icon="mdi:arrow-up-circle" /> 卖出
        </button>
      </div>

      <!-- 资产类型选择 -->
      <div class="asset-type-selector">
        <label class="field-label">资产类型</label>
        <div class="asset-type-tabs">
          <button
            :class="['type-tab', { active: slotProps.assetType === 'crypto' }]"
            @click.stop="slotProps.switchAssetType('crypto')"
          >
            <Icon icon="cryptocurrency:btc" /> 加密货币
          </button>
          <button
            :class="['type-tab', { active: slotProps.assetType === 'us_stock' }]"
            @click.stop="slotProps.switchAssetType('us_stock')"
          >
            <Icon icon="mdi:chart-line" /> 美股
          </button>
        </div>
      </div>

      <!-- 币种选择网格 - 仅在未选择资产或点击更改时显示 -->
      <div class="asset-selector" v-if="!slotProps.selectedSymbol || slotProps.showAssetSelector">
        <label class="field-label">选择资产</label>
        <div class="asset-grid">
          <button
            v-for="symbol in slotProps.availableSymbols"
            :key="symbol"
            :class="['asset-btn', { active: slotProps.selectedSymbol === symbol }]"
            @click.stop="slotProps.selectSymbol(symbol)"
          >
            <Icon :icon="slotProps.getAssetIcon(slotProps.assetType, symbol)" :style="{ color: slotProps.getAssetColor(slotProps.assetType, symbol) }" />
            <span class="asset-code">{{ symbol }}</span>
            <span class="asset-price" v-if="slotProps.getCurrentPrice(symbol)">
              ${{ slotProps.formatAmount(slotProps.getCurrentPrice(symbol)) }}
            </span>
          </button>
        </div>
      </div>

      <!-- 已选择资产显示 & 更改按钮 -->
      <div class="selected-asset-display" v-if="slotProps.selectedSymbol && !slotProps.showAssetSelector">
        <div class="selected-asset-info">
          <Icon :icon="slotProps.getAssetIcon(slotProps.assetType, slotProps.selectedSymbol)" :style="{ color: slotProps.getAssetColor(slotProps.assetType, slotProps.selectedSymbol) }" />
          <span class="selected-asset-name">{{ slotProps.selectedSymbol }}</span>
          <span class="selected-asset-price" v-if="slotProps.getCurrentPrice(slotProps.selectedSymbol)">
            ${{ slotProps.formatAmount(slotProps.getCurrentPrice(slotProps.selectedSymbol)) }}
          </span>
        </div>
        <button class="btn-change-asset" @click="slotProps.openAssetSelector">
          <Icon icon="mdi:swap-horizontal" /> 更改
        </button>
      </div>

      <!-- 交易输入区 -->
      <div class="trade-inputs" v-if="slotProps.selectedSymbol">
        <div class="input-field">
          <label class="field-label">
            数量
            <span class="field-hint" v-if="slotProps.tradeType === 'sell'">
              可卖: {{ slotProps.formatAmount(slotProps.getHoldingAmount(slotProps.selectedSymbol)) }}
            </span>
          </label>
          <div class="input-with-controls">
            <input
              type="number"
              inputmode="decimal"
              :value="slotProps.amount"
              @input="slotProps.updateAmount(parseFloat($event.target.value))"
              placeholder="0.00"
              min="0.00001"
              step="0.00001"
              ref="amountInputRef"
              class="mobile-number-input"
            >
            <span class="input-unit">{{ slotProps.selectedSymbol }}</span>
          </div>
          <!-- 快捷输入按钮 -->
          <div class="quick-amount-buttons" v-if="slotProps.tradeType === 'sell' && slotProps.getHoldingAmount(slotProps.selectedSymbol) > 0">
            <button class="quick-btn primary" @click="slotProps.setQuickAmount(100)">100%</button>
            <button class="quick-btn" @click="slotProps.setQuickAmount(75)">75%</button>
            <button class="quick-btn" @click="slotProps.setQuickAmount(50)">50%</button>
            <button class="quick-btn" @click="slotProps.setQuickAmount(25)">25%</button>
          </div>
          <div class="quick-amount-buttons" v-else-if="slotProps.tradeType === 'buy' && slotProps.cashBalance > 0">
            <button class="quick-btn primary" @click="slotProps.setQuickBuyAmount(100)">100%</button>
            <button class="quick-btn" @click="slotProps.setQuickBuyAmount(75)">75%</button>
            <button class="quick-btn" @click="slotProps.setQuickBuyAmount(50)">50%</button>
            <button class="quick-btn" @click="slotProps.setQuickBuyAmount(25)">25%</button>
          </div>
        </div>

        <div class="input-field">
          <label class="field-label">
            价格
            <span class="field-hint" v-if="slotProps.currentMarketPrice > 0">
              市价: ${{ slotProps.formatAmount(slotProps.currentMarketPrice) }}
            </span>
          </label>
          <div class="input-with-controls">
            <input
              type="number"
              inputmode="decimal"
              :value="slotProps.price"
              @input="slotProps.updatePrice(parseFloat($event.target.value))"
              placeholder="0.00"
              min="0.00001"
              step="0.00001"
              class="mobile-number-input"
            >
            <button
              class="btn-use-market"
              @click="slotProps.useMarketPrice"
              v-if="slotProps.currentMarketPrice > 0"
            >
              使用市价
            </button>
          </div>
        </div>
      </div>
    </div>
  </TradeFormBase>
</template>

<script setup>
import { ref } from 'vue'
import { Icon } from '@iconify/vue'
import TradeFormBase from './TradeFormBase.vue'

const props = defineProps({
  tradeType: {
    type: String,
    default: 'buy'
  },
  assetType: {
    type: String,
    default: 'crypto'
  },
  selectedSymbol: {
    type: String,
    default: ''
  },
  amount: {
    type: Number,
    default: null
  },
  price: {
    type: Number,
    default: null
  },
  currentMarketPrice: {
    type: Number,
    default: 0
  },
  cashBalance: {
    type: Number,
    default: 0
  },
  isLoading: {
    type: Boolean,
    default: false
  },
  isSubmitting: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits([
  'update:tradeType',
  'update:assetType',
  'update:selectedSymbol',
  'update:amount',
  'update:price',
  'submit',
  'reset',
  'symbolSelect'
])

const amountInputRef = ref(null)
</script>

<style scoped>
.trade-form-mobile {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.trade-type-tabs {
  display: flex;
  gap: 0.5rem;
}

.type-tab {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  padding: 0.625rem 1rem;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--bg-secondary);
  color: var(--text-secondary);
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.type-tab:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.type-tab.active {
  background: var(--primary-color);
  border-color: var(--primary-color);
  color: white;
}

.asset-type-selector {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.field-label {
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.asset-type-tabs {
  display: flex;
  gap: 0.5rem;
}

.asset-selector {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.asset-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.5rem;
}

.asset-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.25rem;
  padding: 0.625rem 0.375rem;
  border: 1px solid var(--border-color);
  border-radius: 10px;
  background: var(--bg-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.asset-btn:hover {
  border-color: var(--primary-color);
  transform: translateY(-1px);
}

.asset-btn.active {
  border-color: var(--primary-color);
  background: var(--primary-color);
  color: white;
}

/* 资产按钮图标放大 */
.asset-btn .iconify {
  font-size: 28px;
}

.asset-code {
  font-weight: 600;
  font-size: 0.8rem;
}

.asset-price {
  font-size: 0.7rem;
  opacity: 0.8;
}

/* 已选择资产显示 */
.selected-asset-display {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.625rem 0.875rem;
  background: var(--bg-secondary);
  border-radius: 10px;
  border: 1px solid var(--border-color);
}

.selected-asset-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.selected-asset-name {
  font-weight: 600;
  font-size: 0.95rem;
  color: var(--text-primary);
}

.selected-asset-price {
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.btn-change-asset {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.4rem 0.75rem;
  border: 1px solid var(--primary-color);
  border-radius: 8px;
  background: transparent;
  color: var(--primary-color);
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-change-asset:hover {
  background: var(--primary-color);
  color: white;
}

.trade-inputs {
  display: flex;
  flex-direction: column;
  gap: 0.875rem;
}

.input-field {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.field-hint {
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin-left: auto;
}

.input-with-controls {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.mobile-number-input {
  flex: 1;
  padding: 0.625rem 0.875rem;
  border: 1px solid var(--border-color);
  border-radius: 10px;
  font-size: 1rem;
  background: var(--bg-secondary);
  color: var(--text-primary);
  min-width: 0;
}

.mobile-number-input:focus {
  outline: none;
  border-color: var(--primary-color);
}

.input-unit {
  font-weight: 600;
  color: var(--text-secondary);
  min-width: 50px;
  text-align: right;
  font-size: 0.9rem;
}

.btn-use-market {
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--primary-color);
  border-radius: 8px;
  background: transparent;
  color: var(--primary-color);
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn-use-market:hover {
  background: var(--primary-color);
  color: white;
}

.quick-amount-buttons {
  display: flex;
  gap: 0.375rem;
  flex-wrap: wrap;
}

.quick-btn {
  flex: 1;
  min-width: 50px;
  padding: 0.35rem 0.5rem;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  background: var(--bg-secondary);
  color: var(--text-secondary);
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}

.quick-btn:hover {
  border-color: var(--primary-color);
  color: var(--primary-color);
}

.quick-btn.primary {
  background: var(--primary-color);
  border-color: var(--primary-color);
  color: white;
}

.quick-btn.primary:hover {
  opacity: 0.9;
}
</style>
