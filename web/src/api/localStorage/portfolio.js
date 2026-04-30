// LocalStorage 版本的 Portfolio API（纯前端模式）
// 参考后端 portfolio.go 的实现逻辑

import axios from 'axios'

const STORAGE_KEYS = {
  TRADES: 'portfolio_trades',
  HOLDINGS: 'portfolio_holdings',
  PRICES: 'portfolio_prices',
  PRICE_CHANGES: 'portfolio_price_changes',
  US_STOCK_PRICES: 'us_stock_prices',
  PRICE_UPDATED_AT: 'portfolio_price_updated_at'
}

// CoinCap API 配置
const COINCAP_API_KEY = 'b617d9cf029dbb40f02b058a0e74919176b768cf36fd1ea6fae55a13a1610f41'
const COINCAP_BASE_URL = 'https://rest.coincap.io/v3'

// 腾讯财经API配置（股票价格）
const TENCENT_STOCK_URL = 'https://qt.gtimg.cn'

// 支持的加密货币列表（包含USDT）
const supportedCryptos = {
  'BTC': true,
  'ETH': true,
  'BNB': true,
  'XRP': true,
  'ADA': true,
  'SOL': true,
  'DOGE': true,
  'TRX': true,
  'AVAX': true,
  'HYPE': true,
  'USDT': true
}

// 支持的美股列表
const supportedUSStocks = {
  'AAPL': true,
  'MSFT': true,
  'GOOG': true,
  'AMZN': true,
  'TSLA': true,
  'META': true,
  'NVDA': true,
  'BABA': true,
  'ORCL': true,
  'CRCL': true,
  'MSTR': true
}

// CoinCap ID 映射
const symbolToCoinCapId = {
  'BTC': 'bitcoin',
  'ETH': 'ethereum',
  'BNB': 'binance-coin',
  'XRP': 'xrp',
  'ADA': 'cardano',
  'SOL': 'solana',
  'DOGE': 'dogecoin',
  'TRX': 'tron',
  'AVAX': 'avalanche',
  'HYPE': 'hyperliquid',
  'USDT': 'tether'
}

// ========== 辅助函数 ==========

function getStorageData(key, defaultValue = []) {
  try {
    const data = localStorage.getItem(key)
    return data ? JSON.parse(data) : defaultValue
  } catch {
    return defaultValue
  }
}

function setStorageData(key, value) {
  localStorage.setItem(key, JSON.stringify(value))
}

function generateId() {
  return Date.now().toString(36) + Math.random().toString(36).substr(2)
}

function abs(x) {
  return x < 0 ? -x : x
}

function getCurrencyByAssetType(assetType) {
  switch (assetType) {
    case 'crypto':
      return 'USDT'
    case 'us_stock':
      return 'USD'
    case 'cash':
      return 'USD'
    default:
      return 'USD'
  }
}

// ========== 交易校验 ==========

function validateTradeRequest(req) {
  if (!req.asset_type) {
    throw new Error('资产类型不能为空')
  }
  if (!req.symbol) {
    throw new Error('币种代码不能为空')
  }
  if (!req.type) {
    throw new Error('交易类型不能为空')
  }
  if (req.amount === undefined || req.amount === null) {
    throw new Error('交易数量不能为空')
  }
  if (req.price === undefined || req.price === null) {
    throw new Error('交易价格不能为空')
  }

  if (!['crypto', 'us_stock', 'cash'].includes(req.asset_type)) {
    throw new Error(`无效的资产类型: ${req.asset_type}`)
  }
  if (!['buy', 'sell', 'recharge'].includes(req.type)) {
    throw new Error(`无效的交易类型: ${req.type}`)
  }
  if (req.amount <= 0) {
    throw new Error('交易数量必须大于0')
  }
  if (req.price <= 0) {
    throw new Error('交易价格必须大于0')
  }

  switch (req.type) {
    case 'recharge':
      if (req.asset_type !== 'cash') {
        throw new Error('充值资产类型必须是cash')
      }
      if (req.symbol !== 'USD') {
        throw new Error('充值只支持USD')
      }
      if (req.price !== 1) {
        throw new Error('充值价格必须为1')
      }
      break
    case 'buy':
    case 'sell':
      switch (req.asset_type) {
        case 'crypto':
          if (!supportedCryptos[req.symbol]) {
            throw new Error(`不支持的加密货币: ${req.symbol}`)
          }
          break
        case 'us_stock':
          if (!supportedUSStocks[req.symbol]) {
            throw new Error(`不支持的美股: ${req.symbol}`)
          }
          break
        default:
          throw new Error(`买卖交易不支持资产类型: ${req.asset_type}`)
      }
      break
  }
  return null
}

