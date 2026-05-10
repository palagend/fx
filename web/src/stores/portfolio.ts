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
  btc_price: number
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

export interface ImportData {
  version: string
  trades: Trade[]
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
    value_change_24h: 0,
    btc_price: 0
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

  function mergeDashboardData(newData: Partial<DashboardData>) {
    Object.assign(dashboardData.value, newData)
  }

  function addTrade(newTrade: Trade) {
    trades.value.unshift(newTrade)
  }

  function removeTrade(id: number) {
    const index = trades.value.findIndex(t => t.id === id)
    if (index !== -1) {
      trades.value.splice(index, 1)
    }
  }

  async function fetchDashboard(options: { useCache?: boolean; silent?: boolean } = {}): Promise<TradeResult & { updatedAt?: number }> {
    if (config.isBackend && !userStore.isLoggedIn) {
      resetDashboardData()
      return { success: false, error: '请先登录' }
    }

    const { useCache = true, silent = false } = options

    if (!silent) {
      isLoading.value = true
    }
    error.value = null

    try {
      const [dashboardRes, tradesRes] = await Promise.all([
        portfolioApi.getDashboard(useCache),
        portfolioApi.getTrades()
      ])

      const data = dashboardRes.data as unknown as Record<string, unknown>
      const d = data as {
        prices?: Record<string, number>
        us_stock_prices?: Record<string, number>
        price_changes?: Record<string, number>
        exchange_rates?: Record<string, number>
        portfolio?: Asset[]
        crypto_value?: number
        us_stock_value?: number
        cash_balance?: number
        unrealized_profit_loss?: number
        realized_profit_loss?: number
        unrealized_profit_loss_rate?: number
        realized_profit_loss_rate?: number
        value_change_24h?: number
        btc_price?: number
        updated_at?: number
      }
      mergeDashboardData({
        prices: d.prices || {},
        us_stock_prices: d.us_stock_prices || {},
        price_changes: d.price_changes || {},
        exchange_rates: d.exchange_rates || {},
        portfolio: d.portfolio || [],
        crypto_value: d.crypto_value || 0,
        us_stock_value: d.us_stock_value || 0,
        cash_balance: d.cash_balance || 0,
        unrealized_profit_loss: d.unrealized_profit_loss || 0,
        realized_profit_loss: d.realized_profit_loss || 0,
        unrealized_profit_loss_rate: d.unrealized_profit_loss_rate || 0,
        realized_profit_loss_rate: d.realized_profit_loss_rate || 0,
        value_change_24h: d.value_change_24h || 0,
        btc_price: d.btc_price || 0,
        updated_at: d.updated_at || Date.now()
      })

      const tradesData = tradesRes.data as unknown as { trades?: Trade[] }
      const newTrades = tradesData.trades || []
      reconcileTrades(newTrades)

      return {
        success: true,
        updatedAt: (data.updated_at as number) || Date.now()
      }
    } catch (err) {
      const e = err as { response?: { data?: { error?: string } } }
      const errorMsg = e.response?.data?.error || '获取数据失败'
      error.value = errorMsg
      console.error('获取仪表盘数据失败:', err)
      return { success: false, error: errorMsg }
    } finally {
      if (!silent) {
        isLoading.value = false
      }
    }
  }

  // 记录上次后台刷新的时间
  let lastBackgroundRefresh = 0
  const BACKGROUND_REFRESH_INTERVAL = 2 * 60 * 1000 // 2分钟内不重复后台刷新

  /**
   * 分阶段加载：先使用缓存快速显示，再后台刷新最新数据
   * 优化：增加刷新间隔控制，避免频繁调用 API
   */
  async function fetchDashboardStaged(): Promise<void> {
    // 第一阶段：使用缓存快速显示（如果有缓存）
    const result = await fetchDashboard({ useCache: true, silent: true })

    // 第二阶段：后台刷新最新数据（仅在缓存较旧时）
    const now = Date.now()
    const shouldRefresh = now - lastBackgroundRefresh > BACKGROUND_REFRESH_INTERVAL

    if (shouldRefresh && result.success) {
      lastBackgroundRefresh = now
      // 延迟后台刷新，避免阻塞页面渲染
      setTimeout(async () => {
        await fetchDashboard({ useCache: false, silent: true })
      }, 500)
    }
  }

