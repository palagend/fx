<template>
  <div class="crypto-container">
    <div class="container">
      <div class="dashboard">
        <main class="main-content">
          <!-- 未登录提示 -->
          <section v-if="!userStore.isLoggedIn" class="login-prompt">
            <div class="prompt-content">
              <Icon icon="mdi:lock" class="prompt-icon" />
              <h3>请先登录</h3>
              <p>登录后可查看和管理您的加密资产组合</p>
              <button class="btn-login" @click="userStore.openLoginModal">
                <Icon icon="mdi:login" /> 立即登录
              </button>
            </div>
          </section>

          <template v-else>
            <section class="overview">
              <div class="overview-card">
                <h3><Icon icon="mdi:wallet" /> 加密资产价值</h3>
                <div class="value">${{ formatAmount(cryptoValue) }}</div>
                <div class="change" :class="displayTotalValueChange.class">
                  {{ displayTotalValueChange.sign }}{{ displayTotalValueChange.value }}% (24h)
                </div>
              </div>
              <div class="overview-card">
                <h3><Icon icon="mdi:trending-up" /> 浮动盈亏</h3>
                <div class="value" :class="displayUnrealizedPL.class">
                  {{ displayUnrealizedPL.sign }}${{ displayUnrealizedPL.value }}
                </div>
                <div class="change" :class="displayUnrealizedPL.class">
                  {{ displayUnrealizedPL.rate }}
                </div>
              </div>
              <div class="overview-card">
                <h3><Icon icon="mdi:cash-multiple" /> 实现盈亏</h3>
                <div class="value" :class="displayRealizedPL.class">
                  {{ displayRealizedPL.sign }}${{ displayRealizedPL.value }}
                </div>
                <div class="change" :class="displayRealizedPL.class">
                  已实现盈亏
                </div>
              </div>
              <div class="overview-card usdt-card">
                <h3><Icon icon="mdi:cash-usd" /> USDT余额</h3>
                <div class="value">${{ formatAmount(usdtBalance) }}</div>
                <button class="btn-recharge" @click="showRechargeModal = true">
                  <Icon icon="mdi:plus" /> 充值
                </button>
              </div>
            </section>

            <section class="chart-section">
              <div class="chart-header">
                <h2 class="chart-title"><Icon icon="mdi:chart-pie" /> 资产分布</h2>
                <div class="chart-actions">
                  <span v-if="lastUpdateTime" class="update-time">
                    更新于: {{ lastUpdateTime }}
                  </span>
                  <button 
                    class="btn-refresh" 
                    @click="refreshPrices" 
                    :disabled="refreshing"
                    :class="{ 'refreshing': refreshing }"
                  >
                    <Icon :icon="refreshing ? 'mdi:loading' : 'mdi:refresh'" />
                    {{ refreshing ? '刷新中...' : '刷新价格' }}
                  </button>
                </div>
              </div>

              <div class="chart-container">
                <div class="chart">
                  <div class="pie-chart-wrapper">
                    <div class="pie-chart" :style="pieChartStyle"></div>
                    <div class="pie-center">
                      <span>${{ formatAmount(totalAssetsValue) }}</span>
                      <span>总资产</span>
                    </div>
                  </div>
                </div>

                <div class="chart-legend">
                  <div
                    v-for="(item, index) in assetAllocation"
                    :key="index"
                    class="legend-item"
                  >
                    <div class="legend-color" :style="{ backgroundColor: item.color }"></div>
                    <span>{{ item.name }} ({{ item.percentage }}%)</span>
                  </div>
                </div>
              </div>
            </section>

            <!-- 交易区域 - 左右分栏布局 -->
            <section class="trading-section">
              <div class="trading-container">
                <!-- 左侧：交易表单 -->
                <div class="trading-form">
                  <div class="form-header">
                    <h3><Icon icon="mdi:swap-horizontal" /> 快速交易</h3>
                    <div class="trade-type-tabs">
                      <button 
                        :class="['type-tab', { active: newTrade.type === 'buy' }]" 
                        @click="newTrade.type = 'buy'"
                      >
                        <Icon icon="mdi:arrow-down-circle" /> 买入
                      </button>
                      <button 
                        :class="['type-tab', { active: newTrade.type === 'sell' }]" 
                        @click="newTrade.type = 'sell'"
                      >
                        <Icon icon="mdi:arrow-up-circle" /> 卖出
                      </button>
                    </div>
                  </div>

                  <!-- 币种选择网格 -->
                  <div class="asset-selector">
                    <label class="field-label">选择资产</label>
                    <div class="asset-grid">
                      <button
                        v-for="symbol in availableSymbols"
                        :key="symbol"
                        :class="['asset-btn', { active: newTrade.symbol === symbol }]"
                        @click="selectSymbol(symbol)"
                      >
                        <Icon :icon="getAssetIcon(symbol)" :style="{ color: getAssetColor(symbol) }" />
                        <span class="asset-code">{{ symbol }}</span>
                        <span class="asset-price" v-if="portfolioStore.prices[symbol]">
                          ${{ formatAmount(portfolioStore.prices[symbol]) }}
                        </span>
                      </button>
                    </div>
                  </div>

                  <!-- 交易输入区 -->
                  <div class="trade-inputs" v-if="newTrade.symbol">
                    <div class="input-field">
                      <label class="field-label">
                        数量
                        <span class="field-hint" v-if="newTrade.type === 'sell'">
                          可卖: {{ formatAmount(getHoldingAmount(newTrade.symbol)) }}
                        </span>
                      </label>
                      <div class="input-with-controls">
                        <input
                          type="number"
                          v-model.number="newTrade.amount"
                          placeholder="0.00"
                          min="0.00001"
                          step="0.00001"
                          ref="amountInput"
                        >
                        <span class="input-unit">{{ newTrade.symbol }}</span>
                      </div>
                    </div>

                    <div class="input-field">
                      <label class="field-label">
                        价格
                        <span class="field-hint" v-if="currentMarketPrice > 0">
                          市价: ${{ formatAmount(currentMarketPrice) }}
                        </span>
                      </label>
                      <div class="input-with-controls">
                        <input
                          type="number"
                          v-model.number="newTrade.price"
                          placeholder="0.00"
                          min="0.00001"
                          step="0.00001"
                        >
                        <button 
                          class="btn-use-market" 
                          @click="newTrade.price = currentMarketPrice"
                          v-if="currentMarketPrice > 0"
                        >
                          使用市价
                        </button>
                      </div>
                    </div>
                  </div>

                  <!-- 交易按钮 -->
                  <div class="trade-submit" v-if="newTrade.symbol">
                    <button 
                      class="btn-submit" 
                      @click="addTrade" 
                      :disabled="!isFormValid || portfolioStore.isLoading || isSubmitting.trade"
                      :class="newTrade.type"
                    >
                      <Icon :icon="newTrade.type === 'buy' ? 'mdi:arrow-down' : 'mdi:arrow-up'" />
                      {{ newTrade.type === 'buy' ? '确认买入' : '确认卖出' }}
                      <span class="submit-total" v-if="newTrade.amount && newTrade.price">
                        ${{ formatAmount(newTrade.amount * newTrade.price) }}
                      </span>
                    </button>
                    <button class="btn-reset" @click="clearForm">
                      <Icon icon="mdi:close" /> 重置
                    </button>
                  </div>
                </div>

                <!-- 右侧：交易预览 -->
                <div class="trading-preview" v-if="newTrade.symbol && newTrade.amount && newTrade.price">
                  <div class="preview-header">
                    <h4><Icon icon="mdi:eye" /> 交易预览</h4>
                  </div>
                  
                  <div class="preview-content">
                    <!-- 主要信息 -->
                    <div class="preview-main">
                      <div class="preview-item total">
                        <span class="label">交易总额</span>
                        <span class="value">${{ formatAmount(newTrade.amount * newTrade.price) }}</span>
                      </div>
                    </div>

                    <!-- 买入预览 -->
                    <template v-if="newTrade.type === 'buy'">
                      <div class="preview-divider"></div>
                      <div class="preview-details">
                        <div class="preview-item" v-if="getHoldingAmount(newTrade.symbol) > 0">
                          <span class="label">当前持仓</span>
                          <span class="value">{{ formatAmount(getHoldingAmount(newTrade.symbol)) }}</span>
                        </div>
                        <div class="preview-item" v-if="getHoldingAmount(newTrade.symbol) > 0">
                          <span class="label">买入后持仓</span>
                          <span class="value highlight">{{ formatAmount(getHoldingAmount(newTrade.symbol) + newTrade.amount) }}</span>
                        </div>
                        <div class="preview-item" v-if="getHoldingAmount(newTrade.symbol) > 0">
                          <span class="label">新综合成本</span>
                          <span class="value">${{ formatAmount(calculateNewAvgCost()) }}</span>
                        </div>
                        <div class="preview-item impact">
                          <span class="label">USDT支出</span>
                          <span class="value negative">-${{ formatAmount(newTrade.amount * newTrade.price) }}</span>
                        </div>
                      </div>
                    </template>

                    <!-- 卖出预览 -->
                    <template v-if="newTrade.type === 'sell'">
                      <div class="preview-divider"></div>
                      <div class="preview-details">
                        <div class="preview-item">
                          <span class="label">当前持仓</span>
                          <span class="value">{{ formatAmount(getHoldingAmount(newTrade.symbol)) }}</span>
                        </div>
                        <div class="preview-item">
                          <span class="label">当前成本价</span>
                          <span class="value">${{ formatAmount(portfolio.value.find(c => c.symbol === newTrade.symbol)?.avg_cost || 0) }}</span>
                        </div>
                        <div class="preview-item">
                          <span class="label">卖出后持仓</span>
                          <span class="value">{{ formatAmount(Math.max(0, getHoldingAmount(newTrade.symbol) - newTrade.amount)) }}</span>
                        </div>
                        <div class="preview-item" v-if="newTrade.amount < getHoldingAmount(newTrade.symbol)">
                          <span class="label">卖出后成本价</span>
                          <span class="value highlight">${{ formatAmount(calculateSellAvgCost()) }}</span>
                        </div>
                        <div class="preview-item" v-if="calculateEstimatedRealizedPL() !== 0">
                          <span class="label">预估盈亏</span>
                          <span :class="['value', calculateEstimatedRealizedPL() >= 0 ? 'positive' : 'negative']">
                            {{ calculateEstimatedRealizedPL() >= 0 ? '+' : '-' }}${{ formatAmount(Math.abs(calculateEstimatedRealizedPL())) }}
                          </span>
                        </div>
                        <div class="preview-item impact">
                          <span class="label">USDT收入</span>
                          <span class="value positive">+${{ formatAmount(newTrade.amount * newTrade.price) }}</span>
                        </div>
                      </div>
                    </template>
                  </div>
                </div>

                <!-- 空状态提示 -->
                <div class="trading-preview empty" v-else-if="newTrade.symbol">
                  <div class="preview-placeholder">
                    <Icon icon="mdi:calculator" />
                    <p>输入数量和价格查看交易详情</p>
                  </div>
                </div>
              </div>
            </section>

            <section class="portfolio-section">
              <div class="section-header">
                <h2 class="section-title"><Icon icon="mdi:wallet-outline" /> 资产详情</h2>
                <div class="filter-group">
                  <select v-model="selectedFilter" class="filter-select">
                    <option value="all">全部资产</option>
                    <option v-for="crypto in portfolio" :key="crypto.symbol" :value="crypto.symbol">
                      {{ crypto.symbol }}
                    </option>
                  </select>
                </div>
              </div>

              <div class="table-wrapper">
                <table class="portfolio-table">
                  <thead>
                    <tr>
                      <th>资产</th>
                      <th>持有量</th>
                      <th>成本价</th>
                      <th>当前价</th>
                      <th>总价值</th>
                      <th>浮动盈亏</th>
                      <th>操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="crypto in filteredPortfolio"
                      :key="crypto.id"
                      class="asset-row"
                      :class="{ 'selected': selectedAsset === crypto.symbol }"
                      @click="selectAsset(crypto.symbol)"
                    >
                      <td>
                        <div class="asset-info">
                          <Icon width="32" height="32" :icon="getAssetIcon(crypto.symbol)" :style="{ color: getAssetColor(crypto.symbol) }" />
                          <div>
                            <div class="asset-name">{{ getAssetName(crypto.symbol) }}</div>
                            <div class="asset-symbol">{{ crypto.symbol }}</div>
                          </div>
                        </div>
                      </td>
                      <td class="asset-amount">{{ formatAmount(crypto.amount) }}</td>
                      <td class="asset-price">
                        <template v-if="crypto.avg_cost > 0">${{ formatAmount(crypto.avg_cost) }}</template>
                        <template v-else-if="crypto.avg_cost === 0">
                          <div class="cost-display">
                            <span class="cost-value">$0</span>
                            <span class="cost-badge zero">已回本</span>
                          </div>
                        </template>
                        <template v-else>
                          <div class="cost-display">
                            <span class="cost-value">${{ formatAmount(crypto.avg_cost) }}</span>
                            <span class="cost-badge negative">负成本</span>
                          </div>
                        </template>
                      </td>
                      <td class="asset-price">${{ formatAmount(crypto.current_price) }}</td>
                      <td class="asset-value">${{ formatAmount(crypto.market_value) }}</td>
                      <td class="asset-profit" :class="getProfitClass(crypto)">
                        <!-- avg_cost > 0: 正常情况 -->
                        <template v-if="crypto.avg_cost > 0">
                          <div class="profit-value">
                            {{ (crypto.amount * (crypto.current_price - crypto.avg_cost)) >= 0 ? '+' : '-' }}${{ formatAmount(Math.abs(crypto.amount * (crypto.current_price - crypto.avg_cost))) }}
                          </div>
                          <div class="profit-rate" v-if="crypto.symbol !== 'USDT'">
                            {{ ((crypto.current_price - crypto.avg_cost) / crypto.avg_cost * 100) >= 0 ? '+' : '-' }}{{ Math.abs((crypto.current_price - crypto.avg_cost) / crypto.avg_cost * 100).toFixed(2) }}%
                          </div>
                        </template>
                        
                        <!-- avg_cost = 0: 投资全部收回 -->
                        <template v-else-if="crypto.avg_cost === 0">
                          <div class="profit-value positive">
                            +${{ formatAmount(crypto.amount * crypto.current_price) }}
                          </div>
                          <div class="profit-rate" v-if="crypto.symbol !== 'USDT'">
                            <span class="status-badge recovered">✓ 已回本</span>
                          </div>
                        </template>
                        
                        <!-- avg_cost < 0: 投资回报超过100% -->
                        <template v-else>
                          <div class="profit-value positive">
                            +${{ formatAmount(crypto.amount * crypto.current_price - crypto.cost) }}
                          </div>
                          <div class="profit-rate" v-if="crypto.symbol !== 'USDT'">
                            <span class="status-badge super-profit">🚀 超100%回报</span>
                          </div>
                        </template>
                      </td>
                      <td class="action-cell">
                        <button class="btn-action btn-sell" @click.stop="quickSell(crypto)" title="快速卖出">
                          <Icon icon="mdi:shopping-cart-arrow-up" />
                        </button>
                        <button class="btn-action btn-buy" @click.stop="quickBuy(crypto)" title="快速买入">
                          <Icon icon="mdi:shopping-cart-arrow-down" />
                        </button>
                      </td>
                    </tr>
                    <tr v-if="filteredPortfolio.length === 0">
                      <td colspan="7" class="empty-state">
                        <Icon icon="mdi:inbox" />
                        <p>暂无资产数据，请充值USDT后开始交易</p>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </section>

            <section class="trades-section">
              <div class="section-header">
                <h2 class="section-title"><Icon icon="mdi:history" /> 交易历史</h2>
                <div class="section-actions">
                  <div class="protect-switch" @click="toggleProtectHistory">
                    <Icon :icon="protectHistory ? 'mdi:shield-check' : 'mdi:shield-off'" />
                    <span class="switch-label">保护</span>
                    <div class="switch" :class="{ 'on': protectHistory }">
                      <div class="switch-handle"></div>
                    </div>
                  </div>
                  <div class="filter-group">
                    <select v-model="tradeFilter" class="filter-select">
                      <option value="all">全部交易</option>
                      <option value="buy">买入</option>
                      <option value="sell">卖出</option>
                      <option value="recharge">充值</option>
                    </select>
                  </div>
                  <button class="btn-clear" @click="clearTrades" v-if="filteredTrades.length > 0 && !protectHistory" :disabled="isSubmitting.clear">
                    <Icon icon="mdi:delete-sweep" /> 清空历史
                  </button>
                </div>
              </div>

              <div class="table-wrapper">
                <table class="trades-table">
                  <thead>
                    <tr>
                      <th>时间</th>
                      <th>资产</th>
                      <th>类型</th>
                      <th>数量</th>
                      <th>价格</th>
                      <th>总金额</th>
                      <th>操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="trade in filteredTrades" :key="trade.id" class="trade-row">
                      <td class="trade-time">{{ formatDate(trade.created_at || trade.timestamp) }}</td>
                      <td>
                        <div class="trade-asset">
                          <Icon :icon="getAssetIcon(trade.symbol)" :style="{ color: getAssetColor(trade.symbol) }" />
                          <span>{{ trade.symbol }}</span>
                        </div>
                      </td>
                      <td>
                        <span class="trade-type" :class="trade.type">
                          {{ getTradeTypeText(trade.type) }}
                        </span>
                      </td>
                      <td>{{ formatAmount(trade.amount) }}</td>
                      <td>${{ formatAmount(trade.price) }}</td>
                      <td>${{ formatAmount(trade.total) }}</td>
                      <td>
                        <button class="btn-delete" @click="deleteTrade(trade.id)" :disabled="protectHistory || isSubmitting.delete" title="删除">
                          <Icon icon="mdi:trash-can" />
                        </button>
                      </td>
                    </tr>
                    <tr v-if="filteredTrades.length === 0">
                      <td colspan="7" class="empty-state">
                        <Icon icon="mdi:inbox" />
                        <p>暂无交易记录</p>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </section>
          </template>
        </main>
      </div>
    </div>

    <!-- 充值弹窗 -->
    <div v-if="showRechargeModal" class="modal-overlay" @click.self="showRechargeModal = false">
      <div class="modal">
        <div class="modal-header">
          <h3><Icon icon="mdi:cash-plus" /> 充值USDT</h3>
          <button class="btn-close" @click="showRechargeModal = false">
            <Icon icon="mdi:close" />
          </button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>充值金额 (USDT)</label>
            <input
              type="number"
              v-model.number="rechargeAmount"
              placeholder="输入充值金额"
              min="0.01"
              step="0.01"
              @keyup.enter="rechargeUSDT"
            >
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-cancel" @click="showRechargeModal = false">取消</button>
          <button class="btn-confirm" @click="rechargeUSDT" :disabled="!rechargeAmount || rechargeAmount <= 0 || isSubmitting.recharge">
            确认充值
          </button>
        </div>
      </div>
    </div>

    <!-- 错误提示 -->
    <div v-if="errorMessage" class="error-toast">
      <Icon icon="mdi:alert-circle" />
      <span>{{ errorMessage }}</span>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { Icon } from '@iconify/vue'