// ========== 持仓管理 ==========

function getOrCreateHolding(holdings, symbol, assetType) {
  const key = `${assetType}:${symbol}`
  const existing = holdings.find(h => h.asset_type === assetType && h.symbol === symbol)
  if (existing) {
    return existing
  }
  const newHolding = {
    id: generateId(),
    asset_type: assetType,
    symbol: symbol,
    amount: 0,
    currency: getCurrencyByAssetType(assetType)
  }
  holdings.push(newHolding)
  return newHolding
}

function updateHolding(holdings, holding, delta) {
  holding.amount += delta
  return holding
}

function recalcAllHoldings(trades) {
  const holdings = []
  const cashHolding = {
    id: generateId(),
    asset_type: 'cash',
    symbol: 'USD',
    amount: 0,
    currency: 'USD'
  }

  const sortedTrades = [...trades].sort((a, b) => new Date(a.created_at) - new Date(b.created_at))

  for (const t of sortedTrades) {
    if (t.type === 'recharge') {
      cashHolding.amount += t.amount
      continue
    }

    const holding = getOrCreateHolding(holdings, t.symbol, t.asset_type)

    switch (t.type) {
      case 'buy':
        holding.amount += t.amount
        cashHolding.amount -= t.total
        break
      case 'sell':
        holding.amount -= t.amount
        cashHolding.amount += t.total
        break
    }
  }

  const validHoldings = holdings.filter(h => h.amount !== 0)
  if (cashHolding.amount !== 0) {
    validHoldings.unshift(cashHolding)
  }

  setStorageData(STORAGE_KEYS.HOLDINGS, validHoldings)
  return validHoldings
}

// ========== 价格获取 ==========

async function fetchCryptoPrices() {
  const ids = Object.values(symbolToCoinCapId).join(',')
  const url = `${COINCAP_BASE_URL}/assets?ids=${ids}`

  try {
    const response = await axios.get(url, {
      headers: {
        'Authorization': `Bearer ${COINCAP_API_KEY}`
      },
      timeout: 10000
    })

    const prices = { 'USDT': 1.0 }
    const priceChanges = { 'USDT': 0 }
    let updatedAt = Date.now()

    if (response.data && response.data.data) {
      for (const item of response.data.data) {
        const price = parseFloat(item.priceUsd) || 0
        const change24hPercent = parseFloat(item.changePercent24Hr) || 0
        prices[item.symbol] = price
        priceChanges[item.symbol] = change24hPercent / 100
      }
      updatedAt = response.data.timestamp || Date.now()
    }

    setStorageData(STORAGE_KEYS.PRICES, prices)
    setStorageData(STORAGE_KEYS.PRICE_CHANGES, priceChanges)
    setStorageData(STORAGE_KEYS.PRICE_UPDATED_AT, updatedAt)

    return { prices, priceChanges, updatedAt }
  } catch (error) {
    console.error('获取 CoinCap 价格失败:', error)
    const cachedPrices = getStorageData(STORAGE_KEYS.PRICES, { 'USDT': 1.0 })
    const cachedChanges = getStorageData(STORAGE_KEYS.PRICE_CHANGES, { 'USDT': 0 })
    const cachedUpdatedAt = getStorageData(STORAGE_KEYS.PRICE_UPDATED_AT, Date.now())
    return { prices: cachedPrices, priceChanges: cachedChanges, updatedAt: cachedUpdatedAt }
  }
}

