import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { portfolioApi } from '../api/portfolio'
import { useUserStore } from './user'

export const usePortfolioStore = defineStore('portfolio', () => {
  const userStore = useUserStore()
  
  // 状态
  const holdings = ref([])
  const trades = ref([])
  const prices = ref({})
  const realizedProfitLoss = ref(0)
  const totalValue = ref(0)
  const usdtBalance = ref(0)
  const isLoading = ref(false)
  const error = ref(null)

  // 计算属性
  const portfolio = computed(() => {
    return holdings.value.map(holding => {
      const currentPrice = prices.value[holding.symbol] || holding.avg_cost
      const value = holding.amount * currentPrice
      const profitLoss = (currentPrice - holding.avg_cost) * holding.amount
      const profitLossRate = holding.avg_cost > 0 
        ? ((currentPrice - holding.avg_cost) / holding.avg_cost) * 100 
        : 0
      
      return {
        ...holding,
        currentPrice,
        value,
        profitLoss,
        profitLossRate
      }
    })
  })

  const totalValueChange24h = computed(() => {
    // 简化计算，实际应该基于历史数据
    return 0
  })

  const unrealizedProfitLoss = computed(() => {
    return portfolio.value.reduce((sum, item) => sum + item.profitLoss, 0)
  })

  const unrealizedProfitLossRate = computed(() => {
    const totalCost = portfolio.value.reduce((sum, item) => 
      sum + (item.amount * item.avg_cost), 0)
    return totalCost > 0 ? (unrealizedProfitLoss.value / totalCost) * 100 : 0
  })

  const realizedProfitLossRate = computed(() => {
    // 简化计算
    return realizedProfitLoss.value
  })

  // Actions
  async function fetchPrices() {
    try {
      const response = await portfolioApi.getAllPrices()
      prices.value = response.data.prices
    } catch (err) {
      console.error('获取价格失败:', err)
    }
  }

  async function fetchPortfolio() {
    if (!userStore.isLoggedIn) return
    
    isLoading.value = true
    error.value = null
    
    try {
      // 并行获取数据
      const [summaryRes, holdingsRes, tradesRes] = await Promise.all([
        portfolioApi.getSummary(),
        portfolioApi.getHoldings(),
        portfolioApi.getTrades()
      ])

      totalValue.value = summaryRes.data.total_value
      usdtBalance.value = summaryRes.data.usdt_balance
      realizedProfitLoss.value = summaryRes.data.realized_profit_loss
      holdings.value = summaryRes.data.holdings || holdingsRes.data.holdings
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
      
      // 刷新数据
      await fetchPortfolio()
      
      return { success: true, data: response.data }
    } catch (err) {
      error.value = err.response?.data?.error || '交易失败'
      return { success: false, error: error.value }
    } finally {
      isLoading.value = false
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
      
      // 刷新数据
      await fetchPortfolio()
      
      return { success: true }
    } catch (err) {
      error.value = err.response?.data?.error || '清空交易记录失败'
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
      
      // 刷新数据
      await fetchPortfolio()
      
      return { success: true }
    } catch (err) {
      return { success: false, error: err.response?.data?.error || '删除失败' }
    }
  }

  // 本地存储备份（用于离线支持）
  function saveToLocalStorage() {
    if (!userStore.isLoggedIn) return
    
    const data = {
      holdings: holdings.value,
      trades: trades.value,
      realizedProfitLoss: realizedProfitLoss.value,
      timestamp: Date.now()
    }
    localStorage.setItem(`portfolio_backup_${userStore.user.id}`, JSON.stringify(data))
  }

  function loadFromLocalStorage() {
    if (!userStore.isLoggedIn) return
    
    const saved = localStorage.getItem(`portfolio_backup_${userStore.user.id}`)
    if (saved) {
      try {
        const data = JSON.parse(saved)
        holdings.value = data.holdings || []
        trades.value = data.trades || []
        realizedProfitLoss.value = data.realizedProfitLoss || 0
      } catch (e) {
        console.error('加载本地备份失败:', e)
      }
    }
  }

  function clearLocalStorage() {
    if (!userStore.isLoggedIn) return
    localStorage.removeItem(`portfolio_backup_${userStore.user.id}`)
  }

  return {
    // 状态
    holdings,
    trades,
    prices,
    realizedProfitLoss,
    totalValue,
    usdtBalance,
    isLoading,
    error,
    
    // 计算属性
    portfolio,
    totalValueChange24h,
    unrealizedProfitLoss,
    unrealizedProfitLossRate,
    realizedProfitLossRate,
    
    // Actions
    fetchPrices,
    fetchPortfolio,
    createTrade,
    clearAllTrades,
    deleteTrade,
    saveToLocalStorage,
    loadFromLocalStorage,
    clearLocalStorage
  }
})