import { usePortfolioStore } from '../stores/portfolio'
import { useUserStore } from '../stores/user'

const portfolioStore = usePortfolioStore()
const userStore = useUserStore()

// 支持的加密货币列表
const availableSymbols = ['BTC', 'ETH', 'BNB', 'XRP', 'ADA', 'SOL', 'DOGE', 'TRX', 'AVAX', 'HYPE']

// 状态
const newTrade = ref({
  symbol: '',
  type: 'buy',
  amount: null,
  price: null
})
const currentMarketPrice = ref(0) // 当前选中的币种市价
const tradeFilter = ref('all')
const refreshing = ref(false)
const lastUpdateTime = ref('')
const errorMessage = ref('')
const autoRefresh = ref(false)
const refreshInterval = ref(60)
const selectedFilter = ref('all')
const selectedAsset = ref(null)
const symbolSelect = ref(null)
const amountInput = ref(null)
const showRechargeModal = ref(false)
const rechargeAmount = ref(null)
const protectHistory = ref(false)
let refreshTimer = null

// 重复提交保护状态
const isSubmitting = ref({
  trade: false,
  recharge: false,
  delete: false,
  clear: false
})

// 从store获取数据（后端已计算好）
const portfolio = computed(() => portfolioStore.portfolio)
const trades = computed(() => portfolioStore.trades)
const cryptoValue = computed(() => portfolioStore.cryptoValue) // 加密资产市值（不含 USDT）
const totalAssetsValue = computed(() => portfolioStore.totalAssetsValue) // 总资产价值（含 USDT）
const usdtBalance = computed(() => portfolioStore.usdtBalance)
const unrealizedProfitLoss = computed(() => portfolioStore.unrealizedProfitLoss)
const unrealizedProfitLossRate = computed(() => portfolioStore.unrealizedProfitLossRate)
const realizedProfitLoss = computed(() => portfolioStore.realizedProfitLoss)
const valueChange24h = computed(() => portfolioStore.valueChange24h) // 24小时价值变化率