async function fetchUSStockPrice(symbol) {
  const url = `${TENCENT_STOCK_URL}/q=us${symbol}`

  try {
    const response = await axios.get(url, { timeout: 10000 })
    const text = response.data

    if (!text || !text.includes('v_us')) {
      throw new Error('股票价格数据为空')
    }

    const pattern = new RegExp(`v_us${symbol}="([^"]+)"`, 'i')
    const match = text.match(pattern)
    if (match && match[1]) {
      const data = match[1].split('~')
      if (data.length > 4) {
        return {
          price: parseFloat(data[3]) || 0,
          updated_at: Date.now()
        }
      }
    }

    throw new Error('解析股票价格失败')
  } catch (error) {
    console.error(`获取 ${symbol} 股票价格失败:`, error)
    const cachedPrices = getStorageData(STORAGE_KEYS.US_STOCK_PRICES, {})
    return { price: cachedPrices[symbol] || 0, updated_at: Date.now() }
  }
}

async function fetchUSStockPricesBatch() {
  const symbols = Object.keys(supportedUSStocks)
  const stockCodes = symbols.map(s => `us${s}`).join(',')
  const url = `${TENCENT_STOCK_URL}/q=${stockCodes}`

  try {
    const response = await axios.get(url, { timeout: 10000 })
    const text = response.data
    const prices = {}

    symbols.forEach(symbol => {
      const pattern = new RegExp(`v_us${symbol}="([^"]+)"`, 'i')
      const match = text.match(pattern)
      if (match && match[1]) {
        const data = match[1].split('~')
        if (data.length > 4) {
          prices[symbol] = parseFloat(data[3]) || 0
        }
      }
    })

    setStorageData(STORAGE_KEYS.US_STOCK_PRICES, prices)
    return prices
  } catch (error) {
    console.error('批量获取美股价格失败:', error)
    return getStorageData(STORAGE_KEYS.US_STOCK_PRICES, {})
  }
}

async function fetchAssetPrice(symbol, assetType = 'crypto') {
  if (!symbol) {
    throw new Error('币种代码不能为空')
  }

  switch (assetType) {
    case 'crypto':
      if (symbol === 'USDT') {
        return { price: 1.0, updated_at: Date.now() }
      }
      if (!supportedCryptos[symbol]) {
        throw new Error('不支持的加密货币')
      }

      try {
        const url = `${COINCAP_BASE_URL}/price/bysymbol/${symbol}`
        const response = await axios.get(url, {
          headers: { 'Authorization': `Bearer ${COINCAP_API_KEY}` },
          timeout: 10000
        })

        if (response.data && response.data.data && response.data.data.length > 0) {
          const price = parseFloat(response.data.data[0]) || 0
          return { price, updated_at: response.data.timestamp || Date.now() }
        }
        throw new Error('价格数据为空')
      } catch (error) {
        console.error(`获取 ${symbol} 价格失败:`, error)
        const cachedPrices = getStorageData(STORAGE_KEYS.PRICES, {})
        return { price: cachedPrices[symbol] || 0, updated_at: Date.now() }
      }

    case 'us_stock':
      return fetchUSStockPrice(symbol)

    default:
      throw new Error(`不支持的资产类型: ${assetType}`)
  }
}

// ========== 投资组合统计计算 ==========

