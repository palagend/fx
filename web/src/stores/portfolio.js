import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { portfolioApi } from '../api/portfolio'
import { useUserStore } from './user'

export const usePortfolioStore = defineStore('portfolio', () => {
  const userStore = useUserStore()

  // 状态 - 从dashboard接口获取的原始数据
  const dashboardData = ref(null)
  const trades = ref([])
  const isLoading = ref(false)
  const error = ref(null)

  // 计算属性 - 价格数据
  const prices = computed(() => dashboardData.value?.prices || {})
  const priceChanges = computed(() => dashboardData.value?.price_changes || {})

  // 计算属性 - 投资组合（后端已计算好）
  const portfolio = computed(() => dashboardData.value?.portfolio || [])

  // 计算属性 - 统计数据（后端已计算好）
  const cryptoValue = computed(() => dashboardData.value?.crypto_value || 0) // 加密资产市值（不含 USDT）
  const totalAssetsValue = computed(() => dashboardData.value?.total_assets_value || 0) // 总资产价值（含 USDT）
  const usdtBalance = computed(() => dashboardData.value?.usdt_balance || 0) // USDT余额
  const unrealizedProfitLoss = computed(() => dashboardData.value?.unrealized_profit_loss || 0) // 浮动盈亏
  const realizedProfitLoss = computed(() => dashboardData.value?.realized_profit_loss || 0) // 实现盈亏
  const totalProfitLoss = computed(() => dashboardData.value?.total_profit_loss || 0) // 总盈亏
  const valueChange24h = computed(() => dashboardData.value?.value_change_24h || 0) // 24小时价值变化率

  // 计算属性 - 浮动盈亏率（优先使用后端计算的值）
  const unrealizedProfitLossRate = computed(() => {
    // 如果后端已计算，直接使用
    if (dashboardData.value?.unrealized_profit_loss_rate !== undefined) {
      return dashboardData.value.unrealized_profit_loss_rate
    }
    // 否则前端计算
    const portfolioData = portfolio.value
    const nonUSDTCost = portfolioData
      .filter(item => item.symbol !== 'USDT')
      .reduce((sum, item) => sum + item.cost, 0)
    return nonUSDTCost > 0 ? (unrealizedProfitLoss.value / nonUSDTCost) * 100 : 0
  })

  // Actions
  // 获取仪表盘聚合数据（价格+持仓+统计）
  async function fetchDashboard() {
    if (!userStore.isLoggedIn) {
      dashboardData.value = null
      trades.value = []
      return { success: false, error: '请先登录' }
    }

    isLoading.value = true
    error.value = null

    try {
      const [dashboardRes, tradesRes] = await Promise.all([
        portfolioApi.getDashboard(),
        portfolioApi.getTrades()
      ])

      dashboardData.value = dashboardRes.data
      trades.value = tradesRes.data.trades || []

      return {
        success: true,
        updatedAt: dashboardRes.data.updated_at
      }
    } catch (err) {
      error.value = err.response?.data?.error || '获取数据失败'
      console.error('获取仪表盘数据失败:', err)
      return { success: false, error: error.value }
    } finally {
      isLoading.value = false
    }
  }

  // 获取单个资产价格
  async function fetchAssetPrice(symbol) {
    try {
      const response = await portfolioApi.getAssetPrice(symbol)
      // 更新 dashboardData 中的价格
      if (dashboardData.value) {
        dashboardData.value.prices = {
          ...dashboardData.value.prices,
          [symbol]: response.data.price
        }
      }
      return {
        success: true,
        price: response.data.price,
        updatedAt: response.data.updated_at
      }
    } catch (err) {
      console.error(`获取${symbol}价格失败:`, err)
      return { success: false, error: err.response?.data?.error || '获取价格失败' }
    }
  }

  async function createTrade(trade) {
    if (!userStore.isLoggedIn) {
      return { success: false, error: '请先登录' }
    }

    isLoading.value = true
    error.value = null

    try {
      const response = await portfolioApi.createTrade(trade)
      await fetchDashboard()
      return { success: true, data: response.data }
    } catch (err) {
      error.value = err.response?.data?.error || '交易失败'
      return { success: false, error: error.value }
    } finally {
      isLoading.value = false
    }
  }

  async function deleteTrade(id) {
    if (!userStore.isLoggedIn) {
      return { success: false, error: '请先登录' }
    }

    try {
      await portfolioApi.deleteTrade(id)
      await fetchDashboard()
      return { success: true }
    } catch (err) {
      return { success: false, error: err.response?.data?.error || '删除失败' }
    }
  }

  async function clearAllTrades() {
    if (!userStore.isLoggedIn) {
      return { success: false, error: '请先登录' }
    }

    isLoading.value = true
    error.value = null

    try {
      await portfolioApi.clearTrades()
      await fetchDashboard()
      return { success: true }
    } catch (err) {
      error.value = err.response?.data?.error || '清空交易记录失败'
      return { success: false, error: error.value }
    } finally {
      isLoading.value = false
    }
  }

  return {
    // 状态
    dashboardData,
    trades,
    isLoading,
    error,
    // 计算属性
    prices,
    priceChanges,
    portfolio,
    cryptoValue,
    totalAssetsValue,
    usdtBalance,
    unrealizedProfitLoss,
    unrealizedProfitLossRate,
    realizedProfitLoss,
    totalProfitLoss,
    valueChange24h,
    // Actions
    fetchDashboard,
    fetchAssetPrice,
    createTrade,
    deleteTrade,
    clearAllTrades
  }
})
