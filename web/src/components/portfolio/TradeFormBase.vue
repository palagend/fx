<script setup>
import { ref, computed, nextTick } from 'vue'
import { usePortfolioStore } from '../../stores/portfolio'
import { AVAILABLE_SYMBOLS, AVAILABLE_ASSETS, getAssetColor, getAssetIcon } from '../../config/assets'
import { formatAmount, formatCompactAmount } from '../../utils/format'

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

const portfolioStore = usePortfolioStore()
const amountInputRef = ref(null)

// 移动端：资产选择器显示状态
const showAssetSelector = ref(true)

// 支持的资产列表
const availableSymbols = computed(() => {
  return props.assetType === 'crypto' ? AVAILABLE_SYMBOLS : AVAILABLE_ASSETS.US_STOCK
})

// 获取当前价格
const getCurrentPrice = (symbol) => {
  if (props.assetType === 'crypto') {
    return portfolioStore.prices[symbol]
  } else {
    return portfolioStore.usStockPrices[symbol]
  }
}

// 获取持仓数量
const getHoldingAmount = (symbol) => {
  const asset = portfolioStore.portfolio?.find(c => c.symbol === symbol)
  return asset ? asset.amount : 0
}

// 表单验证
const isFormValid = computed(() => {
  return props.selectedSymbol &&
         props.amount &&
         props.amount > 0 &&
         props.price &&
         props.price > 0
})

// 切换资产类型
const switchAssetType = (type) => {
  emit('update:assetType', type)
  emit('update:selectedSymbol', '')
  emit('update:amount', null)
  emit('update:price', null)
  showAssetSelector.value = true
}

// 选择资产
const selectSymbol = async (symbol) => {
  emit('update:selectedSymbol', symbol)
  emit('symbolSelect', symbol)

  // 自动填充价格
  const cachedPrice = getCurrentPrice(symbol)
  if (cachedPrice) {
    emit('update:price', cachedPrice)
  } else {
    const result = await portfolioStore.fetchAssetPrice(symbol, props.assetType)
    if (result.success) {
      emit('update:price', result.price)
    }
  }

  // 移动端：选择资产后自动折叠选择器
  showAssetSelector.value = false

  // 聚焦数量输入框
  nextTick(() => {
    if (amountInputRef.value) {
      amountInputRef.value.focus()
    }
  })
}

// 重新打开资产选择器
const openAssetSelector = () => {
  showAssetSelector.value = true
}

// 快捷设置卖出数量（百分比）
const setQuickAmount = (percent) => {
  const holding = getHoldingAmount(props.selectedSymbol)
  emit('update:amount', Number((holding * percent / 100).toFixed(4)))
}

// 快捷设置买入数量（基于现金余额百分比）
const setQuickBuyAmount = (percent) => {
  if (props.currentMarketPrice > 0) {
    const maxAmount = (props.cashBalance * percent / 100) / props.currentMarketPrice
    emit('update:amount', Number(maxAmount.toFixed(4)))
  }
}

// 提交交易
const handleSubmit = () => {
  emit('submit', {
    assetType: props.assetType,
    symbol: props.selectedSymbol,
    type: props.tradeType,
    amount: props.amount,
    price: props.price
  })
}

// 重置表单
const handleReset = () => {
  showAssetSelector.value = true
  emit('reset')
}

// 使用市价
const useMarketPrice = () => {
  emit('update:price', props.currentMarketPrice)
}

// 暴露给父组件的方法和状态
defineExpose({
  amountInputRef,
  showAssetSelector,
  openAssetSelector,
  availableSymbols,
  getCurrentPrice,
  getHoldingAmount,
  isFormValid,
  switchAssetType,
  selectSymbol,
  setQuickAmount,
  setQuickBuyAmount,
  handleSubmit,
  handleReset,
  useMarketPrice,
  formatAmount,
  formatCompactAmount,
  getAssetColor,
  getAssetIcon
})
</script>

<template>
  <div class="trade-form-base">
    <!-- 通过插槽暴露所有状态和方法给父组件 -->
    <slot
      :trade-type="tradeType"
      :asset-type="assetType"
      :selected-symbol="selectedSymbol"
      :amount="amount"
      :price="price"
      :current-market-price="currentMarketPrice"
      :cash-balance="cashBalance"
      :is-loading="isLoading"
      :is-submitting="isSubmitting"
      :show-asset-selector="showAssetSelector"
      :available-symbols="availableSymbols"
      :is-form-valid="isFormValid"
      :amount-input-ref="amountInputRef"
      :get-current-price="getCurrentPrice"
      :get-holding-amount="getHoldingAmount"
      :switch-asset-type="switchAssetType"
      :select-symbol="selectSymbol"
      :open-asset-selector="openAssetSelector"
      :set-quick-amount="setQuickAmount"
      :set-quick-buy-amount="setQuickBuyAmount"
      :handle-submit="handleSubmit"
      :handle-reset="handleReset"
      :use-market-price="useMarketPrice"
      :format-amount="formatAmount"
      :format-compact-amount="formatCompactAmount"
      :get-asset-color="getAssetColor"
      :get-asset-icon="getAssetIcon"
      :update-trade-type="(v) => $emit('update:tradeType', v)"
      :update-amount="(v) => $emit('update:amount', v)"
      :update-price="(v) => $emit('update:price', v)"
    />
  </div>
</template>

<style scoped>
.trade-form-base {
  display: contents;
}
</style>