function calculatePortfolioStats(holdings, cryptoPrices, cryptoChanges, usStockPrices, trades) {
  const portfolio = []
  let cryptoValue = 0
  let usStockValue = 0
  let cashBalance = 0
  let totalUnrealizedPL = 0
  let totalRealizedPL = 0
  let totalHistoricalCost = 0

  const assetData = {}

  for (const t of trades) {
    if (t.type === 'recharge') continue

    if (!assetData[t.asset_type]) {
      assetData[t.asset_type] = {}
    }
    if (!assetData[t.asset_type][t.symbol]) {
      assetData[t.asset_type][t.symbol] = {
        amount: 0,
        cost: 0,
        totalIn: 0,
        realizedPL: 0
      }
    }

    const d = assetData[t.asset_type][t.symbol]

    switch (t.type) {
      case 'buy':
        d.amount += t.amount
        d.cost += t.total
        d.totalIn += t.total
        break
      case 'sell':
        if (d.amount > 0 && t.amount > 0) {
          const sellRatio = t.amount / d.amount
          const costRecovered = d.cost * sellRatio
          const realizedPL = t.total - costRecovered

          d.realizedPL += realizedPL
          d.cost -= costRecovered
          d.amount -= t.amount
        }
        break
    }

    assetData[t.asset_type][t.symbol] = d
  }

  for (const assetType in assetData) {
    for (const symbol in assetData[assetType]) {
      if (symbol !== 'USD') {
        totalRealizedPL += assetData[assetType][symbol].realizedPL
        totalHistoricalCost += assetData[assetType][symbol].totalIn
      }
    }
  }

  for (const h of holdings) {
    if (h.asset_type === 'cash' && h.symbol === 'USD') {
      cashBalance = h.amount
      continue
    }

    let price = 0
    switch (h.asset_type) {
      case 'crypto':
        price = cryptoPrices[h.symbol] || 0
        break
      case 'us_stock':
        price = usStockPrices[h.symbol] || 0
        break
    }

    const marketValue = h.amount * price

    const d = assetData[h.asset_type]?.[h.symbol] || { amount: 0, cost: 0, totalIn: 0, realizedPL: 0 }
    const cost = d.cost
    const realizedPL = d.realizedPL

    switch (h.asset_type) {
      case 'crypto':
        cryptoValue += marketValue
        break
      case 'us_stock':
        usStockValue += marketValue
        break
    }

    const avgCost = h.amount !== 0 ? cost / h.amount : 0
    const profitLoss = marketValue - cost
    const plRate = cost !== 0 ? (profitLoss / cost) * 100 : 0

    totalUnrealizedPL += profitLoss

    const realizedPLRate = d.totalIn !== 0 ? (realizedPL / d.totalIn) * 100 : 0

    portfolio.push({
      asset_type: h.asset_type,
      symbol: h.symbol,
      amount: h.amount,
      current_price: price,
      avg_cost: avgCost,
      market_value: marketValue,
      cost: cost,
      profit_loss: profitLoss,
      pl_rate: plRate,
      realized_pl: realizedPL,
      realized_pl_rate: realizedPLRate,
      currency: h.currency
    })
  }

  const totalAssetsValue = cryptoValue + usStockValue + cashBalance

  let totalCost = 0
  for (const h of holdings) {
    if (h.asset_type !== 'cash') {
      const d = assetData[h.asset_type]?.[h.symbol]
      if (d) totalCost += d.cost
    }
  }

  const totalUnrealizedPLRate = totalCost !== 0 ? (totalUnrealizedPL / totalCost) * 100 : 0
  const totalRealizedPLRate = totalHistoricalCost !== 0 ? (totalRealizedPL / totalHistoricalCost) * 100 : 0

  return {
    portfolio,
    cryptoValue,
    usStockValue,
    cashBalance,
    totalAssetsValue,
    totalUnrealizedPL,
    totalUnrealizedPLRate,
    totalRealizedPL,
    totalRealizedPLRate,
    totalValueChange24h: 0
  }
}

// ========== API 响应格式 ==========

function mockResponse(data) {
  return Promise.resolve({ data })
}

// ========== 导出 API ==========