// 格式化的显示值（避免模板中重复计算）
const displayUnrealizedPL = computed(() => {
  const val = unrealizedProfitLoss.value
  return {
    sign: val >= 0 ? '+' : '-',
    value: formatAmount(Math.abs(val)),
    class: val >= 0 ? 'positive' : 'negative',
    rate: (unrealizedProfitLossRate.value >= 0 ? '+' : '-') + Math.abs(unrealizedProfitLossRate.value).toFixed(2) + '%'
  }
})

const displayRealizedPL = computed(() => {
  const val = realizedProfitLoss.value
  return {
    sign: val >= 0 ? '+' : '-',
    value: formatAmount(Math.abs(val)),
    class: val >= 0 ? 'positive' : 'negative'
  }
})

const displayTotalValueChange = computed(() => {
  const val = valueChange24h.value
  return {
    sign: val >= 0 ? '+' : '-',
    value: Math.abs(val).toFixed(2),
    class: getChangeClass(val)
  }
})

const ASSET_CONFIG = {
  COLORS: {
    USDT: '#26a17b',
    BTC: '#f7931a',
    ETH: '#627eea',
    BNB: '#f0b90b',
    XRP: '#0033ad',
    ADA: '#0033ad',
    SOL: '#00ffa3',
    DOGE: '#c2a633',
    TRX: '#eb0029',
    AVAX: '#e84142',
    HYPE: '#89F0E6'
  },
  ICONS: {
    USDT: 'cryptocurrency-color:usdt',
    BTC: 'cryptocurrency-color:btc',
    ETH: 'cryptocurrency-color:eth',
    BNB: 'cryptocurrency-color:bnb',
    XRP: 'cryptocurrency-color:xrp',
    ADA: 'cryptocurrency-color:ada',
    SOL: 'cryptocurrency-color:sol',
    DOGE: 'cryptocurrency-color:doge',
    TRX: 'cryptocurrency-color:trx',
    AVAX: 'cryptocurrency-color:avax',
    HYPE: 'token:hyper-evm'
  },
  NAMES: {
    USDT: 'Tether',
    BTC: 'Bitcoin',
    ETH: 'Ethereum',
    BNB: 'Binance Coin',
    XRP: 'Ripple',
    ADA: 'Cardano',
    SOL: 'Solana',
    DOGE: 'Dogecoin',
    TRX: 'Tron',
    AVAX: 'Avalanche',
    HYPE: 'Hyperliquid'
  }
}

