import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { portfolioApi } from '../api/portfolio'
import { useUserStore } from './user'

export const usePortfolioStore = defineStore('portfolio', () => {
  const userStore = useUserStore()
  
  // 状态 - 所有数据都来自后端
  const holdings = ref([])
  const trades = ref([])
  const prices = ref({})
  const realizedProfitLoss = ref(0)
  const isLoading = ref(false)
  const error = ref(null)

  // 计算属性 - 组合资产数据（包含实时价格计算）
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

  // 计算属性 - 总资产价值（基于实时价格）
  const totalValue = computed(() => {
    return portfolio.value.reduce((sum, item) => sum + item.value, 0)
  })

  // 计算属性 - USDT余额
  const usdtBalance = computed(() => {
    const usdtHolding = holdings.value.find(h => h.symbol === 'USDT')
    return usdtHolding ? usdtHolding.amount : 0
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
      // 同时返回价格变化数据
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
      // 未登录时清空数据
      holdings.value = []
      trades.value = []
      realizedProfitLoss.value = 0
      return
    }
    
    isLoading.value = true
    error.value = null
    
    try {
      // 并行获取数据
      const [summaryRes, holdingsRes, tradesRes] = await Promise.all([
        portfolioApi.getSummary(),
        portfolioApi.getHoldings(),
        portfolioApi.getTrades()
      ])

      realizedProfitLoss.value = summaryRes.data.realized_profit_loss
      holdings.value = summaryRes.data.holdings || holdingsRes.data.holdings || []
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

  // 导出数据 - 从后端获取完整数据
  async function exportData() {
    if (!userStore.isLoggedIn) {
      return { success: false, error: '请先登录' }
    }

    try {
      // 获取最新数据
      const [summaryRes, holdingsRes, tradesRes] = await Promise.all([
        portfolioApi.getSummary(),
        portfolioApi.getHoldings(),
        portfolioApi.getTrades({ page: 1, page_size: 1000 }) // 获取所有交易
      ])

      const data = {
        portfolio: holdingsRes.data.holdings || [],
        trades: tradesRes.data.trades || [],
        realizedProfitLoss: summaryRes.data.realized_profit_loss || 0,
        totalValue: summaryRes.data.total_value || 0,
        usdtBalance: summaryRes.data.usdt_balance || 0
      }

      return { success: true, data }
    } catch (err) {
      return { success: false, error: err.response?.data?.error || '导出失败' }
    }
  }

  // 导入数据 - 覆盖后端数据
  async function importData(data) {
    if (!userStore.isLoggedIn) {
      return { success: false, error: '请先登录' }
    }

    isLoading.value = true
    error.value = null

    try {
      // 调用后端导入接口
      const response = await portfolioApi.importData({
        portfolio: data.portfolio || [],
        trades: data.trades || [],
        realized_profit_loss: data.realizedProfitLoss || 0
      })

      // 刷新数据
      await fetchPortfolio()

      return { success: true, data: response.data }
    } catch (err) {
      error.value = err.response?.data?.error || '导入失败'
      return { success: false, error: error.value }
    } finally {
      isLoading.value = false
    }
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
    exportData,
    importData
  }
})