  function reconcileTrades(newTrades: Trade[]) {
    const existingIds = new Set(trades.value.map(t => t.id))
    const newIds = new Set(newTrades.map(t => t.id))

    const toAdd = newTrades.filter(t => !existingIds.has(t.id))
    const toRemove = trades.value.filter(t => !newIds.has(t.id))

    toRemove.forEach(t => removeTrade(t.id))
    toAdd.forEach(t => addTrade(t))

    trades.value.sort((a, b) => {
      const aTime = a.created_at ? new Date(a.created_at).getTime() : (a.timestamp || 0)
      const bTime = b.created_at ? new Date(b.created_at).getTime() : (b.timestamp || 0)
      return bTime - aTime
    })
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
      value_change_24h: 0,
      btc_price: 0
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
      const updatedAt = new Date(response.data.updated_at as string).getTime()
      return {
        success: true,
        price: response.data.price,
        updatedAt
      }
    } catch (err) {
      const e = err as { response?: { data?: { error?: string } } }
      console.error(`获取${symbol}价格失败:`, err)
      return { success: false, error: e.response?.data?.error || '获取价格失败' }
    }
  }

  async function _createTrade(trade: CreateTradeParams): Promise<TradeResult> {
    isLoading.value = true
    error.value = null

    try {
      const response = await portfolioApi.createTrade(trade)
      return { success: true, data: response.data }
    } catch (err) {
      const e = err as { response?: { data?: { error?: string } } }
      const errorMsg = e.response?.data?.error || (err as Error).message || '交易失败'
      error.value = errorMsg
      return { success: false, error: errorMsg }
    } finally {
      isLoading.value = false
    }
  }

  const createTrade = async (trade: CreateTradeParams, options: { refresh?: boolean } = {}): Promise<TradeResult> => {
    const result = await _createTrade(trade)
    if (result.success && options.refresh !== false) {
      await fetchDashboard()
    }
    return result
  }

  async function _deleteTrade(id: number): Promise<TradeResult> {
    try {
      await portfolioApi.deleteTrade(id)
      return { success: true }
    } catch (err) {
      const e = err as { response?: { data?: { error?: string } } }
      return { success: false, error: e.response?.data?.error || '删除失败' }
    }
  }

  const deleteTrade = async (id: number, options: { refresh?: boolean } = {}): Promise<TradeResult> => {
    const result = await _deleteTrade(id)
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
      const e = err as { response?: { data?: { error?: string } } }
      const errorMsg = e.response?.data?.error || '清空交易记录失败'
      error.value = errorMsg
      return { success: false, error: errorMsg }
    } finally {
      isLoading.value = false
    }
  }

  const clearAllTrades = async (options: { refresh?: boolean } = {}): Promise<TradeResult> => {
    const result = await _clearAllTrades()
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
      const e = err as { response?: { data?: { error?: string } } }
      const errorMsg = e.response?.data?.error || '导出数据失败'
      error.value = errorMsg
      return { success: false, error: errorMsg }
    }
  }

  async function _importPreview(data: ImportData): Promise<TradeResult & { preview?: unknown }> {
    error.value = null

    try {
      const localData = data as unknown as { version: string; trades: { id: number; uuid: string; asset_type: 'crypto' | 'us_stock' | 'cash'; symbol: string; type: 'buy' | 'sell' | 'recharge'; amount: number; price: number; total: number; currency: string; created_at: string }[] }
      const response = await portfolioApi.importPreview(localData)
      return { success: true, preview: response.data.preview }
    } catch (err) {
      const e = err as { response?: { data?: { error?: string } } }
      const errorMsg = e.response?.data?.error || '预览导入数据失败'
      error.value = errorMsg
      return { success: false, error: errorMsg }
    }
  }

  async function _importConfirm(data: ImportData, conflictStrategy: string = 'skip'): Promise<TradeResult & { imported?: number; skipped?: number; overwritten?: number }> {
    isLoading.value = true
    error.value = null

    try {
      const localData = data as unknown as { version: string; trades: { id: number; uuid: string; asset_type: 'crypto' | 'us_stock' | 'cash'; symbol: string; type: 'buy' | 'sell' | 'recharge'; amount: number; price: number; total: number; currency: string; created_at: string }[] }
      const strategy = conflictStrategy === 'overwrite' ? 'overwrite' : 'skip'
      const response = await portfolioApi.importConfirm(localData, strategy)
      return {
        success: true,
        imported: response.data.imported,
        skipped: response.data.skipped,
        overwritten: response.data.overwritten
      }
    } catch (err) {
      const e = err as { response?: { data?: { error?: string } } }
      const errorMsg = e.response?.data?.error || '导入数据失败'
      error.value = errorMsg
      return { success: false, error: errorMsg }
    } finally {
      isLoading.value = false
    }
  }

  const exportData = () => _exportData()
  const importPreview = (data: ImportData) => _importPreview(data)
  const importConfirm = (data: ImportData, conflictStrategy: string) => _importConfirm(data, conflictStrategy)

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
    fetchDashboardStaged,
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