const CHART_COLORS = [
  '#6366f1', '#8b5cf6', '#d946ef', '#ec4899', '#f43f5e',
  '#fb7185', '#fda4af', '#fca5a5', '#f87171', '#fb923c'
]

const getTradeTypeText = (type) => {
  const map = {
    buy: '买入',
    sell: '卖出',
    recharge: '充值'
  }
  return map[type] || type
}

const toggleProtectHistory = () => {
  protectHistory.value = !protectHistory.value
}

const selectSymbol = async (symbol) => {
  newTrade.value.symbol = symbol
  // 使用 GetAssetPrice 接口获取最新价格
  const result = await portfolioStore.fetchAssetPrice(symbol)
  if (result.success) {
    newTrade.value.price = result.price
    currentMarketPrice.value = result.price
  }
  nextTick(() => {
    if (amountInput.value) {
      amountInput.value.focus()
    }
  })
}

const onSymbolChange = async () => {
  if (newTrade.value.symbol) {
    // 使用 GetAssetPrice 接口获取最新价格
    const result = await portfolioStore.fetchAssetPrice(newTrade.value.symbol)
    if (result.success) {
      newTrade.value.price = result.price
      currentMarketPrice.value = result.price
    }
  } else {
    currentMarketPrice.value = 0
  }
  nextTick(() => {
    if (amountInput.value) {
      amountInput.value.focus()
    }
  })
}

const isFormValid = computed(() => {
  return newTrade.value.symbol && 
         newTrade.value.amount && 
         newTrade.value.amount > 0 && 
         newTrade.value.price && 
         newTrade.value.price > 0
})

const getHoldingAmount = (symbol) => {
  const asset = portfolio.value.find(c => c.symbol === symbol)
  return asset ? asset.amount : 0
}

// 根据avg_cost获取盈亏显示的CSS类
const getProfitClass = (crypto) => {
  if (crypto.avg_cost > 0) {
    // 正常情况：根据盈亏判断
    const profit = crypto.amount * (crypto.current_price - crypto.avg_cost)
    return { positive: profit >= 0, negative: profit < 0 }
  } else if (crypto.avg_cost === 0) {
    // 已回本：总是显示为盈利（绿色）
    return { positive: true }
  } else {
    // 超100%回报：总是显示为盈利（绿色）
    return { positive: true }
  }
}

// 计算卖出时的预估实现盈亏（借贷记账法）
// 实现盈亏 = USDT退出 - USDT投入（按卖出比例计算）
const calculateEstimatedRealizedPL = () => {
  if (newTrade.value.type !== 'sell') return 0
  const existing = portfolio.value.find(c => c.symbol === newTrade.value.symbol)
  if (!existing || existing.amount === 0) return 0

  // 本次卖出获得的USDT
  const usdtOut = newTrade.value.price * newTrade.value.amount

  // 按卖出比例计算的USDT投入成本
  const costRatio = newTrade.value.amount / existing.amount
  const usdtIn = existing.cost * costRatio

  return usdtOut - usdtIn
}

// 计算买入后的新综合成本价（借贷记账法）
// 新成本价 = (USDT净投入 + 本次投入) / (当前持仓 + 本次买入量)
const calculateNewAvgCost = () => {
  if (newTrade.value.type !== 'buy' || !newTrade.value.symbol) return 0
  const existing = portfolio.value.find(c => c.symbol === newTrade.value.symbol)
  const currentAmount = existing ? existing.amount : 0
  const currentCost = existing ? existing.cost : 0
  const newAmount = newTrade.value.amount
  const newTotal = newTrade.value.amount * newTrade.value.price

  if (currentAmount === 0) return newTrade.value.price

  const totalCost = currentCost + newTotal
  const totalAmount = currentAmount + newAmount
  return totalCost / totalAmount
}

// 计算卖出后的新综合成本价
// 卖出后成本价 = (当前成本 - 卖出比例对应的成本) / (当前持仓 - 卖出数量)
const calculateSellAvgCost = () => {
  if (newTrade.value.type !== 'sell' || !newTrade.value.symbol) return 0
  const existing = portfolio.value.find(c => c.symbol === newTrade.value.symbol)
  if (!existing || existing.amount === 0) return 0

  const currentAmount = existing.amount
  const currentCost = existing.cost
  const sellAmount = newTrade.value.amount

  // 如果全部卖出，成本价为0
  if (sellAmount >= currentAmount) return 0

  // 按卖出比例计算剩余成本
  const remainingAmount = currentAmount - sellAmount
  const sellRatio = sellAmount / currentAmount
  const remainingCost = currentCost * (1 - sellRatio)

  return remainingCost / remainingAmount
}

