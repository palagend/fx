import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { portfolioApi, config } from '../api'
import { useUserStore } from './user'
import type { Asset, Trade } from '../types'

export interface DashboardData {
  prices: Record<string, number>
  us_stock_prices: Record<string, number>
  price_changes: Record<string, number>
  exchange_rates: Record<string, number>
  portfolio: Asset[]
  crypto_value: number
  us_stock_value: number
  cash_balance: number
  unrealized_profit_loss: number
  realized_profit_loss: number
  unrealized_profit_loss_rate: number
  realized_profit_loss_rate: number
  value_change_24h: number
  updated_at?: number
}

export interface CreateTradeParams {
  symbol: string
  asset_type: 'crypto' | 'us_stock'
  type: 'buy' | 'sell'
  amount: number
  price: number
  fee?: number
}

export interface TradeResult {
  success: boolean
  error?: string
  data?: unknown
}

export const usePortfolioStore = defineStore('portfolio', () => {
  const userStore = useUserStore()

  const dashboardData = ref<DashboardData>({
    prices: {},
    us_stock_prices: {},
    price_changes: {},
    exchange_rates: {},
    portfolio: [],
    crypto_value: 0,
    us_stock_value: 0,
    cash_balance: 0,
    unrealized_profit_loss: 0,
    realized_profit_loss: 0,
    unrealized_profit_loss_rate: 0,
    realized_profit_loss_rate: 0,
    value_change_24h: 0
  })
  const trades = ref<Trade[]>([])
  const isLoading = ref<boolean>(false)
  const error = ref<string | null>(null)

  const prices = computed(() => dashboardData.value.prices)
  const usStockPrices = computed(() => dashboardData.value.us_stock_prices)
  const priceChanges = computed(() => dashboardData.value.price_changes)
  const exchangeRates = computed(() => dashboardData.value.exchange_rates)
  const portfolio = computed(() => dashboardData.value.portfolio)

  const cryptoAssetsValue = computed(() => dashboardData.value.crypto_value)
  const usStockValue = computed(() => dashboardData.value.us_stock_value)
  const cashBalance = computed(() => dashboardData.value.cash_balance)
  const totalAssetsValue = computed(() => 
    dashboardData.value.crypto_value + 
    dashboardData.value.us_stock_value + 
    dashboardData.value.cash_balance
  )

  const unrealizedPL = computed(() => dashboardData.value.unrealized_profit_loss)
  const realizedPL = computed(() => dashboardData.value.realized_profit_loss)
  const totalPL = computed(() => 
    dashboardData.value.unrealized_profit_loss + 
    dashboardData.value.realized_profit_loss
  )

  const unrealizedPLRate = computed(() => dashboardData.value.unrealized_profit_loss_rate)
  const realizedPLRate = computed(() => dashboardData.value.realized_profit_loss_rate)
  const cryptoValueChange24h = computed(() => dashboardData.value.value_change_24h)

  const requireAuth = <T extends (...args: unknown[]) => Promise<TradeResult>>(fn: T) => async (...args: Parameters<T>): Promise<TradeResult> => {
    if (!userStore.isLoggedIn) {
      return { success: false, error: '请先登录' }
    }
    return fn(...args)
  }

  function mergeDashboardData(newData: DashboardData) {
    Object.assign(dashboardData.value, newData)
  }

  function mergePrices(newPrices: Record<string, number>) {
    Object.assign(dashboardData.value.prices, newPrices)
  }

  function mergeUsStockPrices(newPrices: Record<string, number>) {
    Object.assign(dashboardData.value.us_stock_prices, newPrices)
  }

  function updatePortfolioItem(index: number, updates: Partial<Asset>) {
    if (dashboardData.value.portfolio[index]) {
      Object.assign(dashboardData.value.portfolio[index], updates)
    }
  }

  function addTrade(newTrade: Trade) {
    trades.value.unshift(newTrade)
  }

  function removeTrade(id: string) {
    const index = trades.value.findIndex(t => t.id === id)
    if (index !== -1) {
      trades.value.splice(index, 1)
    }
  }

  async function fetchDashboard(): Promise<TradeResult & { updatedAt?: number }> {
    if (config.isBackend && !userStore.isLoggedIn) {
      resetDashboardData()
      return { success: false, error: '请先登录' }
    }

    isLoading.value = true
    error.value = null

    try {
      const [dashboardRes, tradesRes] = await Promise.all([
        portfolioApi.getDashboard(),
        portfolioApi.getTrades()
      ])

      mergeDashboardData(dashboardRes.data)

      const newTrades = tradesRes.data.trades || []
      reconcileTrades(newTrades)

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

  function reconcileTrades(newTrades: Trade[]) {
    const existingIds = new Set(trades.value.map(t => t.id))
    const newIds = new Set(newTrades.map(t => t.id))

    const toAdd = newTrades.filter(t => !existingIds.has(t.id))
    const toRemove = trades.value.filter(t => !newIds.has(t.id))

    toRemove.forEach(t => removeTrade(t.id))
    toAdd.forEach(t => addTrade(t))

    trades.value.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
  }

  function resetDashboardData() {
    dashboardData.value = {
      prices: {},
      us_stock_prices: {},
      price_changes: {},
      exchange_rates: {},
      portfolio: [],
      crypto_value: 0,
      us_stock_value: 0,
      cash_balance: 0,
      unrealized_profit_loss: 0,
      realized_profit_loss: 0,
      unrealized_profit_loss_rate: 0,
      realized_profit_loss_rate: 0,
      value_change_24h: 0
    }
    trades.value = []
  }

  async function fetchAssetPrice(symbol: string, assetType: 'crypto' | 'us_stock' = 'crypto'): Promise<TradeResult & { price?: number; updatedAt?: number }> {
    try {
      const response = await portfolioApi.getAssetPrice(symbol, assetType)
      if (assetType === 'crypto') {
        dashboardData.value.prices = {
          ...dashboardData.value.prices,
          [symbol]: response.data.price
        }
      } else if (assetType === 'us_stock') {
        dashboardData.value.us_stock_prices = {
          ...dashboardData.value.us_stock_prices,
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

  async function _createTrade(trade: CreateTradeParams): Promise<TradeResult> {
    isLoading.value = true
    error.value = null

    try {
      const response = await portfolioApi.createTrade(trade)
      return { success: true, data: response.data }
    } catch (err) {
      error.value = err.response?.data?.error || err.message || '交易失败'
      return { success: false, error: error.value }
    } finally {
      isLoading.value = false
    }
  }

  const createTrade = async (trade: CreateTradeParams, options: { refresh?: boolean } = {}): Promise<TradeResult> => {
    const result = await requireAuth(_createTrade)(trade)
    if (result.success && options.refresh !== false) {
      await fetchDashboard()
    }
    return result
  }

  async function _deleteTrade(id: string): Promise<TradeResult> {
    try {
      await portfolioApi.deleteTrade(id)
      return { success: true }
    } catch (err) {
      return { success: false, error: err.response?.data?.error || '删除失败' }
    }
  }

  const deleteTrade = async (id: string, options: { refresh?: boolean } = {}): Promise<TradeResult> => {
    const result = await requireAuth(_deleteTrade)(id)
    if (result.success) {
      if (options.refresh !== false) {
        await fetchDashboard()
      } else {
        removeTrade(id)
      }
    }
    return result
  }

  async function _clearAllTrades(): Promise<TradeResult> {
    isLoading.value = true
    error.value = null

    try {
      await portfolioApi.clearTrades()
      return { success: true }
    } catch (err) {
      error.value = err.response?.data?.error || '清空交易记录失败'
      return { success: false, error: error.value }
    } finally {
      isLoading.value = false
    }
  }

  const clearAllTrades = async (options: { refresh?: boolean } = {}): Promise<TradeResult> => {
    const result = await requireAuth(_clearAllTrades)()
    if (result.success) {
      if (options.refresh !== false) {
        await fetchDashboard()
      } else {
        resetDashboardData()
      }
    }
    return result
  }

  async function _exportData(): Promise<TradeResult & { data?: unknown }> {
    error.value = null

    try {
      const response = await portfolioApi.exportData()
      return { success: true, data: response.data.data }
    } catch (err) {
      error.value = err.response?.data?.error || '导出数据失败'
      return { success: false, error: error.value }
    }
  }

  async function _importPreview(data: unknown): Promise<TradeResult & { preview?: unknown }> {
    error.value = null

    try {
      const response = await portfolioApi.importPreview(data)
      return { success: true, preview: response.data.preview }
    } catch (err) {
      error.value = err.response?.data?.error || '预览导入数据失败'
      return { success: false, error: error.value }
    }
  }

  async function _importConfirm(data: unknown, conflictStrategy: string = 'skip'): Promise<TradeResult & { imported?: number; skipped?: number; overwritten?: number }> {
    isLoading.value = true
    error.value = null

    try {
      const response = await portfolioApi.importConfirm(data, conflictStrategy)
      return {
        success: true,
        imported: response.data.imported,
        skipped: response.data.skipped,
        overwritten: response.data.overwritten
      }
    } catch (err) {
      error.value = err.response?.data?.error || '导入数据失败'
      return { success: false, error: error.value }
    } finally {
      isLoading.value = false
    }
  }

  const exportData = () => requireAuth(_exportData)()
  const importPreview = (data: unknown) => requireAuth(_importPreview)(data)
  const importConfirm = (data: unknown, conflictStrategy: string) => requireAuth(_importConfirm)(data, conflictStrategy)

  return {
    dashboardData,
    trades,
    isLoading,
    error,
    prices,
    usStockPrices,
    priceChanges,
    exchangeRates,
    portfolio,
    cryptoAssetsValue,
    usStockValue,
    cashBalance,
    totalAssetsValue,
    unrealizedPL,
    realizedPL,
    totalPL,
    unrealizedPLRate,
    realizedPLRate,
    cryptoValueChange24h,
    fetchDashboard,
    fetchAssetPrice,
    createTrade,
    deleteTrade,
    clearAllTrades,
    exportData,
    importPreview,
    importConfirm,
    addTrade,
    removeTrade
  }
})