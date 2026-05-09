<template>
  <div class="portfolio-container">
    <div class="container">
      <div class="dashboard">
        <!-- 运行模式标识 -->
        <div class="mode-indicator" :class="config.mode">
          <Icon :icon="config.isBackend ? 'mdi:server' : 'mdi:database-outline'" />
          <span>{{ config.isBackend ? '后端模式' : '本地模式' }}</span>
        </div>
        <main class="main-content mobile-main-content">
          <!-- 加载状态 - 骨架屏 -->
          <SkeletonLoader v-if="isLoading && !hasLoaded" />
          
          <template v-else>
            <!-- 未登录提示（仅后端模式显示） -->
            <section v-if="config.isBackend && !userStore.isLoggedIn" class="login-prompt">
              <div class="prompt-content">
                <Icon icon="mdi:lock" class="prompt-icon" />
                <h3>请先登录</h3>
                <p>登录后可查看和管理您的资产组合</p>
                <button class="btn-login" @click="userStore.openLoginModal">
                  <Icon icon="mdi:login" /> 立即登录
                </button>
              </div>
            </section>

            <template v-else>
              <!-- 移动端顶部导航栏 -->
              <div class="mobile-top-nav">
                <button
                  v-for="tab in mobileTabs"
                  :key="tab.id"
                  class="top-nav-btn"
                  :class="{ active: activeTab === tab.id }"
                  @click="activeTab = tab.id"
                >
                  <Icon :icon="tab.icon" />
                  <span>{{ tab.name }}</span>
                </button>
              </div>

              <!-- 概览标签页 -->
            <section id="overview" class="overview tab-content" :class="{ 'mobile-hidden': isMobile && activeTab !== 'overview' }">
              <div class="overview-card total-card">
                <div class="card-header-row">
                  <h3><Icon icon="mdi:wallet" /> 总资产</h3>
                  <!-- 价值本位切换开关 -->
                  <button 
                    class="value-mode-toggle" 
                    :class="valueMode"
                    @click="toggleValueMode"
                    :title="valueMode === 'usd' ? '切换到BTC本位' : '切换到USD本位'"
                  >
                    <span class="mode-option" :class="{ active: valueMode === 'usd' }">
                      <Icon icon="mdi:currency-usd" />
                      <span>USD</span>
                    </span>
                    <span class="toggle-divider"></span>
                    <span class="mode-option" :class="{ active: valueMode === 'btc' }">
                      <Icon icon="mdi:currency-btc" />
                      <span>BTC</span>
                    </span>
                  </button>
                </div>
                <div class="value">{{ formatValue(totalAssetsValue) }}</div>
                <div class="sub-value">
                  加密: {{ formatValue(cryptoAssetsValue) }} |
                  美股: {{ formatValue(usStockValue) }} |
                  现金: {{ formatValue(cashBalance) }}
                </div>
              </div>
              <div class="overview-card cash-card">
                <h3><Icon icon="mdi:cash-usd" /> USD现金</h3>
                <div class="value">{{ formatValue(cashBalance) }}</div>
                <button class="btn-recharge" @click="showRechargeModal = true">
                  <Icon icon="mdi:plus" /> 充值
                </button>
              </div>
              <div class="overview-card">
                <h3><Icon icon="mdi:trending-up" /> 浮动盈亏</h3>
                <div class="value" :class="displayUnrealizedPL.class">
                  {{ displayUnrealizedPL.sign }}{{ formatValue(Math.abs(unrealizedPL)) }}
                </div>
                <div class="change" :class="displayUnrealizedPL.class">
                  {{ displayUnrealizedPL.rate }}
                </div>
              </div>
              <div class="overview-card">
                <h3><Icon icon="mdi:cash-multiple" /> 实现盈亏</h3>
                <div class="value" :class="displayRealizedPL.class">
                  {{ displayRealizedPL.sign }}{{ formatValue(Math.abs(realizedPL)) }}
                </div>
                <div class="change" :class="displayRealizedPL.class">
                  {{ displayRealizedPL.rate }}
                </div>
              </div>
            </section>

            <!-- 资产分布标签页 -->
            <section class="chart-section tab-content" :class="{ 'mobile-hidden': isMobile && activeTab !== 'distribution' }">
              <div class="chart-header">
                <h2 class="chart-title"><Icon icon="mdi:chart-pie" /> 资产分布</h2>
                <div class="chart-actions">
                  <!-- 视图切换按钮组 -->
                  <div class="view-toggle">
                    <button
                      v-for="view in chartViews"
                      :key="view.id"
                      class="view-toggle-btn"
                      :class="{ active: currentChartView === view.id }"
                      @click="currentChartView = view.id"
                      :title="view.name"
                    >
                      <Icon :icon="view.icon" />
                      <span class="view-label">{{ view.name }}</span>
                    </button>
                  </div>
                  <button
                    class="btn-refresh"
                    @click="refreshPrices"
                    :disabled="refreshing"
                    :class="{ 'refreshing': refreshing }"
                  >
                    <Icon :icon="refreshing ? 'mdi:loading' : 'mdi:refresh'" />
                    {{ refreshing ? '刷新中...' : '刷新数据' }}
                  </button>
                </div>
              </div>

              <!-- 加载状态 -->
              <div v-if="refreshing && currentAllocation.length === 0" class="chart-loading">
                <div class="loading-spinner">
                  <Icon icon="mdi:loading" class="spin-icon" />
                </div>
                <span>正在加载资产数据...</span>
              </div>

              <!-- 空状态 -->
              <div v-else-if="currentAllocation.length === 0" class="chart-empty">
                <Icon icon="mdi:chart-pie-outline" class="empty-icon" />
                <p>暂无资产数据</p>
                <span>开始交易后将显示资产分布</span>
              </div>

              <!-- 图表内容 -->
              <div v-else class="chart-container">
                <PortfolioChart
                  :allocation="currentAllocation"
                  :total-value="currentTotalValue"
                  :has-loaded="hasLoaded"
                  :format-value="formatValue"
                  :center-label="currentChartCenterLabel"
                />
              </div>
            </section>

            <!-- 交易区域 - 左右分栏布局 (仅PC端显示，移动端通过FAB调出) -->
            <section id="trading" class="trading-section desktop-only">
              <div class="trading-container">
                <!-- 左侧：交易表单 -->
                <TradeForm
                  v-model:trade-type="tradeFormState.tradeType"
                  v-model:asset-type="tradeFormState.assetType"
                  v-model:selected-symbol="tradeFormState.selectedSymbol"
                  v-model:amount="tradeFormState.amount"
                  v-model:price="tradeFormState.price"
                  :current-market-price="tradeFormState.currentMarketPrice"
                  :cash-balance="cashBalance"
                  :is-loading="portfolioStore.isLoading"
                  :is-submitting="isSubmitting.trade"
                  @submit="handleTradeSubmit"
                  @reset="resetTradeForm"
                  @symbol-select="(symbol) => { tradeFormState.currentMarketPrice = tradeFormState.price }"
                />

                <!-- 右侧：交易预览（仅PC端显示） -->
                <TradePreview
                  :trade-type="tradeFormState.tradeType"
                  :symbol="tradeFormState.selectedSymbol"
                  :amount="tradeFormState.amount || 0"
                  :price="tradeFormState.price || 0"
                  :current-holding="getHoldingAmount(tradeFormState.selectedSymbol)"
                  :current-avg-cost="getHoldingAvgCost(tradeFormState.selectedSymbol)"
                  :show-empty="!!tradeFormState.selectedSymbol"
                />
              </div>
            </section>

            <!-- 资产详情标签页 -->
            <section id="portfolio" class="portfolio-section tab-content" :class="{ 'mobile-hidden': isMobile && activeTab !== 'portfolio' }">
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

              <!-- PC端表格视图 -->
              <div class="table-wrapper desktop-view">
                <table class="portfolio-table">
                  <thead>
                    <tr>
                      <th>资产</th>
                      <th>持有量</th>
                      <th>成本价</th>
                      <th>当前价</th>
                      <th>总价值</th>
                      <th>浮动盈亏</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="crypto in holdingsWithMeta"
                      :key="crypto.id"
                      class="asset-row"
                      :class="{ 'selected': selectedAsset === crypto.symbol, 'liquidated': crypto.amount === 0 }"
                      @click="selectAsset(crypto.symbol)"
                    >
                      <td>
                        <div class="asset-info">
                          <Icon width="32" height="32" :icon="crypto.icon" :style="{ color: crypto.color }" />
                          <div>
                            <div class="asset-name">{{ crypto.name }}</div>
                            <div class="asset-symbol">{{ crypto.symbol }}</div>
                          </div>
                        </div>
                      </td>
                      <td class="asset-amount">{{ crypto.formattedAmount }}</td>
                      <td class="asset-price">
                        <template v-if="crypto.avg_cost > 0">${{ crypto.formattedCost }}</template>
                        <template v-else-if="crypto.avg_cost === 0">
                          <div class="cost-display">
                            <span class="cost-value">$0</span>
                            <span class="cost-badge zero">已回本</span>
                          </div>
                        </template>
                        <template v-else>
                          <div class="cost-display">
                            <span class="cost-value">${{ crypto.formattedCost }}</span>
                            <span class="cost-badge negative">负成本</span>
                          </div>
                        </template>
                      </td>
                      <td class="asset-price">${{ formatAmount(crypto.current_price) }}</td>
                      <td class="asset-value">{{ crypto.formattedMarketValue }}</td>
                      <td class="asset-profit" :class="getProfitClass(crypto)">
                        <template v-if="crypto.avg_cost > 0">
                          <div class="profit-value">
                            {{ (crypto.amount * (crypto.current_price - crypto.avg_cost)) >= 0 ? '+' : '-' }}{{ formatValue(Math.abs(crypto.amount * (crypto.current_price - crypto.avg_cost))) }}
                          </div>
                          <div class="profit-rate" v-if="crypto.symbol !== 'USDT'">
                            {{ ((crypto.current_price - crypto.avg_cost) / crypto.avg_cost * 100) >= 0 ? '+' : '-' }}{{ Math.abs((crypto.current_price - crypto.avg_cost) / crypto.avg_cost * 100).toFixed(2) }}%
                          </div>
                        </template>
                        <template v-else-if="crypto.avg_cost === 0">
                          <div class="profit-value positive">
                            +{{ formatValue(crypto.amount * crypto.current_price) }}
                          </div>
                          <div class="profit-rate" v-if="crypto.symbol !== 'USDT'">
                            <span class="status-badge recovered">✓ 已回本</span>
                          </div>
                        </template>
                        <template v-else>
                          <div class="profit-value positive">
                            +{{ formatValue(crypto.amount * crypto.current_price - crypto.avg_cost * crypto.amount) }}
                          </div>
                          <div class="profit-rate" v-if="crypto.symbol !== 'USDT'">
                            <span class="status-badge super-profit">🚀 超100%回报</span>
                          </div>
                        </template>
                      </td>
                    </tr>
                    <tr v-if="filteredHoldings.length === 0">
                      <td colspan="6" class="empty-state">
                        <Icon icon="mdi:inbox" />
                        <p>暂无资产数据，请充值USD后开始交易</p>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <!-- 移动端卡片视图 -->
              <div class="mobile-asset-list mobile-view">
                <MobileAssetCard
                  v-for="crypto in holdingsWithMeta"
                  :key="crypto.id"
                  :symbol="crypto.symbol"
                  :name="crypto.name"
                  :icon="crypto.icon"
                  :color="crypto.color"
                  :amount="crypto.amount"
                  :avg-cost="crypto.avg_cost"
                  :current-price="crypto.current_price"
                  :market-value="crypto.market_value"
                  :realized-pl="crypto.realized_pl"
                  :selected="selectedAsset === crypto.symbol"
                  :format-value-fn="formatValue"
                  @click="selectAsset"
                />
                <div v-if="holdingsWithMeta.length === 0" class="empty-state mobile-empty">
                  <Icon icon="mdi:inbox" />
                  <p>暂无资产数据</p>
                  <span>充值USD后开始交易</span>
                </div>
              </div>
            </section>

            <!-- 历史标签页 -->
            <section id="trades" class="trades-section tab-content" :class="{ 'mobile-hidden': isMobile && activeTab !== 'history' }">
              <TradeHistory
                v-model:filter="tradeFilter"
                v-model:protect-history="protectHistory"
                :trades="trades"
                :is-submitting-export="isSubmitting.export"
                :is-submitting-clear="isSubmitting.clear"
                :is-submitting-delete="isSubmitting.delete"
                @delete="deleteTrade"
                @clear="clearTrades"
                @export="exportData"
                @import="showImportDialog = true"
              />
            </section>
            </template>
          </template>
        </main>
      </div>
    </div>

    <!-- 充值弹窗 -->
    <div v-if="showRechargeModal" class="modal-overlay" @click.self="showRechargeModal = false">
      <div class="modal">
        <div class="modal-header">
          <h3><Icon icon="mdi:cash-plus" /> 充值USD</h3>
          <button class="btn-close" @click="showRechargeModal = false">
            <Icon icon="mdi:close" />
          </button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>充值金额 (USD)</label>
            <input
              type="number"
              v-model.number="rechargeAmount"
              placeholder="输入充值金额"
              min="0.01"
              step="0.01"
              @keyup.enter="rechargeUSD"
            >
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-cancel" @click="showRechargeModal = false">取消</button>
          <button class="btn-confirm" @click="rechargeUSD" :disabled="!rechargeAmount || rechargeAmount <= 0 || isSubmitting.recharge">
            确认充值
          </button>
        </div>
      </div>
    </div>

    <!-- 导入数据弹窗 -->
    <div v-if="showImportDialog" class="modal-overlay" @click.self="closeImportDialog">
      <div class="modal import-modal">
        <div class="modal-header">
          <h3><Icon icon="mdi:upload" /> 导入数据</h3>
          <button class="btn-close" @click="closeImportDialog">
            <Icon icon="mdi:close" />
          </button>
        </div>
        <div class="modal-body">
          <!-- 步骤1: 选择文件 -->
          <div v-if="importStep === 'select'" class="import-step">
            <div
              class="drop-zone"
              :class="{ 'drag-over': isDragging }"
              @dragover.prevent="isDragging = true"
              @dragleave.prevent="isDragging = false"
              @drop.prevent="handleFileDrop"
              @click="triggerFileInput"
            >
              <Icon icon="mdi:cloud-upload" class="drop-icon" />
              <p>点击或拖拽 JSON 文件到此处</p>
              <span class="drop-hint">支持 .json 格式，最大 10MB</span>
              <input
                ref="fileInput"
                type="file"
                accept=".json,application/json"
                @change="handleFileSelect"
                style="display: none"
              >
            </div>
          </div>

          <!-- 步骤2: 预览确认 -->
          <div v-if="importStep === 'preview'" class="import-step">
            <div class="preview-summary">
              <div class="preview-item">
                <span class="preview-label">总记录数</span>
                <span class="preview-value">{{ importPreview.total_trades }}</span>
              </div>
              <div class="preview-item success">
                <span class="preview-label">新记录</span>
                <span class="preview-value">{{ importPreview.new_trades }}</span>
              </div>
              <div class="preview-item warning" v-if="importPreview.conflicts > 0">
                <span class="preview-label">冲突</span>
                <span class="preview-value">{{ importPreview.conflicts }}</span>
              </div>
            </div>

            <div v-if="importPreview.conflicts > 0" class="conflict-section">
              <h4>冲突处理</h4>
              <div class="conflict-options">
                <label class="radio-label">
                  <input type="radio" v-model="conflictStrategy" value="skip">
                  <span>跳过冲突记录（推荐）</span>
                </label>
                <label class="radio-label">
                  <input type="radio" v-model="conflictStrategy" value="overwrite">
                  <span>覆盖现有记录</span>
                </label>
              </div>

              <div class="conflict-list">
                <div v-for="(item, index) in importPreview.conflict_items" :key="index" class="conflict-item">
                  <Icon icon="mdi:alert" class="conflict-icon" />
                  <div class="conflict-info">
                    <span class="conflict-symbol">{{ item.trade.symbol }}</span>
                    <span class="conflict-type">{{ formatTradeType(item.trade.type) }}</span>
                    <span class="conflict-reason">{{ item.reason }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 步骤3: 导入结果 -->
          <div v-if="importStep === 'result'" class="import-step">
            <div class="result-success" v-if="importResult.success">
              <Icon icon="mdi:check-circle" class="result-icon success" />
              <h4>导入成功</h4>
              <div class="result-stats">
                <div class="result-stat">
                  <span class="stat-value">{{ importResult.imported }}</span>
                  <span class="stat-label">成功导入</span>
                </div>
                <div class="result-stat" v-if="importResult.skipped > 0">
                  <span class="stat-value">{{ importResult.skipped }}</span>
                  <span class="stat-label">已跳过</span>
                </div>
                <div class="result-stat" v-if="importResult.overwritten > 0">
                  <span class="stat-value">{{ importResult.overwritten }}</span>
                  <span class="stat-label">已覆盖</span>
                </div>
              </div>
            </div>
            <div class="result-error" v-else>
              <Icon icon="mdi:close-circle" class="result-icon error" />
              <h4>导入失败</h4>
              <p>{{ importError }}</p>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-cancel" @click="closeImportDialog">取消</button>
          <button
            v-if="importStep === 'preview'"
            class="btn-confirm"
            @click="confirmImport"
            :disabled="isSubmitting.import"
          >
            确认导入
          </button>
          <button
            v-if="importStep === 'result' && importResult.success"
            class="btn-confirm"
            @click="closeImportDialog"
          >
            完成
          </button>
        </div>
      </div>
    </div>

    <!-- 错误提示 -->
    <div v-if="errorMessage" class="error-toast">
      <Icon icon="mdi:alert-circle" />
      <span>{{ errorMessage }}</span>
    </div>

    <!-- 移动端FAB按钮 -->
    <button
      v-if="isMobile"
      class="fab-btn"
      :class="{ hidden: !showFab }"
      @click="showTradeModal = true"
      title="快速交易"
    >
      <Icon icon="mdi:swap-horizontal" />
    </button>

    <!-- 移动端交易弹窗 -->
    <div v-if="isMobile && showTradeModal" class="trade-modal-overlay" @click="showTradeModal = false" @touchmove.stop.prevent>
      <div class="trade-modal" @click.stop @touchmove.stop>
        <div class="trade-modal-header">
          <h3><Icon icon="mdi:swap-horizontal" /> 快速交易</h3>
          <button class="trade-modal-close" @click="showTradeModal = false">
            <Icon icon="mdi:close" />
          </button>
        </div>
        <div class="trade-modal-body">
          <TradeFormMobile
            v-model:trade-type="tradeFormState.tradeType"
            v-model:asset-type="tradeFormState.assetType"
            v-model:selected-symbol="tradeFormState.selectedSymbol"
            v-model:amount="tradeFormState.amount"
            v-model:price="tradeFormState.price"
            :current-market-price="tradeFormState.currentMarketPrice"
            :cash-balance="cashBalance"
            :is-loading="portfolioStore.isLoading"
            :is-submitting="isSubmitting.trade"
            @submit="handleTradeSubmit"
            @reset="resetTradeForm"
            @symbol-select="(symbol) => { tradeFormState.currentMarketPrice = tradeFormState.price }"
          />

          <TradePreview
            :trade-type="tradeFormState.tradeType"
            :symbol="tradeFormState.selectedSymbol"
            :amount="tradeFormState.amount || 0"
            :price="tradeFormState.price || 0"
            :current-holding="getHoldingAmount(tradeFormState.selectedSymbol)"
            :current-avg-cost="getHoldingAvgCost(tradeFormState.selectedSymbol)"
            :show-empty="false"
          />
        </div>

        <!-- 交易按钮 - 固定在底部 -->
        <div class="trade-modal-footer" v-if="tradeFormState.selectedSymbol">
          <button
            class="btn-submit"
            @click="handleTradeSubmit(tradeFormState)"
            :disabled="!tradeFormState.amount || !tradeFormState.price || portfolioStore.isLoading || isSubmitting.trade"
            :class="tradeFormState.tradeType"
          >
            <Icon :icon="tradeFormState.tradeType === 'buy' ? 'mdi:cart-plus' : 'mdi:cart-remove'" />
            {{ tradeFormState.tradeType === 'buy' ? '买入' : '卖出' }}
            <span class="submit-total" v-if="tradeFormState.amount && tradeFormState.price">
              {{ formatCompactAmount(tradeFormState.amount * tradeFormState.price) }}
            </span>
          </button>
          <button class="btn-reset" @click="resetTradeForm">
            <Icon icon="bx:reset" />
            重置
          </button>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick, watch, defineAsyncComponent } from 'vue'