const addTrade = async () => {
  if (isSubmitting.value.trade) return

  if (!newTrade.value.symbol) {
    errorMessage.value = '请选择加密货币'
    setTimeout(() => errorMessage.value = '', 3000)
    return
  }

  if (newTrade.value.amount <= 0) {
    errorMessage.value = '请输入大于 0 的数量'
    setTimeout(() => errorMessage.value = '', 3000)
    return
  }

  if (newTrade.value.price <= 0) {
    errorMessage.value = '请输入大于 0 的价格'
    setTimeout(() => errorMessage.value = '', 3000)
    return
  }

  isSubmitting.value.trade = true

  try {
    const result = await portfolioStore.createTrade({
      symbol: newTrade.value.symbol,
      type: newTrade.value.type,
      amount: newTrade.value.amount,
      price: newTrade.value.price
    })

    if (!result.success) {
      errorMessage.value = result.error
      setTimeout(() => errorMessage.value = '', 3000)
      return
    }

    refreshPrices()

    newTrade.value = {
      symbol: '',
      type: 'buy',
      amount: null,
      price: null
    }
  } finally {
    isSubmitting.value.trade = false
  }
}

const rechargeUSDT = async () => {
  if (isSubmitting.value.recharge) return

  if (!rechargeAmount.value || rechargeAmount.value <= 0) {
    errorMessage.value = '请输入有效的充值金额'
    setTimeout(() => errorMessage.value = '', 3000)
    return
  }

  isSubmitting.value.recharge = true

  try {
    const result = await portfolioStore.createTrade({
      symbol: 'USDT',
      type: 'recharge',
      amount: rechargeAmount.value,
      price: 1
    })

    if (!result.success) {
      errorMessage.value = result.error
      setTimeout(() => errorMessage.value = '', 3000)
      return
    }

    rechargeAmount.value = null
    showRechargeModal.value = false
  } finally {
    isSubmitting.value.recharge = false
  }
}

const deleteTrade = async (id) => {
  if (isSubmitting.value.delete) return

  if (protectHistory.value) {
    errorMessage.value = '保护开关已开启，禁止删除交易历史'
    setTimeout(() => errorMessage.value = '', 3000)
    return
  }
  if (!confirm('确认删除该交易？该操作将同步更新资产详情中的持仓量和成本价。')) {
    return
  }

  isSubmitting.value.delete = true

  try {
    const result = await portfolioStore.deleteTrade(id)
    if (!result.success) {
      errorMessage.value = result.error
      setTimeout(() => errorMessage.value = '', 3000)
    }
  } finally {
    isSubmitting.value.delete = false
  }
}

const clearTrades = async () => {
  if (isSubmitting.value.clear) return

  if (protectHistory.value) {
    errorMessage.value = '保护开关已开启，禁止删除交易历史'
    setTimeout(() => errorMessage.value = '', 3000)
    return
  }
  if (!confirm('确认清空所有交易历史？此操作将重置所有数据。')) {
    return
  }

  isSubmitting.value.clear = true

  try {
    const result = await portfolioStore.clearAllTrades()
    if (!result.success) {
      errorMessage.value = result.error
      setTimeout(() => errorMessage.value = '', 3000)
    }
  } finally {
    isSubmitting.value.clear = false
  }
}

const clearForm = () => {
  newTrade.value = {
    symbol: '',
    type: 'buy',
    amount: null,
    price: null
  }
  nextTick(() => {
    if (symbolSelect.value) {
      symbolSelect.value.focus()
    }
  })
}

const selectAsset = (symbol) => {
  // toggle 模式：点击已选中的资产则取消过滤，显示全部
  if (selectedAsset.value === symbol) {
    selectedAsset.value = null
    selectedFilter.value = 'all'
  } else {
    selectedAsset.value = symbol
    selectedFilter.value = symbol
  }
}

const quickSell = async (crypto) => {
  newTrade.value.symbol = crypto.symbol
  newTrade.value.type = 'sell'
  newTrade.value.amount = crypto.amount
  // 使用 GetAssetPrice 接口获取最新价格
  const result = await portfolioStore.fetchAssetPrice(crypto.symbol)
  if (result.success) {
    newTrade.value.price = result.price
    currentMarketPrice.value = result.price
  } else {
    newTrade.value.price = crypto.currentPrice || (crypto.avg_cost || crypto.price)
  }
}

const quickBuy = async (crypto) => {
  newTrade.value.symbol = crypto.symbol
  newTrade.value.type = 'buy'
  // 使用 GetAssetPrice 接口获取最新价格
  const result = await portfolioStore.fetchAssetPrice(crypto.symbol)
  if (result.success) {
    newTrade.value.price = result.price
    currentMarketPrice.value = result.price
  } else {
    newTrade.value.price = crypto.currentPrice || (crypto.avg_cost || crypto.price)
  }
  nextTick(() => {
    if (amountInput.value) {
      amountInput.value.focus()
    }
  })
}

const getAssetName = (symbol) => {
  return ASSET_CONFIG.NAMES[symbol] || symbol
}

const getAssetColor = (symbol) => {
  return ASSET_CONFIG.COLORS[symbol] || '#667eea'
}

const getAssetIcon = (symbol) => {
  return ASSET_CONFIG.ICONS[symbol] || symbol.charAt(0)
}

const formatAmount = (amount) => {
  if (!amount && amount !== 0) return '0.0000'
  return amount.toLocaleString('en-US', {
    minimumFractionDigits: 4,
    maximumFractionDigits: 4
  })
}

// 缓存日期格式化器，避免重复创建
const dateFormatter = new Intl.DateTimeFormat('zh-CN', {
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
  hour: '2-digit',
  minute: '2-digit',
  second: '2-digit'
})

const formatDate = (timestamp) => {
  if (!timestamp) return '-'
  const date = new Date(timestamp)
  return dateFormatter.format(date)
}

const getChangeClass = (change) => {
  if (change > 0) return 'positive'
  if (change < 0) return 'negative'
  return ''
}

// 防抖计时器
let refreshDebounceTimer = null
const REFRESH_DEBOUNCE_MS = 5000 // 5秒内禁止重复刷新

const refreshPrices = async () => {
  if (refreshing.value) return
  
  // 防抖检查
  if (refreshDebounceTimer) {
    errorMessage.value = '刷新过于频繁，请稍后再试'
    setTimeout(() => errorMessage.value = '', 2000)
    return
  }
  
  // 设置防抖计时器
  refreshDebounceTimer = setTimeout(() => {
    refreshDebounceTimer = null
  }, REFRESH_DEBOUNCE_MS)
  
  refreshing.value = true
  errorMessage.value = ''

  try {
    const result = await portfolioStore.fetchDashboard()

    if (result.success) {
      lastUpdateTime.value = formatDate(result.updatedAt)
    } else {
      errorMessage.value = result.error || '获取价格失败'
    }
  } catch (error) {
    console.error('Failed to fetch prices:', error)
    errorMessage.value = '获取价格失败，请检查网络连接'
  }

  refreshing.value = false
}

const toggleAutoRefresh = () => {
  if (autoRefresh.value) {
    refreshTimer = setInterval(() => {
      refreshPrices()
    }, refreshInterval.value * 60 * 1000)
  } else {
    if (refreshTimer) {
      clearInterval(refreshTimer)
      refreshTimer = null
    }
  }
}

const filteredPortfolio = computed(() => {
  const filter = selectedFilter.value
  return portfolio.value.filter(c => 
    c.symbol !== 'USDT' && (filter === 'all' || c.symbol === filter)
  )
})

const filteredTrades = computed(() => {
  const filter = tradeFilter.value
  return filter === 'all' ? trades.value : trades.value.filter(t => t.type === filter)
})

