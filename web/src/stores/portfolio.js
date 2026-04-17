import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { portfolioApi } from '../api/portfolio'
import { useUserStore } from './user'

export const usePortfolioStore = defineStore('portfolio', () => {
  const userStore = useUserStore()

  // 状态
  const holdings = ref([])
  const investments = ref([])
  const trades = ref([])
  const prices = ref({})
  const isLoading = ref(false)
  const error = ref(null)

  // 计算属性 - 投资组合（包含实时价格计算）
  const portfolio = computed(() => {
    return holdings.value.map(holding => {
      const currentPrice = prices.value[holding.symbol] || 0
      const value = holding.amount * currentPrice
      const cost = holding.amount * holding.avg_cost
      const profitLoss = value - cost
      const profitLossRate =  (profitLoss / cost) * 100

      return {
        ...holding,
        currentPrice,
        value,
        cost,
        profitLoss,
        profitLossRate
      }
    })
  })

  // 计算属性 - 总资产价值（包含USDT）
  const totalValue = computed(() => {
    return portfolio.value.reduce((sum, item) => sum + item.value, 0)
  })

  // 计算属性 - USDT余额
  const usdtBalance = computed(() => {
    const usdtHolding = holdings.value.find(h => h.symbol === 'USDT')
    return usdtHolding ? usdtHolding.amount : 0
  })

  // 计算属性 - 浮动盈亏（未实现盈亏）
  const unrealizedProfitLoss = computed(() => {
    return portfolio.value
      .filter(item => item.symbol !== 'USDT')
      .reduce((sum, item) => sum + item.profitLoss, 0)
  })

  // 计算属性 - 实现盈亏
  const realizedProfitLoss = computed(() => {
    return investments.value.reduce((sum, inv) => sum + inv.realized_pl, 0)
  })

  // 计算属性 - 总盈亏 = 浮动盈亏 + 实现盈亏
  const totalProfitLoss = computed(() => {
    return unrealizedProfitLoss.value + realizedProfitLoss.value
  })

  // Actions
  async function fetchPrices() {
    try {
      const response = await portfolioApi.getAllPrices()
      prices.value = response.data.prices
      return {
        success: true,
        prices: response.data.prices,
        priceChanges: response.data.price_changes,
        updatedAt: response.data.updated_at
      }
    } catch (err) {
      console.error('获取价格失败:', err)
      return { success: false, error: err.response?.data?.error || '获取价格失败' }
    }
  }

  async function fetchPortfolio() {
    if (!userStore.isLoggedIn) {
      holdings.value = []
      investments.value = []
      trades.value = []
      return
    }

    isLoading.value = true
    error.value = null

    try {
      // 并行获取数据
      const [portfolioRes, investmentsRes, tradesRes] = await Promise.all([
        portfolioApi.getPortfolio(),
        portfolioApi.getInvestments(),
        portfolioApi.getTrades()
      ])

      holdings.value = portfolioRes.data.portfolio || []
      investments.value = investmentsRes.data.investments || []
      trades.value = tradesRes.data.trades || []

      // 获取最新价格
      await fetchPrices()
    } catch (err) {
      error.value = err.response?.data?.error || '获取资产数据失败'
      console.error('获取资产组合失败:', err)
    } finally {
      isLoading.value = false
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
      await fetchPortfolio()
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
      await fetchPortfolio()
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
      await fetchPortfolio()
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
    holdings,
    investments,
    trades,
    prices,
    isLoading,
    error,
    // 计算属性
    portfolio,
    totalValue,
    usdtBalance,
    unrealizedProfitLoss,
    realizedProfitLoss,
    totalProfitLoss,
    // Actions
    fetchPrices,
    fetchPortfolio,
    createTrade,
    deleteTrade,
    clearAllTrades
  }
})