import { Icon } from '@iconify/vue'
import { usePortfolioStore } from '../stores/portfolio'
import { useUserStore } from '../stores/user'
import { config } from '../config'
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js'
import { Doughnut } from 'vue-chartjs'
import { getAssetColor, getAssetIcon, getAssetName } from '../config/assets'
import { formatAmount, formatCompactAmount, formatDateTime, getChangeClass } from '../utils/format'
import MobileAssetCard from './MobileAssetCard.vue'
import MobileTradeCard from './MobileTradeCard.vue'
import SkeletonLoader from './SkeletonLoader.vue'
import TradeForm from './portfolio/TradeForm.vue'
import TradeFormMobile from './portfolio/TradeFormMobile.vue'
import TradePreview from './portfolio/TradePreview.vue'
import TradeHistory from './portfolio/TradeHistory.vue'

// 延迟加载非关键组件
const PortfolioChart = defineAsyncComponent(() => import('./PortfolioChart.vue'))

ChartJS.register(ArcElement, Tooltip, Legend)

const portfolioStore = usePortfolioStore()
const userStore = useUserStore()

// 交易表单状态
const tradeFormState = ref({
  tradeType: 'buy',
  assetType: 'crypto',
  selectedSymbol: '',
  amount: null,
  price: null,
  currentMarketPrice: 0
})
const tradeFilter = ref('all')
const refreshing = ref(false)
const errorMessage = ref('')
const autoRefresh = ref(false)
const showAssetSelector = ref(false)
const isMobile = ref(false)
const showFab = ref(false)
const lastScrollTop = ref(0)
const refreshInterval = ref(60)
const selectedFilter = ref('all')
const selectedAsset = ref(null)
const showRechargeModal = ref(false)
const rechargeAmount = ref(null)
const protectHistory = ref(true)
let refreshTimer = null