const assetAllocation = computed(() => {
  const allHoldings = portfolio.value
  if (allHoldings.length === 0) return []

  // 使用总资产价值（含 USDT）计算分布
  const total = totalAssetsValue.value
  if (total <= 0) return []

  return allHoldings
    .map((crypto, index) => ({
      name: crypto.symbol,
      percentage: parseFloat(((crypto.market_value || 0) / total * 100).toFixed(1)),
      value: crypto.market_value || 0,
      color: ASSET_CONFIG.COLORS[crypto.symbol] || CHART_COLORS[index % CHART_COLORS.length]
    }))
    .filter(item => item.value > 0)
    .sort((a, b) => b.value - a.value)
})

const pieChartStyle = computed(() => {
  const allocation = assetAllocation.value
  if (allocation.length === 0) return {}

  const gradient = allocation.reduce((acc, item, index) => {
    const startAngle = index === 0 ? 0 : acc.angle
    const endAngle = startAngle + (item.percentage * 3.6)
    const separator = index < allocation.length - 1 ? ', ' : ''
    return {
      str: acc.str + `${item.color} ${startAngle}deg ${endAngle}deg${separator}`,
      angle: endAngle
    }
  }, { str: 'conic-gradient(', angle: 0 })

  return { background: gradient.str + ')' }
})

onMounted(() => {
  if (userStore.isLoggedIn) {
    portfolioStore.fetchDashboard()
  }
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})

// 监听登录状态变化
watch(() => userStore.isLoggedIn, async (isLoggedIn) => {
  if (isLoggedIn) {
    await portfolioStore.fetchDashboard()
  }
})
</script>

<style scoped>
.crypto-container {
  min-height: calc(100vh - 120px);
}

.container {
  max-width: 1400px;
  margin: 0 auto;
}

.dashboard {
  display: block;
}

/* 登录提示 */
.login-prompt {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
  padding: 40px;
}

.prompt-content {
  text-align: center;
  background: white;
  padding: 60px;
  border-radius: 16px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

.dark .prompt-content {
  background: #1e1e1e;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
}

.prompt-icon {
  font-size: 64px;
  color: #6366f1;
  margin-bottom: 20px;
}

.prompt-content h3 {
  font-size: 24px;
  margin-bottom: 12px;
  color: #1f2937;
}

.dark .prompt-content h3 {
  color: #f3f4f6;
}

.prompt-content p {
  color: #6b7280;
  margin-bottom: 24px;
}

.dark .prompt-content p {
  color: #9ca3af;
}

.btn-login {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 32px;
  background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-login:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.4);
}

.overview {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 20px;
  margin-bottom: 30px;
}

.overview-card {
  background-color: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  transition: all 0.3s ease;
}

.dark .overview-card {
  background-color: #1e1e1e;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.overview-card:hover {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.dark .overview-card:hover {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
}

.overview-card h3 {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #6b7280;
  margin-bottom: 12px;
  font-weight: 500;
}

.dark .overview-card h3 {
  color: #9ca3af;
}

.overview-card .value {
  font-size: 28px;
  font-weight: 700;
  color: #1f2937;
  margin-bottom: 8px;
}

.dark .overview-card .value {
  color: #f3f4f6;
}

.overview-card .change {
  font-size: 14px;
  font-weight: 500;
}

.overview-card .positive {
  color: #10b981;
}

.overview-card .negative {
  color: #ef4444;
}

.usdt-card {
  position: relative;
}

.btn-recharge {
  position: absolute;
  top: 24px;
  right: 24px;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-recharge:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.4);
}

.chart-section {
  background-color: white;
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 30px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.dark .chart-section {
  background-color: #1e1e1e;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.chart-header {
  margin-bottom: 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.chart-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
}

.dark .chart-title {
  color: #f3f4f6;
}

.chart-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.update-time {
  font-size: 12px;
  color: #6b7280;
}

.dark .update-time {
  color: #9ca3af;
}

.btn-refresh {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-refresh:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.4);
}

.btn-refresh:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.btn-refresh.refreshing svg {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.chart-container {
  display: flex;
  align-items: center;
  gap: 40px;
  flex-wrap: wrap;
}

.chart {
  flex: 0 0 300px;
}

.pie-chart-wrapper {
  position: relative;
  width: 250px;
  height: 250px;
}

.pie-chart {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  transition: all 0.3s ease;
}

.pie-center {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
  background: white;
  border-radius: 50%;
  width: 140px;
  height: 140px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.dark .pie-center {
  background: #1e1e1e;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.pie-center span:first-child {
  font-size: 20px;
  font-weight: 700;
  color: #1f2937;
}

.dark .pie-center span:first-child {
  color: #f3f4f6;
}

.pie-center span:last-child {
  font-size: 12px;
  color: #6b7280;
  margin-top: 4px;
}

.dark .pie-center span:last-child {
  color: #9ca3af;
}

.chart-legend {
  flex: 1;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 12px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #4b5563;
}

.dark .legend-item {
  color: #d1d5db;
}

.legend-color {
  width: 12px;
  height: 12px;
  border-radius: 3px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 16px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
}

.dark .section-title {
  color: #f3f4f6;
}

.section-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  align-items: center;
}

.input-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

/* ==================== 交易区域新样式 ==================== */

/* 交易区域容器 */
.trading-section {
  background-color: white;
  border-radius: 16px;
  padding: 24px;
  margin-bottom: 30px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

.dark .trading-section {
  background-color: #1e1e1e;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
}

/* 左右分栏布局 */
.trading-container {
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: 24px;
}

@media (max-width: 900px) {
  .trading-container {
    grid-template-columns: 1fr;
  }
}

/* 左侧表单区域 */
.trading-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* 表单头部 */
.form-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}

.form-header h3 {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
  margin: 0;
}

.dark .form-header h3 {
  color: #f3f4f6;
}

/* 买入/卖出切换标签 */
.trade-type-tabs {
  display: flex;
  gap: 8px;
  background: #f3f4f6;
  padding: 4px;
  border-radius: 10px;
}

.dark .trade-type-tabs {
  background: #2d2d2d;
}

.type-tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  background: transparent;
  color: #6b7280;
}

.dark .type-tab {
  color: #9ca3af;
}

.type-tab:hover {
  color: #374151;
}

.type-tab.active {
  background: white;
  color: #1f2937;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.dark .type-tab.active {
  background: #404040;
  color: #f3f4f6;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

/* 字段标签 */
.field-label {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
  font-weight: 500;
  color: #6b7280;
  margin-bottom: 8px;
}

.dark .field-label {
  color: #9ca3af;
}

.field-hint {
  font-size: 12px;
  color: #6366f1;
  font-weight: 400;
}

/* 币种选择网格 */
.asset-selector {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.asset-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
  gap: 8px;
}

.asset-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 12px 8px;
  background: #f9fafb;
  border: 2px solid transparent;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s ease;
  color: #374151;
}

.dark .asset-btn {
  background: #2d2d2d;
  color: #e5e7eb;
}

.asset-btn:hover {
  background: #f3f4f6;
  border-color: #e5e7eb;
  transform: translateY(-2px);
}

.dark .asset-btn:hover {
  background: #404040;
  border-color: #525252;
}