export const localPortfolioApi = {
  async getDashboard() {
    const trades = getStorageData(STORAGE_KEYS.TRADES)
    const [cryptoResult, usStockPrices] = await Promise.all([
      fetchCryptoPrices(),
      fetchUSStockPricesBatch()
    ])

    const { prices, priceChanges, updatedAt } = cryptoResult
    const holdings = recalcAllHoldings(trades)
    const stats = calculatePortfolioStats(holdings, prices, priceChanges, usStockPrices, trades)

    return mockResponse({
      prices: prices,
      us_stock_prices: usStockPrices,
      price_changes: priceChanges,
      crypto_updated_at: updatedAt,
      btc_price: prices['BTC'] || 0,
      portfolio: stats.portfolio,
      crypto_value: stats.cryptoValue,
      us_stock_value: stats.usStockValue,
      cash_balance: stats.cashBalance,
      total_assets_value: stats.totalAssetsValue,
      unrealized_profit_loss: stats.totalUnrealizedPL,
      unrealized_profit_loss_rate: stats.totalUnrealizedPLRate,
      realized_profit_loss: stats.totalRealizedPL,
      realized_profit_loss_rate: stats.totalRealizedPLRate,
      total_profit_loss: stats.totalUnrealizedPL + stats.totalRealizedPL,
      value_change_24h: stats.totalValueChange24h
    })
  },

  getTrades() {
    const trades = getStorageData(STORAGE_KEYS.TRADES)
    const sortedTrades = [...trades].sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
    return mockResponse({ trades: sortedTrades })
  },

  createTrade(req) {
    validateTradeRequest(req)

    const trades = getStorageData(STORAGE_KEYS.TRADES)
    const total = req.amount * req.price
    const holdings = getStorageData(STORAGE_KEYS.HOLDINGS, [])

    let cashHolding = holdings.find(h => h.asset_type === 'cash' && h.symbol === 'USD')
    if (!cashHolding) {
      cashHolding = { asset_type: 'cash', symbol: 'USD', amount: 0 }
    }

    switch (req.type) {
      case 'recharge':
        cashHolding.amount += req.amount
        break
      case 'buy':
        if (cashHolding.amount < total) {
          throw new Error('USD现金余额不足')
        }
        cashHolding.amount -= total
        break
      case 'sell':
        const assetHolding = holdings.find(h => h.asset_type === req.asset_type && h.symbol === req.symbol)
        if (!assetHolding || assetHolding.amount < req.amount) {
          throw new Error('持仓不足')
        }
        cashHolding.amount += total
        break
    }

    const newTrade = {
      id: generateId(),
      asset_type: req.asset_type,
      symbol: req.symbol,
      type: req.type,
      amount: req.amount,
      price: req.price,
      total: total,
      currency: getCurrencyByAssetType(req.asset_type),
      created_at: new Date().toISOString()
    }

    trades.push(newTrade)
    setStorageData(STORAGE_KEYS.TRADES, trades)
    recalcAllHoldings(trades)

    return mockResponse({
      id: newTrade.id,
      asset_type: newTrade.asset_type,
      symbol: newTrade.symbol,
      type: newTrade.type,
      amount: newTrade.amount,
      price: newTrade.price,
      total: newTrade.total,
      currency: newTrade.currency,
      created_at: newTrade.created_at
    })
  },

  deleteTrade(id) {
    const trades = getStorageData(STORAGE_KEYS.TRADES)
    const tradeIndex = trades.findIndex(t => t.id === id)

    if (tradeIndex === -1) {
      throw new Error('交易记录不存在')
    }

    const trade = trades[tradeIndex]
    const tradeTime = new Date(trade.created_at)
    const now = new Date()
    if (now - tradeTime > 24 * 60 * 60 * 1000) {
      throw new Error('只能删除24小时内的交易记录')
    }

    const remainingTrades = trades.filter(t => t.id !== id)
    const simulatedHoldings = recalcAllHoldings([...remainingTrades])

    const cashBalance = simulatedHoldings.find(h => h.asset_type === 'cash' && h.symbol === 'USD')?.amount || 0
    if (cashBalance < 0) {
      throw new Error(`删除该交易会导致 USD 现金余额为负数(${cashBalance.toFixed(2)})，无法删除`)
    }

    for (const h of simulatedHoldings) {
      if (h.amount < 0) {
        throw new Error(`删除该交易会导致 ${h.symbol}(${h.asset_type}) 持仓为负数(${h.amount.toFixed(8)})，无法删除`)
      }
    }

    trades.splice(tradeIndex, 1)
    setStorageData(STORAGE_KEYS.TRADES, trades)
    recalcAllHoldings(trades)

    return mockResponse({
      message: '交易记录已删除',
      deleted_trade: {
        id: trade.id,
        asset_type: trade.asset_type,
        symbol: trade.symbol,
        type: trade.type,
        amount: trade.amount,
        price: trade.price,
        created_at: trade.created_at
      }
    })
  },

  clearTrades() {
    setStorageData(STORAGE_KEYS.TRADES, [])
    setStorageData(STORAGE_KEYS.HOLDINGS, [])
    return mockResponse({ message: '所有数据已清空' })
  },

  async getAssetPrice(symbol, assetType = 'crypto') {
    const result = await fetchAssetPrice(symbol, assetType)
    return mockResponse({
      symbol: symbol,
      price: result.price,
      asset_type: assetType,
      currency: assetType === 'crypto' ? 'USDT' : 'USD',
      updated_at: new Date(result.updated_at).toISOString()
    })
  },

  exportData() {
    const trades = getStorageData(STORAGE_KEYS.TRADES)
    const sortedTrades = [...trades].sort((a, b) => new Date(a.created_at) - new Date(b.created_at))

    const tradeExports = sortedTrades.map(t => ({
      id: t.id,
      asset_type: t.asset_type,
      symbol: t.symbol,
      type: t.type,
      amount: t.amount,
      price: t.price,
      total: t.total,
      currency: t.currency,
      created_at: t.created_at
    }))

    return mockResponse({
      data: {
        version: '1.0',
        exported: new Date().toISOString(),
        trades: tradeExports
      }
    })
  },

  importPreview(data) {
    if (data.version !== '1.0') {
      throw new Error(`不支持的版本: ${data.version}`)
    }

    const existingTrades = getStorageData(STORAGE_KEYS.TRADES)
    const existingMap = new Set()

    for (const t of existingTrades) {
      const key = `${t.asset_type}_${t.symbol}_${t.type}_${t.created_at}`
      existingMap.add(key)
    }

    const preview = {
      total_trades: data.trades.length,
      new_trades: 0,
      conflicts: 0,
      conflict_items: []
    }

    for (const trade of data.trades) {
      const key = `${trade.asset_type}_${trade.symbol}_${trade.type}_${trade.created_at}`
      if (existingMap.has(key)) {
        preview.conflicts++
        preview.conflict_items.push({
          trade: trade,
          reason: '与现有记录时间戳相同'
        })
      } else {
        preview.new_trades++
      }
    }

    return mockResponse({ preview })
  },

  importConfirm(data, conflictStrategy = 'skip') {
    if (data.version !== '1.0') {
      throw new Error(`不支持的版本: ${data.version}`)
    }

    if (conflictStrategy !== 'skip' && conflictStrategy !== 'overwrite') {
      conflictStrategy = 'skip'
    }

    const existingTrades = getStorageData(STORAGE_KEYS.TRADES)
    const existingMap = new Map()

    for (const t of existingTrades) {
      const key = `${t.asset_type}_${t.symbol}_${t.type}_${t.created_at}`
      existingMap.set(key, t.id)
    }

    let imported = 0
    let skipped = 0
    let overwritten = 0

    for (const trade of data.trades) {
      const key = `${trade.asset_type}_${trade.symbol}_${trade.type}_${trade.created_at}`

      if (existingMap.has(key)) {
        if (conflictStrategy === 'overwrite') {
          const existingId = existingMap.get(key)
          const index = existingTrades.findIndex(t => t.id === existingId)
          if (index !== -1) {
            existingTrades.splice(index, 1)
          }
          overwritten++
        } else {
          skipped++
          continue
        }
      }

      const newTrade = {
        id: generateId(),
        asset_type: trade.asset_type,
        symbol: trade.symbol,
        type: trade.type,
        amount: trade.amount,
        price: trade.price,
        total: trade.total,
        currency: trade.currency,
        created_at: trade.created_at
      }

      existingTrades.push(newTrade)
      imported++
    }

    setStorageData(STORAGE_KEYS.TRADES, existingTrades)
    recalcAllHoldings(existingTrades)

    return mockResponse({ imported, skipped, overwritten })
  }
}