// 加载状态
const isLoading = ref(true)
const hasLoaded = ref(false)

// 本位模式（USD本位或BTC本位）- 从localStorage读取或默认USD
const valueMode = ref('usd') // 'usd' 或 'btc'
const btcPrice = computed(() => dashboardData.value?.btc_price || 0)

// 在挂载时从localStorage读取
onMounted(() => {
  const savedMode = localStorage.getItem('portfolio_value_mode')
  if (savedMode === 'btc' || savedMode === 'usd') {
    valueMode.value = savedMode
  }
})

// 移动端标签页配置
const mobileTabs = [
  { id: 'overview', name: '概览', icon: 'mdi:view-dashboard' },
  { id: 'distribution', name: '资产分布', icon: 'mdi:chart-pie' },
  { id: 'portfolio', name: '资产详情', icon: 'mdi:wallet' },
  { id: 'history', name: '历史', icon: 'mdi:history' }
]

// 当前激活的标签页
const activeTab = ref('overview')

// 移动端交易弹窗显示状态
const showTradeModal = ref(false)

// 处理交易提交
const handleTradeSubmit = async (tradeData) => {
  await addTrade(tradeData)
  if (!errorMessage.value) {
    showTradeModal.value = false
  }
}

// 重置交易表单
const resetTradeForm = () => {
  tradeFormState.value = {
    tradeType: tradeFormState.value.tradeType,
    assetType: tradeFormState.value.assetType,
    selectedSymbol: '',
    amount: null,
    price: null,
    currentMarketPrice: 0
  }
}

// 切换本位模式（移动端使用）
const toggleValueMode = () => {
  valueMode.value = valueMode.value === 'usd' ? 'btc' : 'usd'
  // 保存到localStorage
  localStorage.setItem('portfolio_value_mode', valueMode.value)
}

// 重复提交保护状态
const isSubmitting = ref({
  trade: false,
  recharge: false,
  delete: false,
  clear: false,
  export: false,
  import: false
})

// 导入相关状态
const showImportDialog = ref(false)
const importStep = ref('select')
const isDragging = ref(false)
const fileInput = ref(null)
const importData = ref(null)
const importPreview = ref({
  total_trades: 0,
  new_trades: 0,
  conflicts: 0,
  conflict_items: []
})
const conflictStrategy = ref('skip')
const importResult = ref({
  success: false,
  imported: 0,
  skipped: 0,
  overwritten: 0
})
const importError = ref('')

// 图表视图选项
const chartViews = [
  { id: 'total', name: '总资产', icon: 'mdi:chart-pie' },
  { id: 'non_cash', name: '非现金', icon: 'mdi:chart-donut' },
  { id: 'crypto', name: '加密资产', icon: 'mdi:bitcoin' },
  { id: 'us_stock', name: '美股', icon: 'mdi:chart-line' }
]
const currentChartView = ref('total')

// 当前图表中心标签
const currentChartCenterLabel = computed(() => {
  const view = chartViews.find(v => v.id === currentChartView.value)
  return view ? view.name : '总资产'
})

// 从store获取数据（后端已计算好）
const dashboardData = computed(() => portfolioStore.dashboardData)
const portfolio = computed(() => portfolioStore.portfolio)
const trades = computed(() => portfolioStore.trades)
const cryptoAssetsValue = computed(() => portfolioStore.cryptoAssetsValue) // 加密资产市值
const usStockValue = computed(() => portfolioStore.usStockValue) // 美股市值
const cashBalance = computed(() => portfolioStore.cashBalance) // USD现金余额
const totalAssetsValue = computed(() => portfolioStore.totalAssetsValue) // 总资产
const unrealizedPL = computed(() => portfolioStore.unrealizedPL) // 浮动盈亏
const unrealizedPLRate = computed(() => portfolioStore.unrealizedPLRate) // 浮动盈亏率
const realizedPL = computed(() => portfolioStore.realizedPL) // 实现盈亏
const realizedPLRate = computed(() => portfolioStore.realizedPLRate) // 实现盈亏率

// 创建显示值的通用函数
const createPLDisplay = (val, rate = null) => ({
  sign: val >= 0 ? '+' : '-',
  value: formatAmount(Math.abs(val)),
  class: val >= 0 ? 'positive' : 'negative',
  ...(rate !== null && {
    rate: (rate >= 0 ? '+' : '-') + Math.abs(rate).toFixed(2) + '%'
  })
})

// 格式化的显示值
const displayUnrealizedPL = computed(() =>
  createPLDisplay(unrealizedPL.value, unrealizedPLRate.value))

const displayRealizedPL = computed(() =>
  createPLDisplay(realizedPL.value, realizedPLRate.value))

// 本位格式化函数
const formatValue = (valueInUSD) => {
  if (valueMode.value === 'btc' && btcPrice.value > 0) {
    // 转换为 BTC (sats)
    const btcValue = valueInUSD / btcPrice.value
    if (btcValue >= 1) {
      return `₿${btcValue.toFixed(4)}`
    } else {
      // 小于1 BTC时显示为 sats
      const sats = Math.round(btcValue * 100000000)
      return `${sats.toLocaleString()} sats`
    }
  }
  // 默认 USD 本位
  return formatCompactAmount(valueInUSD)
}

// 获取持仓数量（用于交易预览）
const getHoldingAmount = (symbol) => {
  const asset = portfolio.value?.find(c => c.symbol === symbol)
  return asset ? asset.amount : 0
}