.asset-btn.active {
  background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
  border-color: transparent;
  color: white;
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.3);
}

.asset-btn svg {
  font-size: 24px;
}

.asset-code {
  font-size: 14px;
  font-weight: 600;
}

.asset-price {
  font-size: 11px;
  opacity: 0.8;
}

.asset-btn.active .asset-price {
  color: rgba(255, 255, 255, 0.9);
}

/* 交易输入区 */
.trade-inputs {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.input-field {
  display: flex;
  flex-direction: column;
}

.input-with-controls {
  display: flex;
  align-items: center;
  gap: 8px;
  position: relative;
}

.input-with-controls input {
  flex: 1;
  padding: 12px 16px;
  padding-right: 60px;
  border: 2px solid #e5e7eb;
  border-radius: 10px;
  font-size: 16px;
  font-weight: 500;
  background: white;
  color: #1f2937;
  transition: all 0.3s ease;
}

.dark .input-with-controls input {
  background: #2d2d2d;
  border-color: #404040;
  color: #f3f4f6;
}

.input-with-controls input:focus {
  outline: none;
  border-color: #6366f1;
  box-shadow: 0 0 0 4px rgba(99, 102, 241, 0.1);
}

.input-unit {
  position: absolute;
  right: 16px;
  font-size: 13px;
  font-weight: 500;
  color: #9ca3af;
  pointer-events: none;
}

.btn-use-market {
  padding: 8px 12px;
  background: #e0e7ff;
  border: none;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  color: #6366f1;
  cursor: pointer;
  transition: all 0.3s ease;
  white-space: nowrap;
}

.dark .btn-use-market {
  background: #3730a3;
  color: #a5b4fc;
}

.btn-use-market:hover {
  background: #c7d2fe;
}

.dark .btn-use-market:hover {
  background: #4338ca;
}

/* 交易提交按钮 */
.trade-submit {
  display: flex;
  gap: 12px;
  margin-top: 8px;
}

.btn-submit {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 14px 24px;
  border: none;
  border-radius: 12px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-submit.buy {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  box-shadow: 0 4px 14px rgba(16, 185, 129, 0.3);
}

.btn-submit.buy:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(16, 185, 129, 0.4);
}

.btn-submit.sell {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
  box-shadow: 0 4px 14px rgba(239, 68, 68, 0.3);
}

.btn-submit.sell:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(239, 68, 68, 0.4);
}

.btn-submit:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.submit-total {
  margin-left: 8px;
  padding-left: 12px;
  border-left: 1px solid rgba(255, 255, 255, 0.3);
  font-size: 14px;
  opacity: 0.9;
}

.btn-reset {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 14px 20px;
  background: #f3f4f6;
  border: none;
  border-radius: 12px;
  font-size: 14px;
  font-weight: 500;
  color: #6b7280;
  cursor: pointer;
  transition: all 0.3s ease;
}

.dark .btn-reset {
  background: #2d2d2d;
  color: #9ca3af;
}

.btn-reset:hover {
  background: #e5e7eb;
  color: #4b5563;
}

.dark .btn-reset:hover {
  background: #404040;
  color: #d1d5db;
}

/* 右侧预览区域 */
.trading-preview {
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  border-radius: 16px;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.dark .trading-preview {
  background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
}

.trading-preview.empty {
  justify-content: center;
  align-items: center;
  min-height: 200px;
}

.preview-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: #64748b;
  margin-bottom: 4px;
}

.dark .preview-header {
  color: #94a3b8;
}

.preview-header h4 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.preview-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  color: #94a3b8;
  text-align: center;
}

.preview-placeholder svg {
  font-size: 48px;
  opacity: 0.5;
}

.preview-placeholder p {
  margin: 0;
  font-size: 14px;
}