// 获取当前持仓成本价
const getHoldingAvgCost = (symbol) => {
  const asset = portfolio.value?.find(c => c.symbol === symbol)
  return asset ? asset.avg_cost : 0
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

const addTrade = async (tradeData) => {
  if (isSubmitting.value.trade) return

  if (!tradeData.symbol) {
    errorMessage.value = '请选择资产'
    setTimeout(() => errorMessage.value = '', 3000)
    return
  }

  if (!tradeData.amount || tradeData.amount <= 0) {
    errorMessage.value = '请输入大于 0 的数量'
    setTimeout(() => errorMessage.value = '', 3000)
    return
  }

  if (!tradeData.price || tradeData.price <= 0) {
    errorMessage.value = '请输入大于 0 的价格'
    setTimeout(() => errorMessage.value = '', 3000)
    return
  }

  isSubmitting.value.trade = true

  try {
    const result = await portfolioStore.createTrade({
      asset_type: tradeData.assetType,
      symbol: tradeData.symbol,
      type: tradeData.type,
      amount: tradeData.amount,
      price: tradeData.price
    })

    if (!result.success) {
      errorMessage.value = result.error
      setTimeout(() => errorMessage.value = '', 3000)
      return
    }

    // 重置表单
    resetTradeForm()
  } finally {
    isSubmitting.value.trade = false
  }
}

const rechargeUSD = async () => {
  if (isSubmitting.value.recharge) return

  if (!rechargeAmount.value || rechargeAmount.value <= 0) {
    errorMessage.value = '请输入有效的充值金额'
    setTimeout(() => errorMessage.value = '', 3000)
    return
  }

  isSubmitting.value.recharge = true

  try {
    const result = await portfolioStore.createTrade({
      asset_type: 'cash',  // 现金类型
      symbol: 'USD',
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

// 使用导入的工具函数
const formatDate = formatDateTime

const refreshPrices = async () => {
  if (refreshing.value) return

  refreshing.value = true
  errorMessage.value = ''

  try {
    const result = await portfolioStore.fetchDashboard({ useCache: false })

    if (!result.success) {
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
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
}

// 启动自动刷新
const startAutoRefresh = () => {
  // 页面不可见时不启动
  if (document.hidden) return
  
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
  refreshTimer = setInterval(() => {
    // 页面可见时才刷新
    if (!document.hidden) {
      refreshPrices()
    }
  }, refreshInterval.value * 60 * 1000)
}

// 停止自动刷新
const stopAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

// 页面可见性变化处理
const handleVisibilityChange = () => {
  if (document.hidden) {
    // 页面隐藏时停止自动刷新
    if (refreshTimer) {
      clearInterval(refreshTimer)
      refreshTimer = null
    }
  } else {
    // 页面显示时恢复自动刷新
    if (autoRefresh.value) {
      startAutoRefresh()
    }
  }
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

// 过滤后的持仓
const filteredHoldings = computed(() => {
  const filter = selectedFilter.value
  return portfolio.value?.filter(c =>
    c.symbol !== 'USDT' &&
    c.amount > 0 &&
    (filter === 'all' || c.symbol === filter)
  ) || []
})

// 缓存资产元数据
const holdingsWithMetaCache = createComputedCache()

function computeHoldingsWithMeta() {
  return filteredHoldings.value.map(item => ({
    ...item,
    icon: getAssetIcon(item.asset_type, item.symbol),
    color: getAssetColor(item.asset_type, item.symbol),
    name: getAssetName(item.asset_type, item.symbol),
    formattedAmount: formatAmount(item.amount),
    formattedMarketValue: formatValue(item.market_value),
    formattedCost: formatAmount(item.avg_cost),
    formattedPL: formatAmount(item.realized_pl)
  }))
}

const holdingsWithMeta = computed(() => {
  const cacheKey = `${selectedFilter.value}-${portfolio.value?.length || 0}-${valueMode.value}`
  return holdingsWithMetaCache.get(cacheKey, computeHoldingsWithMeta)
})

// 总资产净值 = 加密资产市值 + 美股市值 + USD现金余额
const totalNetWorth = computed(() => cryptoAssetsValue.value + usStockValue.value + cashBalance.value)

// 非现金资产总值
const nonCashAssetsValue = computed(() => cryptoAssetsValue.value + usStockValue.value)

// 为没有预定义颜色的资产分配颜色（使用 HSL 生成和谐配色）
const generateColor = (index, total) => {
  // 使用黄金角 (~137.5°) 生成均匀分布的色相
  const goldenAngle = 137.508
  const hue = (index * goldenAngle) % 360
  // 饱和度和亮度保持在舒适范围
  const saturation = 65 + (index % 3) * 10 // 65%-85%
  const lightness = 50 + (index % 2) * 10  // 50%-60%
  return `hsl(${hue}, ${saturation}%, ${lightness}%)`
}

// 合并低于5%的资产为"其他"
const mergeSmallItems = (allocation, minPercentage = 5) => {
  const significantItems = []
  let otherValue = 0
  let otherRawPercentage = 0

  allocation.forEach((item) => {
    if (item.rawPercentage >= minPercentage) {
      significantItems.push(item)
    } else {
      otherValue += item.value
      otherRawPercentage += item.rawPercentage
    }
  })

  // 如果有"其他"类别，添加到列表
  if (otherValue > 0) {
    significantItems.push({
      name: '其他',
      rawPercentage: otherRawPercentage,
      value: otherValue,
      color: '#9ca3af' // 灰色
    })
  }

  return significantItems
}

// 使用最大余数法确保百分比总和为100%
const normalizePercentages = (items) => {
  let remainingPercentage = 100

  // 先向下取整并计算余数
  const itemsWithRemainder = items.map(alloc => {
    const floorPercentage = Math.floor(alloc.rawPercentage)
    const remainder = alloc.rawPercentage - floorPercentage
    remainingPercentage -= floorPercentage
    return { ...alloc, floorPercentage, remainder }
  })

  // 按余数从大到小排序，给余数大的项+1
  itemsWithRemainder.sort((a, b) => b.remainder - a.remainder)

  // 调整百分比
  for (let i = 0; i < itemsWithRemainder.length; i++) {
    if (i < remainingPercentage) {
      itemsWithRemainder[i].percentage = itemsWithRemainder[i].floorPercentage + 1
    } else {
      itemsWithRemainder[i].percentage = itemsWithRemainder[i].floorPercentage
    }
  }

  // 恢复原始排序（按价值降序），保持颜色与资产对应
  return itemsWithRemainder.sort((a, b) => b.value - a.value)
}

// 构建资产分布列表的通用函数
const buildAllocation = (items, total, includeCash = false, cashValue = 0) => {
  if (total <= 0) return []

  const allocation = []

  // 添加资产项
  items.forEach((item) => {
    const marketValue = item.market_value
    if (marketValue > 0) {
      allocation.push({
        name: item.symbol,
        rawPercentage: (marketValue / total) * 100,
        value: marketValue,
        color: getAssetColor(item.asset_type, item.symbol)
      })
    }
  })

  // 添加现金
  if (includeCash && cashValue > 0) {
    allocation.push({
      name: 'USD',
      rawPercentage: (cashValue / total) * 100,
      value: cashValue,
      color: '#10b981' // 绿色表示现金
    })
  }

  // 按价值降序排列
  allocation.sort((a, b) => b.value - a.value)

  // 为没有预定义颜色的资产分配颜色
  allocation.forEach((alloc, index) => {
    if (!alloc.color) {
      alloc.color = generateColor(index, allocation.length)
    }
  })

  // 合并小额资产并标准化百分比
  const significantItems = mergeSmallItems(allocation)
  return normalizePercentages(significantItems)
}

// 总资产分布（包含加密资产、美股和现金）
const portfolioAllocation = computed(() => {
  const portfolioItems = portfolio.value || []
  if (portfolioItems.length === 0 && cashBalance.value <= 0) return []

  return buildAllocation(portfolioItems, totalNetWorth.value, true, cashBalance.value)
})

// 非现金资产分布（仅加密资产和美股）
const nonCashAllocation = computed(() => {
  const portfolioItems = portfolio.value || []
  if (portfolioItems.length === 0) return []

  return buildAllocation(portfolioItems, nonCashAssetsValue.value, false, 0)
})

// 加密资产分布
const cryptoAllocation = computed(() => {
  const portfolioItems = portfolio.value || []
  const cryptoItems = portfolioItems.filter(item => item.asset_type === 'crypto')

  if (cryptoItems.length === 0) return []

  return buildAllocation(cryptoItems, cryptoAssetsValue.value, false, 0)
})

// 美股资产分布
const usStockAllocation = computed(() => {
  const portfolioItems = portfolio.value || []
  const stockItems = portfolioItems.filter(item => item.asset_type === 'us_stock')

  if (stockItems.length === 0) return []

  return buildAllocation(stockItems, usStockValue.value, false, 0)
})

// 当前显示的资产分布
const currentAllocation = computed(() => {
  switch (currentChartView.value) {
    case 'non_cash':
      return nonCashAllocation.value
    case 'crypto':
      return cryptoAllocation.value
    case 'us_stock':
      return usStockAllocation.value
    case 'total':
    default:
      return portfolioAllocation.value
  }
})

// 当前显示的总值
const currentTotalValue = computed(() => {
  switch (currentChartView.value) {
    case 'non_cash':
      return nonCashAssetsValue.value
    case 'crypto':
      return cryptoAssetsValue.value
    case 'us_stock':
      return usStockValue.value
    case 'total':
    default:
      return totalAssetsValue.value
  }
})

let resizeTimer = null
const checkMobile = () => {
  if (resizeTimer) clearTimeout(resizeTimer)
  resizeTimer = setTimeout(() => {
    isMobile.value = window.innerWidth <= 768
  }, 250)
}

// 处理滚动事件，控制FAB按钮显示/隐藏
// 逻辑：页面加载后隐藏，只有向上滚动时才显示
const handleScroll = () => {
  const scrollTop = window.pageYOffset || document.documentElement.scrollTop

  // 向上滚动时显示FAB
  if (scrollTop < lastScrollTop.value) {
    showFab.value = true
  }
  // 向下滚动时隐藏FAB
  else if (scrollTop > lastScrollTop.value) {
    showFab.value = false
  }

  lastScrollTop.value = scrollTop
}

onMounted(async () => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
  // 添加滚动监听
  window.addEventListener('scroll', handleScroll, { passive: true })

  if (!config.isBackend || userStore.isLoggedIn) {
    isLoading.value = true
    try {
      // 使用分阶段加载：先显示缓存数据，再后台刷新
      await portfolioStore.fetchDashboardStaged()
    } finally {
      isLoading.value = false
      hasLoaded.value = true
    }
  } else {
    hasLoaded.value = true
  }
  // 监听页面可见性变化
  document.addEventListener('visibilitychange', handleVisibilityChange)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
  window.removeEventListener('scroll', handleScroll)
  if (resizeTimer) clearTimeout(resizeTimer)
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
  // 移除页面可见性监听
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})

// ========== 导入/导出方法 ==========

// 导出数据
const exportData = async () => {
  if (isSubmitting.value.export) return

  isSubmitting.value.export = true
  try {
    const result = await portfolioStore.exportData()
    if (result.success) {
      // 生成 JSON 文件并下载
      const blob = new Blob([JSON.stringify(result.data, null, 2)], {
        type: 'application/json'
      })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `portfolio-backup-${formatDate(new Date()).replace(/[/:]/g, '-')}.json`
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
    } else {
      errorMessage.value = result.error || '导出失败'
      setTimeout(() => errorMessage.value = '', 3000)
    }
  } catch (error) {
    console.error('Export error:', error)
    errorMessage.value = '导出失败，请稍后重试'
    setTimeout(() => errorMessage.value = '', 3000)
  } finally {
    isSubmitting.value.export = false
  }
}

// 触发文件选择
const triggerFileInput = () => {
  fileInput.value?.click()
}

// 处理文件选择
const handleFileSelect = (event) => {
  const file = event.target.files[0]
  if (file) {
    processFile(file)
  }
}

// 处理文件拖拽
const handleFileDrop = (event) => {
  isDragging.value = false
  const file = event.dataTransfer.files[0]
  if (file) {
    processFile(file)
  }
}

// 处理文件
const processFile = (file) => {
  // 验证文件类型
  if (!file.name.endsWith('.json') && file.type !== 'application/json') {
    errorMessage.value = '请选择 JSON 格式的文件'
    setTimeout(() => errorMessage.value = '', 3000)
    return
  }

  // 验证文件大小 (10MB)
  if (file.size > 10 * 1024 * 1024) {
    errorMessage.value = '文件大小不能超过 10MB'
    setTimeout(() => errorMessage.value = '', 3000)
    return
  }

  const reader = new FileReader()
  reader.onload = async (e) => {
    try {
      const data = JSON.parse(e.target.result)

      // 验证基本结构
      if (!data.version || !data.trades || !Array.isArray(data.trades)) {
        errorMessage.value = '无效的数据文件格式'
        setTimeout(() => errorMessage.value = '', 3000)
        return
      }

      importData.value = data

      // 调用预览接口
      const result = await portfolioStore.importPreview(data)
      if (result.success) {
        importPreview.value = result.preview
        importStep.value = 'preview'
      } else {
        errorMessage.value = result.error || '预览失败'
        setTimeout(() => errorMessage.value = '', 3000)
      }
    } catch (error) {
      console.error('File parse error:', error)
      errorMessage.value = '文件解析失败，请检查文件格式'
      setTimeout(() => errorMessage.value = '', 3000)
    }
  }
  reader.readAsText(file)
}

// 确认导入
const confirmImport = async () => {
  if (isSubmitting.value.import || !importData.value) return

  isSubmitting.value.import = true
  try {
    const result = await portfolioStore.importConfirm(importData.value, conflictStrategy.value)
    if (result.success) {
      importResult.value = {
        success: true,
        imported: result.imported,
        skipped: result.skipped,
        overwritten: result.overwritten
      }
      importStep.value = 'result'
      // 刷新数据
      await portfolioStore.fetchDashboard()
    } else {
      importResult.value = { success: false }
      importError.value = result.error || '导入失败'
      importStep.value = 'result'
    }
  } catch (error) {
    console.error('Import error:', error)
    importResult.value = { success: false }
    importError.value = '导入失败，请稍后重试'
    importStep.value = 'result'
  } finally {
    isSubmitting.value.import = false
  }
}

// 关闭导入对话框
const closeImportDialog = () => {
  showImportDialog.value = false
  // 重置状态
  setTimeout(() => {
    importStep.value = 'select'
    importData.value = null
    importPreview.value = {
      total_trades: 0,
      new_trades: 0,
      conflicts: 0,
      conflict_items: []
    }
    conflictStrategy.value = 'skip'
    importResult.value = {
      success: false,
      imported: 0,
      skipped: 0,
      overwritten: 0
    }
    importError.value = ''
    if (fileInput.value) {
      fileInput.value.value = ''
    }
  }, 300)
}

// 格式化交易类型
const formatTradeType = (type) => {
  const typeMap = {
    'buy': '买入',
    'sell': '卖出',
    'recharge': '充值'
  }
  return typeMap[type] || type
}

// 监听登录状态变化
watch(() => userStore.isLoggedIn, async (isLoggedIn) => {
  if (isLoggedIn) {
    await portfolioStore.fetchDashboard()
  }
})
</script>

<style scoped>
/* iOS Safari 橡皮筋效果修复 - 类似微信的做法 */
.portfolio-container {
  min-height: calc(100vh - 120px);
  /* 确保在 iOS 上底部内容不被遮挡 */
  padding-bottom: constant(safe-area-inset-bottom);
  padding-bottom: env(safe-area-inset-bottom);
}

.container {
  max-width: 1400px;
  margin: 0 auto;
}

.dashboard {
  display: block;
}

/* 运行模式标识 */
.mode-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
  margin-bottom: 20px;
  width: fit-content;
}

.mode-indicator.backend {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.mode-indicator.frontend {
  background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
  color: white;
}

.mode-indicator svg {
  font-size: 16px;
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

.cash-card {
  position: relative;
}

.total-card .sub-value {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 8px;
  opacity: 0.8;
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

/* 视图切换按钮组 */
.view-toggle {
  display: flex;
  align-items: center;
  background: #f3f4f6;
  border-radius: 10px;
  padding: 4px;
  gap: 2px;
}

.dark .view-toggle {
  background: #374151;
}

.view-toggle-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: #6b7280;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.dark .view-toggle-btn {
  color: #9ca3af;
}

.view-toggle-btn:hover {
  color: #4b5563;
  background: rgba(255, 255, 255, 0.5);
}

.dark .view-toggle-btn:hover {
  color: #e5e7eb;
  background: rgba(255, 255, 255, 0.1);
}

.view-toggle-btn.active {
  background: white;
  color: #6366f1;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.dark .view-toggle-btn.active {
  background: #4b5563;
  color: #818cf8;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.view-toggle-btn svg {
  font-size: 16px;
}

.view-label {
  white-space: nowrap;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .view-toggle {
    padding: 3px;
  }

  .view-toggle-btn {
    padding: 6px 10px;
    font-size: 12px;
  }

  .view-toggle-btn .view-label {
    display: none;
  }

  .view-toggle-btn svg {
    font-size: 18px;
  }
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

/* 加载状态 */
.chart-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #6b7280;
  gap: 16px;
}

.dark .chart-loading {
  color: #9ca3af;
}

.loading-spinner {
  font-size: 48px;
  color: #6366f1;
}

.spin-icon {
  animation: spin 1s linear infinite;
}

/* 空状态 */
.chart-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
}

.empty-icon {
  font-size: 64px;
  color: #d1d5db;
  margin-bottom: 16px;
}

.dark .empty-icon {
  color: #4b5563;
}

.chart-empty p {
  font-size: 18px;
  font-weight: 600;
  color: #374151;
  margin: 0 0 8px 0;
}

.dark .chart-empty p {
  color: #e5e7eb;
}

.chart-empty span {
  font-size: 14px;
  color: #9ca3af;
}

.dark .chart-empty span {
  color: #6b7280;
}

/* 图表容器 */
.chart-container {
  width: 100%;
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
.trade-type-tabs,
.asset-type-tabs {
  display: flex;
  gap: 8px;
  background: #f3f4f6;
  padding: 4px;
  border-radius: 10px;
}

.dark .trade-type-tabs,
.dark .asset-type-tabs {
  background: #2d2d2d;
}

.type-tab {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 10px 20px;
  min-height: 44px;
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
  justify-content: center;
  gap: 4px;
  padding: 12px 8px;
  min-height: 64px;
  background: #f9fafb;
  border: 2px solid transparent;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.15s ease;
  color: #374151;
  touch-action: manipulation;
}

.asset-btn:active {
  transform: scale(0.98);
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
  flex-wrap: wrap;
}

.input-with-controls input {
  flex: 1;
  min-width: 0;
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

/* 快捷数量按钮 */
.quick-amount-buttons {
  display: flex;
  gap: 12px;
  margin-top: 12px;
}

.quick-btn {
  flex: 1;
  padding: 14px 12px;
  min-height: 48px;
  border: 1px solid #e5e7eb;
  background: white;
  color: #6b7280;
  border-radius: 12px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  touch-action: manipulation;
}

.quick-btn:hover {
  border-color: #6366f1;
  color: #6366f1;
}

.quick-btn:active {
  transform: scale(0.96);
}

.quick-btn.primary {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  border-color: #6366f1;
  color: white;
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.3);
}

.quick-btn.primary:hover {
  background: linear-gradient(135deg, #4f46e5, #7c3aed);
  border-color: #4f46e5;
}

.quick-btn.primary:active {
  box-shadow: 0 2px 6px rgba(99, 102, 241, 0.3);
}

.dark .quick-btn {
  background: #2d2d2d;
  border-color: #404040;
  color: #9ca3af;
}

.dark .quick-btn:hover {
  border-color: #4a90e2;
  color: #4a90e2;
}

.dark .quick-btn.primary {
  background: linear-gradient(135deg, #4a90e2, #6a5acd);
  border-color: #4a90e2;
  box-shadow: 0 4px 12px rgba(74, 144, 226, 0.3);
}

.dark .quick-btn.primary:hover {
  background: linear-gradient(135deg, #3d7bc6, #5b4cdb);
  border-color: #3d7bc6;
}

.btn-use-market {
  padding: 12px 20px;
  min-height: 44px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  border: 1px solid #6366f1;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  color: white;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
  flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.3);
  touch-action: manipulation;
}

.btn-use-market:hover {
  background: linear-gradient(135deg, #4f46e5, #7c3aed);
  box-shadow: 0 6px 16px rgba(99, 102, 241, 0.4);
}

.btn-use-market:active {
  transform: scale(0.96);
  box-shadow: 0 2px 6px rgba(99, 102, 241, 0.3);
}

.dark .btn-use-market {
  background: linear-gradient(135deg, #4a90e2, #6a5acd);
  border-color: #4a90e2;
  box-shadow: 0 4px 12px rgba(74, 144, 226, 0.3);
}

.dark .btn-use-market:hover {
  background: linear-gradient(135deg, #3d7bc6, #5b4cdb);
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
  transition: all 0.15s ease;
  touch-action: manipulation;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.btn-submit:active:not(:disabled) {
  transform: scale(0.98);
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
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
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
/* PC端交易预览 - 使用与移动端相同的样式 */
.trading-preview {
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  border-radius: 12px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.dark .trading-preview {
  background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
}

.trading-preview.empty {
  justify-content: center;
  align-items: center;
  min-height: 200px;
}

.trading-preview .preview-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 12px;
  border-bottom: 1px solid #e2e8f0;
}

.dark .trading-preview .preview-header-row {
  border-bottom-color: #334155;
}

.trading-preview .preview-title {
  font-size: 14px;
  font-weight: 600;
  color: #374151;
}

.dark .trading-preview .preview-title {
  color: #d1d5db;
}

.trading-preview .preview-total-value {
  font-size: 20px;
  font-weight: 700;
  color: #1f2937;
}

.dark .trading-preview .preview-total-value {
  color: #f3f4f6;
}

.trading-preview .preview-details-full {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.trading-preview .preview-detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.trading-preview .detail-label {
  font-size: 13px;
  color: #64748b;
}

.dark .trading-preview .detail-label {
  color: #94a3b8;
}

.trading-preview .detail-value {
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
}

.dark .trading-preview .detail-value {
  color: #f3f4f6;
}

.trading-preview .detail-value.highlight {
  color: #6366f1;
}

.trading-preview .detail-value.positive {
  color: #10b981;
}

.trading-preview .detail-value.negative {
  color: #ef4444;
}

.trading-preview .preview-detail-row.impact {
  margin-top: 4px;
  padding-top: 10px;
  border-top: 1px dashed #cbd5e1;
}

.dark .trading-preview .preview-detail-row.impact {
  border-top-color: #475569;
}

.trading-preview .preview-detail-row.impact .detail-label {
  font-weight: 600;
  color: #475569;
}

.dark .trading-preview .preview-detail-row.impact .detail-label {
  color: #94a3b8;
}

/* 空状态 */
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
  -webkit-overflow-scrolling: touch;
  scrollbar-width: thin;
}

.table-wrapper::-webkit-scrollbar {
  height: 6px;
}

.table-wrapper::-webkit-scrollbar-track {
  background: #f1f5f9;
  border-radius: 3px;
}

.table-wrapper::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 3px;
}

.table-wrapper::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
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

/* 已清仓资产样式 */
.asset-row.liquidated {
  opacity: 0.7;
  background-color: #f3f4f6;
}

.dark .asset-row.liquidated {
  background-color: #252525;
}

.asset-row.liquidated:hover {
  opacity: 0.9;
  background-color: #e5e7eb;
}

.dark .asset-row.liquidated:hover {
  background-color: #2d2d2d;
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

.btn-delete {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s ease;
  touch-action: manipulation;
  background: transparent;
  color: #9ca3af;
}

.btn-delete:active {
  transform: scale(0.95);
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
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes slideDown {
  from {
    opacity: 1;
    transform: translateY(0);
  }
  to {
    opacity: 0;
    transform: translateY(20px);
  }
}

.modal {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 400px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.2);
  animation: slideUp 0.3s ease;
}

@media (max-width: 768px) {
  .modal-overlay {
    align-items: flex-end;
  }

  .modal {
    width: 100%;
    max-width: 100%;
    max-height: 85vh;
    border-radius: 20px 20px 0 0;
    animation: slideUpFromBottom 0.3s ease;
  }

  @keyframes slideUpFromBottom {
    from {
      opacity: 0;
      transform: translateY(100%);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
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
  bottom: 100px;
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

  .portfolio-table,
  .trades-table {
    font-size: 12px;
    min-width: 768px;
  }

  .portfolio-table th,
  .portfolio-table td,
  .trades-table th,
  .trades-table td {
    padding: 10px 12px;
    white-space: nowrap;
  }

  .portfolio-table .asset-info {
    gap: 8px;
  }

  .portfolio-table .asset-info svg {
    width: 24px;
    height: 24px;
  }

  .btn-delete {
    width: 36px;
    height: 36px;
    border-radius: 10px;
  }

  .btn-delete svg {
    font-size: 18px;
  }
}

@media (max-width: 480px) {
  .table-wrapper {
    margin: 0 -16px;
    padding: 0 16px;
    width: calc(100% + 32px);
  }

  .portfolio-table th.realized-pl-col,
  .portfolio-table td.realized-pl-col {
    display: none;
  }

  .trades-table th.trade-price-col,
  .trades-table td.trade-price-col {
    display: none;
  }

  .btn-action,
  .btn-delete {
    width: 40px;
    height: 40px;
  }
}

/* ========== 导入对话框样式 ========== */
.import-modal {
  max-width: 560px;
  width: 90%;
}

@media (max-width: 768px) {
  .import-modal {
    width: 100%;
    max-width: 100%;
    max-height: 90vh;
    border-radius: 20px 20px 0 0;
  }
}

.import-step {
  padding: 20px 0;
}

/* 拖拽区域 */
.drop-zone {
  border: 2px dashed #d1d5db;
  border-radius: 12px;
  padding: 48px 32px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s ease;
  background: #f9fafb;
}

.dark .drop-zone {
  border-color: #4b5563;
  background: #1e1e1e;
}

.drop-zone:hover,
.drop-zone.drag-over {
  border-color: #6366f1;
  background: #eff6ff;
}

.dark .drop-zone:hover,
.dark .drop-zone.drag-over {
  background: rgba(99, 102, 241, 0.1);
}

.drop-icon {
  font-size: 48px;
  color: #9ca3af;
  margin-bottom: 16px;
}

.drop-zone:hover .drop-icon,
.drop-zone.drag-over .drop-icon {
  color: #6366f1;
}

.drop-zone p {
  font-size: 16px;
  color: #374151;
  margin-bottom: 8px;
}

.dark .drop-zone p {
  color: #e5e7eb;
}

.drop-hint {
  font-size: 13px;
  color: #9ca3af;
}

/* 预览摘要 */
.preview-summary {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
}

.preview-item {
  flex: 1;
  background: #f3f4f6;
  border-radius: 8px;
  padding: 16px;
  text-align: center;
}

.dark .preview-item {
  background: #2d2d2d;
}

.preview-item.success {
  background: #ecfdf5;
}

.dark .preview-item.success {
  background: rgba(16, 185, 129, 0.1);
}

.preview-item.warning {
  background: #fffbeb;
}

.dark .preview-item.warning {
  background: rgba(245, 158, 11, 0.1);
}

.preview-label {
  display: block;
  font-size: 13px;
  color: #6b7280;
  margin-bottom: 4px;
}

.dark .preview-label {
  color: #9ca3af;
}

.preview-value {
  display: block;
  font-size: 24px;
  font-weight: 600;
  color: #1f2937;
}

.dark .preview-value {
  color: #f3f4f6;
}

.preview-item.success .preview-value {
  color: #059669;
}

.preview-item.warning .preview-value {
  color: #d97706;
}

/* 冲突区域 */
.conflict-section {
  border-top: 1px solid #e5e7eb;
  padding-top: 20px;
}

.dark .conflict-section {
  border-color: #374151;
}

.conflict-section h4 {
  font-size: 14px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 12px;
}

.dark .conflict-section h4 {
  color: #e5e7eb;
}

.conflict-options {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 16px;
}

.radio-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 14px;
  color: #374151;
}

.dark .radio-label {
  color: #e5e7eb;
}

.radio-label input[type="radio"] {
  width: 16px;
  height: 16px;
  accent-color: #6366f1;
}

/* 冲突列表 */
.conflict-list {
  max-height: 200px;
  overflow-y: auto;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 12px;
}

.dark .conflict-list {
  border-color: #374151;
}

.conflict-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 0;
  border-bottom: 1px solid #f3f4f6;
}

.dark .conflict-item {
  border-color: #374151;
}

.conflict-item:last-child {
  border-bottom: none;
}

.conflict-icon {
  color: #f59e0b;
  font-size: 18px;
}

.conflict-info {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.conflict-symbol {
  font-weight: 600;
  color: #1f2937;
}

.dark .conflict-symbol {
  color: #f3f4f6;
}

.conflict-type {
  font-size: 12px;
  padding: 2px 8px;
  background: #f3f4f6;
  border-radius: 4px;
  color: #6b7280;
}

.dark .conflict-type {
  background: #374151;
  color: #9ca3af;
}

.conflict-reason {
  font-size: 12px;
  color: #9ca3af;
}

/* 导入结果 */
.result-success,
.result-error {
  text-align: center;
  padding: 32px;
}

.result-icon {
  font-size: 64px;
  margin-bottom: 16px;
}

.result-icon.success {
  color: #10b981;
}

.result-icon.error {
  color: #ef4444;
}

.result-success h4,
.result-error h4 {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 16px;
}

.result-success h4 {
  color: #059669;
}

.result-error h4 {
  color: #dc2626;
}

.result-stats {
  display: flex;
  justify-content: center;
  gap: 32px;
  margin-top: 24px;
}

.result-stat {
  text-align: center;
}

.stat-value {
  display: block;
  font-size: 32px;
  font-weight: 700;
  color: #6366f1;
}

.stat-label {
  display: block;
  font-size: 13px;
  color: #6b7280;
  margin-top: 4px;
}

.dark .stat-label {
  color: #9ca3af;
}

.result-error p {
  color: #6b7280;
  font-size: 14px;
}

.dark .result-error p {
  color: #9ca3af;
}

/* 响应式 */
@media (max-width: 640px) {
  .preview-summary {
    flex-direction: column;
    gap: 12px;
  }

  .result-stats {
    flex-direction: column;
    gap: 16px;
  }

  .drop-zone {
    padding: 32px 20px;
  }
}

/* 移动端快速导航 */
.mobile-quick-nav {
  display: none;
}

@media (max-width: 768px) {
  .mobile-quick-nav {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    z-index: 100;
    display: flex;
    gap: 8px;
    padding: 12px;
    padding-top: calc(12px + env(safe-area-inset-top));
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(20px);
    border-radius: 0;
    border-bottom: 1px solid var(--border-color);
    box-shadow: 0 2px 20px rgba(0, 0, 0, 0.08);
  }

  .mobile-main-content {
    padding-top: calc(80px + env(safe-area-inset-top));
  }

  .quick-nav-btn {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    padding: 10px 6px;
    border: none;
    background: transparent;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.15s ease;
    color: var(--text-secondary);
    touch-action: manipulation;
  }

  .quick-nav-btn svg {
    font-size: 19px;
  }

  .quick-nav-btn span {
    font-size: 10px;
  }

  .quick-nav-btn:active {
    background: rgba(0, 0, 0, 0.05);
    transform: scale(0.98);
  }

  .dark .quick-nav-btn:active {
    background: rgba(255, 255, 255, 0.05);
  }
}

/* 视图切换：桌面端和移动端 */
@media (min-width: 769px) {
  .mobile-view {
    display: none;
  }

  .mobile-only {
    display: none;
  }
}

@media (max-width: 768px) {
  .desktop-view {
    display: none;
  }

  .mobile-view {
    display: block;
  }

  .desktop-only {
    display: none;
  }

  /* 移动端隐藏模式标识 */
  .mode-indicator {
    display: none !important;
  }
}

/* 价值本位切换开关 - 新设计 */
.value-mode-toggle {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 4px;
  background: #f1f5f9;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
  cursor: pointer;
  transition: all 0.2s ease;
  touch-action: manipulation;
  font-size: 12px;
  font-weight: 600;
}

.value-mode-toggle:hover {
  background: #e2e8f0;
  transform: translateY(-1px);
}

.value-mode-toggle:active {
  transform: scale(0.98);
}

.mode-option {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 10px;
  border-radius: 8px;
  transition: all 0.2s ease;
  color: #64748b;
}

.mode-option svg {
  font-size: 14px;
}

.mode-option.active {
  background: white;
  color: #4361ee;
  box-shadow: 0 2px 8px rgba(67, 97, 238, 0.15);
}

.value-mode-toggle.btc .mode-option.active {
  color: #f7931a;
  box-shadow: 0 2px 8px rgba(247, 147, 26, 0.15);
}

.toggle-divider {
  width: 1px;
  height: 16px;
  background: #cbd5e1;
}

/* 深色模式价值切换开关 */
.dark .value-mode-toggle {
  background: #2d3748;
  border-color: #4a5568;
}

.dark .value-mode-toggle:hover {
  background: #374151;
}

.dark .mode-option {
  color: #9ca3af;
}

.dark .mode-option.active {
  background: #1e293b;
  color: #60a5fa;
}

.dark .value-mode-toggle.btc .mode-option.active {
  color: #fbbf24;
}

.dark .toggle-divider {
  background: #4b5563;
}

/* 卡片头部行 */
.card-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

/* 移动端顶部导航栏 */
.mobile-top-nav {
  display: none;
}

@media (max-width: 768px) {
  .mobile-top-nav {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    z-index: 100;
    display: flex;
    gap: 4px;
    padding: 8px 12px;
    padding-top: calc(8px + env(safe-area-inset-top));
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(20px);
    border-bottom: 1px solid var(--border-color);
    box-shadow: 0 2px 20px rgba(0, 0, 0, 0.08);
  }

  .top-nav-btn {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
    padding: 8px 4px;
    border: none;
    background: transparent;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.15s ease;
    color: var(--text-secondary);
    touch-action: manipulation;
  }

  .top-nav-btn svg {
    font-size: 18px;
  }

  .top-nav-btn span {
    font-size: 10px;
    white-space: nowrap;
  }

  .top-nav-btn:active {
    background: rgba(0, 0, 0, 0.05);
    transform: scale(0.98);
  }

  .top-nav-btn.active {
    background: linear-gradient(135deg, #6366f1, #8b5cf6);
    color: white;
  }

  .mobile-main-content {
    padding-top: calc(70px + env(safe-area-inset-top));
  }

  /* 移动端隐藏标签内容 */
  .tab-content.mobile-hidden {
    display: none;
  }

  /* 标签内容区域 - 修复 iOS 橡皮筋效果 */
  .tab-content {
    min-height: calc(100vh - 70px - env(safe-area-inset-top) - env(safe-area-inset-bottom));
    /* 确保内容区域可以正常滚动 */
    overflow-y: auto;
    -webkit-overflow-scrolling: touch;
    /* 增加底部内边距防止被遮挡 */
    padding-bottom: calc(20px + env(safe-area-inset-bottom));
    /* 弹窗打开时禁用外部滚动 */
    position: relative;
  }

  /* 弹窗打开时，禁用外部容器滚动 */
  .tab-content.modal-open {
    overflow: hidden;
  }
}

/* 深色模式移动端导航 */
@media (max-width: 768px) {
  .dark .mobile-top-nav {
    background: rgba(30, 30, 30, 0.95);
    border-bottom-color: rgba(255, 255, 255, 0.08);
  }

  .dark .top-nav-btn {
    color: #9ca3af;
  }

  .dark .top-nav-btn:active {
    background: rgba(255, 255, 255, 0.05);
  }

  .dark .top-nav-btn.active {
    background: linear-gradient(135deg, #4a90e2, #6a5acd);
    color: white;
  }
}

/* 移动端资产列表 */
.mobile-asset-list {
  margin-top: 8px;
}

.mobile-empty {
  padding: 32px 16px !important;
}

.mobile-empty span {
  display: block;
  font-size: 12px;
  margin-top: 4px;
}

/* FAB按钮 */
.fab-btn {
  position: fixed;
  right: 20px;
  bottom: calc(80px + env(safe-area-inset-bottom));
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
  color: white;
  border: none;
  box-shadow: 0 4px 20px rgba(99, 102, 241, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.3s ease;
  z-index: 200;
  touch-action: manipulation;
}

.fab-btn svg {
  font-size: 28px;
}

.fab-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 24px rgba(99, 102, 241, 0.5);
}

.fab-btn:active {
  transform: scale(0.95);
}

.fab-btn.hidden {
  transform: translateY(100px) scale(0.8);
  opacity: 0;
  pointer-events: none;
}

.dark .fab-btn {
  background: linear-gradient(135deg, #4a90e2 0%, #6a5acd 100%);
  box-shadow: 0 4px 20px rgba(74, 144, 226, 0.4);
}

/* 移动端交易弹窗 */
.trade-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  animation: fadeIn 0.2s ease;
}

.dark .trade-modal-overlay {
  background: rgba(0, 0, 0, 0.8);
}

.trade-modal {
  background: white;
  width: 100%;
  max-width: 420px;
  max-height: 85vh;
  border-radius: 20px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  animation: scaleIn 0.2s ease;
  position: relative;
}

.dark .trade-modal {
  background: #1e1e1e;
}

@keyframes scaleIn {
  from {
    opacity: 0;
    transform: scale(0.9);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

.trade-modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #e5e7eb;
  background: white;
  flex-shrink: 0;
}

.dark .trade-modal-header {
  background: #1e1e1e;
  border-bottom-color: #404040;
}

.trade-modal-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
  display: flex;
  align-items: center;
  gap: 8px;
}

.dark .trade-modal-header h3 {
  color: #f3f4f6;
}

.trade-modal-close {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: none;
  background: #f3f4f6;
  color: #6b7280;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s ease;
  touch-action: manipulation;
}

.trade-modal-close:hover {
  background: #e5e7eb;
}

.dark .trade-modal-close {
  background: #404040;
  color: #9ca3af;
}

.dark .trade-modal-close:hover {
  background: #525252;
}

.trade-modal-body {
  padding: 16px;
  overflow-y: auto;
  flex: 1;
  /* 底部留足空间，确保价格输入框不被底部按钮遮挡 */
  padding-bottom: calc(140px + env(safe-area-inset-bottom));
  /* 确保弹窗内部可以正常滚动 */
  -webkit-overflow-scrolling: touch;
  /* 阻止滚动事件传播到外部 */
  overscroll-behavior: contain;
}

/* 移动端优化 */
@media (max-width: 768px) {
  .trade-modal-body {
    padding: 12px;
    padding-bottom: calc(150px + env(safe-area-inset-bottom));
  }
}

/* 移动端交易预览 - 简化版 */
.trade-preview-compact {
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  border-radius: 12px;
  padding: 12px 16px;
  margin: 12px 0;
}

.dark .trade-preview-compact {
  background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
}

.preview-total-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.preview-total-label {
  font-size: 14px;
  color: #6b7280;
  font-weight: 500;
}

.dark .preview-total-label {
  color: #9ca3af;
}

.preview-total-value {
  font-size: 20px;
  font-weight: 700;
  color: #1f2937;
}

.dark .preview-total-value {
  color: #f3f4f6;
}

.btn-toggle-preview {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 8px;
  border: none;
  background: rgba(99, 102, 241, 0.1);
  color: #6366f1;
  font-size: 13px;
  font-weight: 500;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.dark .btn-toggle-preview {
  background: rgba(99, 102, 241, 0.2);
  color: #818cf8;
}

.btn-toggle-preview:active {
  transform: scale(0.98);
}

.preview-details-compact {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #e2e8f0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.dark .preview-details-compact {
  border-top-color: #334155;
}

.preview-detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
}

.preview-detail-row span:first-child {
  color: #64748b;
}

.dark .preview-detail-row span:first-child {
  color: #94a3b8;
}

.preview-detail-row span:last-child {
  font-weight: 600;
  color: #1f2937;
}

.dark .preview-detail-row span:last-child {
  color: #f3f4f6;
}

.preview-detail-row .highlight {
  color: #6366f1;
}

.preview-detail-row .positive {
  color: #10b981;
}

.preview-detail-row .negative {
  color: #ef4444;
}

/* 移动端交易弹窗底部固定按钮 */
.trade-modal-footer {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  background: white;
  border-top: 1px solid #e5e7eb;
  padding: 12px 16px;
  padding-bottom: calc(12px + env(safe-area-inset-bottom));
  display: flex;
  gap: 12px;
  z-index: 10;
}

/* 移动端优化 */
@media (max-width: 768px) {
  .trade-modal-footer {
    padding: 10px 12px;
    padding-bottom: calc(10px + env(safe-area-inset-bottom));
  }
}

.dark .trade-modal-footer {
  background: #1e1e1e;
  border-top-color: #404040;
}

.trade-modal-footer .btn-submit {
  flex: 1;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 0 16px;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  touch-action: manipulation;
}

/* 移动端优化 */
@media (max-width: 768px) {
  .trade-modal-footer .btn-submit {
    height: 40px;
    font-size: 13px;
  }
}

.trade-modal-footer .btn-submit.buy {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
}

.trade-modal-footer .btn-submit.sell {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
}

.trade-modal-footer .btn-submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.trade-modal-footer .btn-submit:not(:disabled):active {
  transform: scale(0.98);
}

.trade-modal-footer .btn-reset {
  width: 70px;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 0 12px;
  border: 1px solid #e5e7eb;
  background: white;
  color: #6b7280;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  touch-action: manipulation;
}

/* 移动端优化 */
@media (max-width: 768px) {
  .trade-modal-footer .btn-reset {
    width: 60px;
    height: 40px;
    font-size: 12px;
  }
}

.dark .trade-modal-footer .btn-reset {
  background: #2d2d2d;
  border-color: #404040;
  color: #9ca3af;
}

.trade-modal-footer .btn-reset:active {
  transform: scale(0.98);
  background: #f3f4f6;
}

/* 资产选择器头部 */
.asset-selector-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.btn-back-to-input {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  border: none;
  background: rgba(99, 102, 241, 0.1);
  color: #6366f1;
  font-size: 13px;
  font-weight: 500;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.dark .btn-back-to-input {
  background: rgba(99, 102, 241, 0.2);
  color: #818cf8;
}

.btn-back-to-input:active {
  transform: scale(0.98);
}

/* 已选择资产显示 */
.selected-asset-display {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  border-radius: 12px;
  padding: 12px 16px;
  margin: 12px 0;
}

.dark .selected-asset-display {
  background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
}

.selected-asset-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.selected-asset-info .iconify {
  font-size: 24px;
}

.selected-asset-name {
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
}

.dark .selected-asset-name {
  color: #f3f4f6;
}

.selected-asset-price {
  font-size: 13px;
  color: #6b7280;
  background: rgba(0, 0, 0, 0.05);
  padding: 2px 8px;
  border-radius: 6px;
}

.dark .selected-asset-price {
  color: #9ca3af;
  background: rgba(255, 255, 255, 0.1);
}

.btn-change-asset {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 8px 14px;
  border: 1px solid #e5e7eb;
  background: white;
  color: #6366f1;
  font-size: 13px;
  font-weight: 500;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.dark .btn-change-asset {
  background: #2d2d2d;
  border-color: #404040;
  color: #818cf8;
}

.btn-change-asset:active {
  transform: scale(0.98);
  background: #f3f4f6;
}

/* 交易预览 - 完整版 */
.trade-preview-full {
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  border-radius: 12px;
  padding: 16px;
  margin: 12px 0;
}

.dark .trade-preview-full {
  background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
}

.preview-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e2e8f0;
}

.dark .preview-header-row {
  border-bottom-color: #334155;
}

.preview-title {
  font-size: 14px;
  font-weight: 600;
  color: #374151;
}

.dark .preview-title {
  color: #d1d5db;
}

.preview-details-full {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.preview-details-full .preview-detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.preview-details-full .detail-label {
  font-size: 13px;
  color: #64748b;
}

.dark .preview-details-full .detail-label {
  color: #94a3b8;
}

.preview-details-full .detail-value {
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
}

.dark .preview-details-full .detail-value {
  color: #f3f4f6;
}

.preview-details-full .detail-value.highlight {
  color: #6366f1;
}

.preview-details-full .detail-value.positive {
  color: #10b981;
}

.preview-details-full .detail-value.negative {
  color: #ef4444;
}

.preview-details-full .preview-detail-row.impact {
  margin-top: 4px;
  padding-top: 10px;
  border-top: 1px dashed #cbd5e1;
}

.dark .preview-details-full .preview-detail-row.impact {
  border-top-color: #475569;
}

.preview-details-full .preview-detail-row.impact .detail-label {
  font-weight: 600;
  color: #475569;
}

.dark .preview-details-full .preview-detail-row.impact .detail-label {
  color: #94a3b8;
}

/* 移动端资产列表底部安全区域 - 修复 iOS 橡皮筋回弹遮挡问题 */
@media (max-width: 768px) {
  .assets-list {
    padding-bottom: calc(120px + env(safe-area-inset-bottom));
    /* 确保列表区域可以独立滚动 */
    -webkit-overflow-scrolling: touch;
  }

  .mobile-trades-list {
    padding-bottom: calc(120px + env(safe-area-inset-bottom));
    -webkit-overflow-scrolling: touch;
  }

  /* 移动端主容器增加底部安全区域 */
  .portfolio-container {
    padding-bottom: calc(env(safe-area-inset-bottom) + 20px);
  }
}
</style>