/* 预览内容 */
.preview-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.preview-main {
  background: white;
  border-radius: 12px;
  padding: 16px;
  text-align: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.dark .preview-main {
  background: #1e1e1e;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

.preview-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
}

.preview-item.total {
  flex-direction: column;
  gap: 8px;
  padding: 8px 0;
}

.preview-item.total .label {
  font-size: 13px;
  color: #6b7280;
}

.dark .preview-item.total .label {
  color: #9ca3af;
}

.preview-item.total .value {
  font-size: 28px;
  font-weight: 700;
  color: #1f2937;
}

.dark .preview-item.total .value {
  color: #f3f4f6;
}

.preview-divider {
  height: 1px;
  background: #e2e8f0;
  margin: 4px 0;
}

.dark .preview-divider {
  background: #334155;
}

.preview-details {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.preview-details .preview-item {
  padding: 8px 0;
}

.preview-details .label {
  font-size: 13px;
  color: #64748b;
}

.dark .preview-details .label {
  color: #94a3b8;
}

.preview-details .value {
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
}

.dark .preview-details .value {
  color: #f3f4f6;
}

.preview-details .value.highlight {
  color: #6366f1;
}

.preview-details .value.positive {
  color: #10b981;
}

.preview-details .value.negative {
  color: #ef4444;
}

.preview-item.impact {
  margin-top: 8px;
  padding-top: 12px;
  border-top: 1px dashed #cbd5e1;
}

.dark .preview-item.impact {
  border-top-color: #475569;
}

.preview-item.impact .label {
  font-weight: 600;
  color: #475569;
}

.dark .preview-item.impact .label {
  color: #94a3b8;
}

.preview-item.impact .value {
  font-size: 16px;
  font-weight: 700;
}

.portfolio-section,
.trades-section {
  background-color: white;
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 30px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.dark .portfolio-section,
.dark .trades-section {
  background-color: #1e1e1e;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.filter-group {
  display: flex;
  gap: 8px;
}

.filter-select {
  padding: 8px 12px;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  font-size: 14px;
  background-color: white;
  color: #1f2937;
  cursor: pointer;
}

.dark .filter-select {
  background-color: #2d2d2d;
  border-color: #404040;
  color: #f3f4f6;
}

.table-wrapper {
  overflow-x: auto;
}

.portfolio-table,
.trades-table {
  width: 100%;
  border-collapse: collapse;
}

.portfolio-table th,
.trades-table th {
  text-align: left;
  padding: 12px 16px;
  font-size: 12px;
  font-weight: 600;
  color: #6b7280;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-bottom: 1px solid #e5e7eb;
}

.dark .portfolio-table th,
.dark .trades-table th {
  color: #9ca3af;
  border-bottom-color: #404040;
}

.portfolio-table td,
.trades-table td {
  padding: 16px;
  border-bottom: 1px solid #f3f4f6;
}

.dark .portfolio-table td,
.dark .trades-table td {
  border-bottom-color: #2d2d2d;
}

.asset-row {
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.asset-row:hover {
  background-color: #f9fafb;
}

.dark .asset-row:hover {
  background-color: #2d2d2d;
}

.asset-row.selected {
  background-color: #eff6ff;
}

.dark .asset-row.selected {
  background-color: #1e3a5f;
}

.asset-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.asset-name {
  font-weight: 600;
  color: #1f2937;
}

.dark .asset-name {
  color: #f3f4f6;
}

.asset-symbol {
  font-size: 12px;
  color: #6b7280;
}

.dark .asset-symbol {
  color: #9ca3af;
}

.asset-amount,
.asset-price,
.asset-value {
  font-family: 'Courier New', monospace;
  font-size: 14px;
  color: #4b5563;
}

.dark .asset-amount,
.dark .asset-price,
.dark .asset-value {
  color: #d1d5db;
}

.asset-profit {
  font-family: 'Courier New', monospace;
}

.asset-profit .positive {
  color: #10b981;
}

.asset-profit .negative {
  color: #ef4444;
}

.profit-value {
  font-weight: 600;
  font-size: 14px;
}

.profit-rate {
  font-size: 12px;
  opacity: 0.8;
}

/* 徽章基础样式 */
.status-badge,
.cost-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}

.status-badge {
  gap: 4px;
}

/* 回本状态 */
.status-badge.recovered,
.cost-badge.zero {
  background: #d1fae5;
  color: #059669;
}

.dark .status-badge.recovered,
.dark .cost-badge.zero {
  background: #064e3b;
  color: #34d399;
}

/* 超额回报状态 */
.status-badge.super-profit,
.cost-badge.negative {
  background: #fef3c7;
  color: #d97706;
}

.dark .status-badge.super-profit,
.dark .cost-badge.negative {
  background: #78350f;
  color: #fbbf24;
}

/* 成本价展示 */
.cost-display {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.cost-value {
  font-size: 14px;
  font-weight: 500;
}

.action-cell {
  display: flex;
  gap: 8px;
}

.btn-action,
.btn-delete {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-sell {
  background: #fef2f2;
  color: #ef4444;
}

.dark .btn-sell {
  background: rgba(239, 68, 68, 0.1);
}

.btn-sell:hover {
  background: #fecaca;
}

.btn-buy {
  background: #ecfdf5;
  color: #10b981;
}

.dark .btn-buy {
  background: rgba(16, 185, 129, 0.1);
}

.btn-buy:hover {
  background: #a7f3d0;
}

.btn-delete {
  background: transparent;
  color: #9ca3af;
}

.btn-delete:hover:not(:disabled) {
  color: #ef4444;
}

.btn-delete:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.empty-state {
  text-align: center;
  padding: 40px;
  color: #9ca3af;
}

.empty-state svg {
  font-size: 48px;
  margin-bottom: 12px;
}

.empty-state p {
  font-size: 14px;
}

.protect-switch {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 6px 12px;
  background: #f3f4f6;
  border-radius: 6px;
  font-size: 14px;
  color: #4b5563;
}

.dark .protect-switch {
  background: #2d2d2d;
  color: #d1d5db;
}

.switch-label {
  font-size: 12px;
}

.switch {
  width: 36px;
  height: 20px;
  background: #d1d5db;
  border-radius: 10px;
  position: relative;
  transition: background 0.3s ease;
}

.dark .switch {
  background: #4b5563;
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
  transition: transform 0.3s ease;
}

.switch.on .switch-handle {
  transform: translateX(16px);
}

.btn-clear,
.btn-export,
.btn-import {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-clear {
  background: #fef2f2;
  color: #ef4444;
  border: none;
}

.dark .btn-clear {
  background: rgba(239, 68, 68, 0.1);
}

.btn-clear:hover {
  background: #fecaca;
}

.btn-export {
  background: #eff6ff;
  color: #3b82f6;
  border: none;
}

.dark .btn-export {
  background: rgba(59, 130, 246, 0.1);
}

.btn-export:hover {
  background: #dbeafe;
}

.btn-import {
  background: #f3f4f6;
  color: #4b5563;
  border: none;
}

.dark .btn-import {
  background: #2d2d2d;
  color: #d1d5db;
}

.btn-import:hover {
  background: #e5e7eb;
}

.trade-row {
  transition: background-color 0.2s ease;
}

.trade-row:hover {
  background-color: #f9fafb;
}

.dark .trade-row:hover {
  background-color: #2d2d2d;
}

.trade-time {
  font-family: 'Courier New', monospace;
  font-size: 13px;
  color: #6b7280;
}

.dark .trade-time {
  color: #9ca3af;
}

.trade-asset {
  display: flex;
  align-items: center;
  gap: 8px;
}

.trade-type {
  display: inline-block;
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
}

.trade-type.buy {
  background: #ecfdf5;
  color: #059669;
}

.dark .trade-type.buy {
  background: rgba(16, 185, 129, 0.1);
}

.trade-type.sell {
  background: #fef2f2;
  color: #dc2626;
}

.dark .trade-type.sell {
  background: rgba(239, 68, 68, 0.1);
}

.trade-type.recharge {
  background: #eff6ff;
  color: #2563eb;
}

.dark .trade-type.recharge {
  background: rgba(59, 130, 246, 0.1);
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 400px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.2);
}

.dark .modal {
  background: #1e1e1e;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.4);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #e5e7eb;
}

.dark .modal-header {
  border-bottom-color: #404040;
}

.modal-header h3 {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
}

.dark .modal-header h3 {
  color: #f3f4f6;
}

.btn-close {
  background: none;
  border: none;
  color: #9ca3af;
  cursor: pointer;
  font-size: 20px;
  padding: 4px;
  transition: color 0.3s ease;
}

.btn-close:hover {
  color: #4b5563;
}

.dark .btn-close:hover {
  color: #d1d5db;
}

.modal-body {
  padding: 24px;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-size: 14px;
  font-weight: 500;
  color: #4b5563;
}

.dark .form-group label {
  color: #d1d5db;
}

.form-group input {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  font-size: 16px;
  background-color: white;
  color: #1f2937;
  transition: all 0.3s ease;
}

.dark .form-group input {
  background-color: #2d2d2d;
  border-color: #404040;
  color: #f3f4f6;
}

.form-group input:focus {
  outline: none;
  border-color: #6366f1;
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 20px 24px;
  border-top: 1px solid #e5e7eb;
}

.dark .modal-footer {
  border-top-color: #404040;
}

.btn-cancel {
  padding: 10px 20px;
  background: #f3f4f6;
  color: #4b5563;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.dark .btn-cancel {
  background: #2d2d2d;
  color: #d1d5db;
}

.btn-cancel:hover {
  background: #e5e7eb;
}

.btn-confirm {
  padding: 10px 20px;
  background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-confirm:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.4);
}

.btn-confirm:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.error-toast {
  position: fixed;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  background: #fef2f2;
  color: #dc2626;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  z-index: 1001;
  animation: slideUp 0.3s ease;
}

.dark .error-toast {
  background: rgba(239, 68, 68, 0.1);
  color: #f87171;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateX(-50%) translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateX(-50%) translateY(0);
  }
}

@media (max-width: 768px) {
  .overview {
    grid-template-columns: 1fr;
  }

  .section-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .section-actions {
    width: 100%;
    flex-wrap: wrap;
  }

  .chart-container {
    flex-direction: column;
    align-items: center;
  }

  .chart {
    flex: 0 0 auto;
  }

  .chart-legend {
    width: 100%;
  }

  .portfolio-table,
  .trades-table {
    font-size: 12px;
  }

  .portfolio-table th,
  .portfolio-table td,
  .trades-table th,
  .trades-table td {
    padding: 8px;
  }

  .action-cell {
    flex-direction: column;
    gap: 4px;
  }

  .btn-action,
  .btn-delete {
    width: 28px;
    height: 28px;
  }
}
</style>
