// LocalStorage 版本的 Portfolio API（纯前端模式）
// 参考后端 portfolio.go 的实现逻辑

import axios from 'axios'

const STORAGE_KEYS = {
  TRADES: 'portfolio_trades',
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

// 支持的加密货币列表（不含USDT）
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
  'HYPE': true
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
  'HYPE': 'hyperliquid'
}

// ========== 辅助函数 ==========

// 获取存储的数据
function getStorageData(key, defaultValue = []) {
  try {
    const data = localStorage.getItem(key)
    return data ? JSON.parse(data) : defaultValue
  } catch {
    return defaultValue
  }
}

// 设置存储的数据
function setStorageData(key, value) {
  localStorage.setItem(key, JSON.stringify(value))
}

// 生成唯一 ID
function generateId() {
  return Date.now().toString(36) + Math.random().toString(36).substr(2)
}

// 绝对值
function abs(x) {
  return x < 0 ? -x : x
}

// 验证交易请求
function validateTradeRequest(req) {
  // 必填字段
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

  // 类型有效性
  if (!['buy', 'sell', 'recharge'].includes(req.type)) {
    throw new Error(`无效的交易类型: ${req.type}`)
  }

  // 数值有效性
  if (req.amount <= 0) {
    throw new Error('交易数量必须大于0')
  }
  if (req.price <= 0) {
    throw new Error('交易价格必须大于0')
  }

  // 类型特定校验
  switch (req.type) {
    case 'recharge':
      // 充值必须是USDT
      if (req.symbol !== 'USDT') {
        throw new Error('充值只支持USDT')
      }
      // 充值时价格和数量应该一致（1:1）
      if (req.price !== 1) {
        throw new Error('USDT充值价格必须为1')
      }
      break
    case 'buy':
    case 'sell':
      // 买卖不能是USDT
      if (req.symbol === 'USDT') {
        throw new Error('不能直接买卖USDT，请使用充值功能')
      }
      // 检查是否是支持的加密货币或股票
      if (!supportedCryptos[req.symbol] && !supportedUSStocks[req.symbol]) {
        throw new Error(`不支持的资产: ${req.symbol}`)
      }
      break
  }

  // 一致性检查（允许0.01的误差）
  const total = req.amount * req.price
  if (abs(total - (req.total || total)) > 0.01) {
    throw new Error(`交易金额计算不一致: ${total.toFixed(2)} != ${(req.total || total).toFixed(2)}`)
  }
}

// 验证导入的交易数据
function validateImportTrade(trade) {
  // 必填字段
  if (!trade.symbol) {
    return '币种代码不能为空'
  }
  if (!trade.type) {
    return '交易类型不能为空'
  }

  // 类型有效性
  if (!['buy', 'sell', 'recharge'].includes(trade.type)) {
    return `无效的交易类型: ${trade.type}`
  }

  // 币种有效性
  if (trade.symbol !== 'USDT' && !supportedCryptos[trade.symbol] && !supportedUSStocks[trade.symbol]) {
    return `不支持的资产: ${trade.symbol}`
  }

  // 数值有效性
  if (trade.amount <= 0) {
    return '交易数量必须大于0'
  }
  if (trade.price <= 0) {
    return '交易价格必须大于0'
  }

  // 一致性检查（允许0.01的误差）
  const expectedTotal = trade.amount * trade.price
  if (abs(expectedTotal - trade.total) > 0.01) {
    return `交易金额计算不一致: ${expectedTotal.toFixed(2)} != ${trade.total.toFixed(2)}`
  }

  // 类型特定校验
  if (trade.type === 'recharge' && trade.symbol !== 'USDT') {
    return '充值必须是USDT'
  }
  if (trade.type === 'recharge' && trade.price !== 1) {
    return 'USDT充值价格必须为1'
  }
  if ((trade.type === 'buy' || trade.type === 'sell') && trade.symbol === 'USDT') {
    return '不能直接买卖USDT'
  }

  return null
}

// ========== 持仓计算 ==========

// 重新计算所有持仓（参考 recalcAllHoldings）
function recalcAllHoldings(trades) {
  const holdings = {}

  // 按时间顺序处理交易
  const sortedTrades = [...trades].sort((a, b) => new Date(a.created_at) - new Date(b.created_at))

  for (const t of sortedTrades) {
    const total = t.amount * t.price
    switch (t.type) {
      case 'buy':
        holdings[t.symbol] = (holdings[t.symbol] || 0) + t.amount
        holdings['USDT'] = (holdings['USDT'] || 0) - total
        break
      case 'sell':
        holdings[t.symbol] = (holdings[t.symbol] || 0) - t.amount
        holdings['USDT'] = (holdings['USDT'] || 0) + total
        break
      case 'recharge':
        holdings['USDT'] = (holdings['USDT'] || 0) + t.amount
        break
    }
  }

  return holdings
}

// 重新计算单个资产持仓（参考 recalcAsset）
function recalcAsset(trades, symbol) {
  const symbolTrades = trades
    .filter(t => t.symbol === symbol)
    .sort((a, b) => new Date(a.created_at) - new Date(b.created_at))

  let amount = 0
  for (const t of symbolTrades) {
    switch (t.type) {
      case 'buy':
        amount += t.amount
        break
      case 'sell':
        amount -= t.amount
        break
    }
  }

  return amount
}

// 重新计算USDT持仓（参考 recalcUSDT）
function recalcUSDT(trades) {
  let recharge = 0
  let buyTotal = 0
  let sellTotal = 0

  for (const t of trades) {
    switch (t.type) {
      case 'recharge':
        recharge += t.amount
        break
      case 'buy':
        buyTotal += t.amount * t.price
        break
      case 'sell':
        sellTotal += t.amount * t.price
        break
    }
  }

  return recharge - buyTotal + sellTotal
}

// ========== 投资组合统计计算 ==========

// 计算投资组合统计（参考 calculatePortfolioStats）
function calculatePortfolioStats(holdings, prices, priceChanges, trades, usStockPrices = {}) {
  const portfolio = []
  let totalValue = 0
  let totalAssetsValue = 0
  let usdtBalance = 0
  let totalUnrealizedPL = 0
  let totalCost = 0
  let weightedChange = 0
  let totalRealizedPL = 0
  let totalHistoricalCost = 0
  let usStockValue = 0

  // 按时间顺序遍历交易，计算各币种的成本和实现盈亏
  const assetData = {}

  for (const t of trades) {
    if (t.symbol === 'USDT') continue

    if (!assetData[t.symbol]) {
      assetData[t.symbol] = {
        amount: 0,
        cost: 0,
        totalIn: 0,
        realizedPL: 0
      }
    }

    const d = assetData[t.symbol]
    const total = t.amount * t.price

    switch (t.type) {
      case 'buy':
        d.amount += t.amount
        d.cost += total
        d.totalIn += total
        break
      case 'sell':
        if (d.amount > 0 && t.amount > 0) {
          const sellRatio = t.amount / d.amount
          const costRecovered = d.cost * sellRatio
          const realizedPL = total - costRecovered

          d.realizedPL += realizedPL
          d.cost -= costRecovered
          d.amount -= t.amount
        }
        break
    }

    assetData[t.symbol] = d
  }

  // 先计算所有有交易记录的币种的总实现盈亏（包括已清仓的）
  for (const symbol in assetData) {
    if (symbol !== 'USDT') {
      totalRealizedPL += assetData[symbol].realizedPL
      totalHistoricalCost += assetData[symbol].totalIn
    }
  }

  // 处理持仓
  for (const symbol in holdings) {
    const amount = holdings[symbol]
    if (symbol === 'USDT') {
      usdtBalance = amount
      portfolio.push({
        symbol: symbol,
        amount: amount,
        current_price: 1.00,
        avg_cost: 0,
        market_value: usdtBalance,
        cost: 0,
        profit_loss: 0,
        pl_rate: 0,
        realized_pl: 0,
        realized_pl_rate: 0
      })
      continue
    }

    const price = prices[symbol] || 0
    const marketValue = amount * price
    totalAssetsValue += marketValue

    const d = assetData[symbol] || { amount: 0, cost: 0, totalIn: 0, realizedPL: 0 }
    const cost = d.cost
    const realizedPL = d.realizedPL

    // 持仓为0的资产只展示实现盈亏，不参与总市值计算
    if (amount === 0 && realizedPL !== 0) {
      portfolio.push({
        symbol: symbol,
        amount: 0,
        current_price: price,
        avg_cost: 0,
        market_value: 0,
        cost: 0,
        profit_loss: 0,
        pl_rate: 0,
        realized_pl: realizedPL,
        realized_pl_rate: 0
      })
      continue
    }

    totalValue += marketValue

    const avgCost = amount !== 0 ? cost / amount : 0
    const profitLoss = marketValue - cost
    const plRate = cost !== 0 ? (profitLoss / cost) * 100 : 0

    totalUnrealizedPL += profitLoss
    totalCost += cost
    weightedChange += marketValue * (priceChanges[symbol] || 0)

    // 区分加密货币和美股
    const isUSStock = usStockPrices[symbol] !== undefined
    if (isUSStock) {
      usStockValue += marketValue
    }

    // 计算实现盈亏率
    const realizedPLRate = d.totalIn !== 0 ? (realizedPL / d.totalIn) * 100 : 0

    portfolio.push({
      symbol: symbol,
      amount: amount,
      current_price: price,
      avg_cost: avgCost,
      market_value: marketValue,
      cost: cost,
      profit_loss: profitLoss,
      pl_rate: plRate,
      realized_pl: realizedPL,
      realized_pl_rate: realizedPLRate
    })
  }

  const totalUnrealizedPLRate = totalCost !== 0 ? (totalUnrealizedPL / totalCost) * 100 : 0
  const totalValueChange24h = totalValue !== 0 ? (weightedChange / totalValue) * 100 : 0
  const totalRealizedPLRate = totalHistoricalCost !== 0 ? (totalRealizedPL / totalHistoricalCost) * 100 : 0

  return {
    portfolio,
    totalValue,
    usStockValue,
    totalAssetsValue,
    usdtBalance,
    totalUnrealizedPL,
    totalUnrealizedPLRate,
    totalRealizedPL,
    totalRealizedPLRate,
    totalValueChange24h
  }
}

// ========== 价格获取 ==========

// 从 CoinCap 获取价格数据（参考 fetchPrices）
async function fetchPrices() {
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

    // 缓存价格数据
    setStorageData(STORAGE_KEYS.PRICES, prices)
    setStorageData(STORAGE_KEYS.PRICE_CHANGES, priceChanges)
    setStorageData(STORAGE_KEYS.PRICE_UPDATED_AT, updatedAt)

    return { prices, priceChanges, updatedAt }
  } catch (error) {
    console.error('获取 CoinCap 价格失败:', error)
    // 返回缓存的价格
    const cachedPrices = getStorageData(STORAGE_KEYS.PRICES, { 'USDT': 1.0 })
    const cachedChanges = getStorageData(STORAGE_KEYS.PRICE_CHANGES, { 'USDT': 0 })
    const cachedUpdatedAt = getStorageData(STORAGE_KEYS.PRICE_UPDATED_AT, Date.now())
    return { prices: cachedPrices, priceChanges: cachedChanges, updatedAt: cachedUpdatedAt }
  }
}

// 获取单个加密货币价格
async function fetchCryptoPrice(symbol) {
  if (symbol === 'USDT') {
    return { price: 1.0, updated_at: Date.now() }
  }

  const url = `${COINCAP_BASE_URL}/price/bysymbol/${symbol}`

  try {
    const response = await axios.get(url, {
      headers: {
        'Authorization': `Bearer ${COINCAP_API_KEY}`
      },
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
}

// 获取单个美股价格（从腾讯财经API）
async function fetchUSStockPrice(symbol) {
  const url = `${TENCENT_STOCK_URL}/q=us${symbol}`

  try {
    const response = await axios.get(url, {
      timeout: 10000
    })

    const text = response.data
    if (!text || !text.includes('v_us')) {
      throw new Error('股票价格数据为空')
    }

    const pattern = new RegExp(`v_us${symbol}="([^"]+)"`, 'i')
    const match = text.match(pattern)
    if (match && match[1]) {
      const data = match[1].split('~')
      if (data.length > 4) {
        const price = parseFloat(data[3]) || 0
        return { price, updated_at: Date.now() }
      }
    }

    throw new Error('解析股票价格失败')
  } catch (error) {
    console.error(`获取 ${symbol} 股票价格失败:`, error)
    const cachedPrices = getStorageData(STORAGE_KEYS.US_STOCK_PRICES, {})
    return { price: cachedPrices[symbol] || 0, updated_at: Date.now() }
  }
}

// 批量获取美股价格（从腾讯财经API）
async function fetchUSStockPricesBatch() {
  const symbols = Object.keys(supportedUSStocks)
  const stockCodes = symbols.map(s => `us${s}`).join(',')
  const url = `${TENCENT_STOCK_URL}/q=${stockCodes}`

  try {
    const response = await axios.get(url, {
      timeout: 10000
    })

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

// 获取单个资产价格（参考 GetAssetPrice）
async function fetchAssetPrice(symbol, assetType = 'crypto') {
  if (!symbol) {
    throw new Error('币种代码不能为空')
  }

  switch (assetType) {
    case 'crypto':
      if (symbol !== 'USDT' && !supportedCryptos[symbol]) {
        throw new Error('不支持的加密货币')
      }
      return fetchCryptoPrice(symbol)
    case 'us_stock':
      if (!supportedUSStocks[symbol]) {
        throw new Error('不支持的美股')
      }
      return fetchUSStockPrice(symbol)
    default:
      throw new Error(`不支持的资产类型: ${assetType}`)
  }
}

// ========== API 响应格式 ==========

function mockResponse(data) {
  return Promise.resolve({ data })
}

// ========== 导出 API ==========

export const localPortfolioApi = {
  // 获取仪表盘数据（参考 GetDashboard）
  async getDashboard() {
    const trades = getStorageData(STORAGE_KEYS.TRADES)

    // 并行获取加密货币和美股价格数据
    const [cryptoResult, usStockPrices] = await Promise.all([
      fetchPrices(),
      fetchUSStockPricesBatch()
    ])

    const { prices, priceChanges, updatedAt } = cryptoResult

    // 合并价格数据
    const allPrices = { ...prices, ...usStockPrices }

    // 重新计算所有持仓
    const holdings = recalcAllHoldings(trades)

    // 计算统计数据
    const stats = calculatePortfolioStats(holdings, allPrices, priceChanges, trades, usStockPrices)

    return mockResponse({
      prices: prices,
      us_stock_prices: usStockPrices,
      price_changes: priceChanges,
      updated_at: new Date(updatedAt).toISOString(),
      portfolio: stats.portfolio,
      crypto_value: stats.totalValue,
      us_stock_value: stats.usStockValue,
      total_assets_value: stats.totalAssetsValue,
      usdt_balance: stats.usdtBalance,
      unrealized_profit_loss: stats.totalUnrealizedPL,
      unrealized_profit_loss_rate: stats.totalUnrealizedPLRate,
      realized_profit_loss: stats.totalRealizedPL,
      realized_profit_loss_rate: stats.totalRealizedPLRate,
      total_profit_loss: stats.totalUnrealizedPL + stats.totalRealizedPL,
      value_change_24h: stats.totalValueChange24h
    })
  },

  // 获取交易记录（参考 GetTrades）
  getTrades() {
    const trades = getStorageData(STORAGE_KEYS.TRADES)
    // 按时间倒序排列
    const sortedTrades = [...trades].sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
    return mockResponse({ trades: sortedTrades })
  },

  // 创建交易（参考 CreateTrade）
  createTrade(trade) {
    // 业务层参数校验
    validateTradeRequest(trade)

    const trades = getStorageData(STORAGE_KEYS.TRADES)
    const total = trade.amount * trade.price

    // 检查余额/持仓
    const holdings = recalcAllHoldings(trades)

    switch (trade.type) {
      case 'buy':
        // 检查USDT余额
        if ((holdings['USDT'] || 0) < total) {
          throw new Error('USDT余额不足')
        }
        break
      case 'sell':
        // 检查持仓
        if ((holdings[trade.symbol] || 0) < trade.amount) {
          throw new Error('持仓不足')
        }
        break
      case 'recharge':
        // 充值不需要检查
        break
    }

    // 创建交易记录
    const newTrade = {
      id: generateId(),
      symbol: trade.symbol,
      type: trade.type,
      amount: trade.amount,
      price: trade.price,
      total: total,
      created_at: new Date().toISOString()
    }

    trades.push(newTrade)
    setStorageData(STORAGE_KEYS.TRADES, trades)

    return mockResponse({
      id: newTrade.id,
      symbol: newTrade.symbol,
      type: newTrade.type,
      amount: newTrade.amount,
      price: newTrade.price,
      total: newTrade.total,
      created_at: newTrade.created_at
    })
  },

  // 删除交易（参考 DeleteTrade）
  deleteTrade(id) {
    const trades = getStorageData(STORAGE_KEYS.TRADES)
    const tradeIndex = trades.findIndex(t => t.id === id)

    if (tradeIndex === -1) {
      throw new Error('交易记录不存在')
    }

    const trade = trades[tradeIndex]

    // 保护校验：模拟删除后的持仓状态
    const remainingTrades = trades.filter(t => t.id !== id)
    const simulatedHoldings = recalcAllHoldings(remainingTrades)

    // 保护校验：删除后不能导致任何资产负持仓
    for (const symbol in simulatedHoldings) {
      if (simulatedHoldings[symbol] < 0) {
        throw new Error(`删除该交易会导致 ${symbol} 持仓为负数(${simulatedHoldings[symbol].toFixed(8)})，无法删除`)
      }
    }

    // 执行删除
    trades.splice(tradeIndex, 1)
    setStorageData(STORAGE_KEYS.TRADES, trades)

    return mockResponse({
      success: true,
      message: '交易记录已删除',
      deleted_trade: {
        id: trade.id,
        symbol: trade.symbol,
        type: trade.type,
        amount: trade.amount,
        price: trade.price,
        created_at: trade.created_at
      }
    })
  },

  // 清空交易（参考 ClearTrades）
  clearTrades() {
    setStorageData(STORAGE_KEYS.TRADES, [])
    return mockResponse({ success: true, message: '所有数据已清空' })
  },

  // 获取资产价格（参考 GetAssetPrice）
  async getAssetPrice(symbol, assetType = 'crypto') {
    const result = await fetchAssetPrice(symbol, assetType)
    return mockResponse({
      symbol: symbol,
      price: result.price,
      updated_at: new Date(result.updated_at).toISOString()
    })
  },

  // 导出数据（参考 ExportDataHandler）
  exportData() {
    const trades = getStorageData(STORAGE_KEYS.TRADES)

    const exportTrades = trades.map(t => ({
      symbol: t.symbol,
      type: t.type,
      amount: t.amount,
      price: t.price,
      total: t.total,
      created_at: t.created_at,
      notes: t.notes || ''
    }))

    // 按时间正序排列
    exportTrades.sort((a, b) => new Date(a.created_at) - new Date(b.created_at))

    return mockResponse({
      success: true,
      data: {
        version: '1.0',
        export_time: new Date().toISOString(),
        app_name: 'fx-portfolio',
        trades: exportTrades
      }
    })
  },

  // 导入预览（参考 ImportPreviewHandler）
  importPreview(data) {
    // 验证版本
    if (data.version !== '1.0') {
      throw new Error(`不支持的版本: ${data.version}`)
    }

    const existingTrades = getStorageData(STORAGE_KEYS.TRADES)

    // 构建现有交易的时间戳集合
    const existingMap = new Set()
    for (const t of existingTrades) {
      const key = `${t.symbol}_${t.type}_${t.created_at}`
      existingMap.add(key)
    }

    const preview = {
      total_trades: data.trades.length,
      new_trades: 0,
      conflicts: 0,
      conflict_items: []
    }

    for (const trade of data.trades) {
      // 验证交易数据
      const error = validateImportTrade(trade)
      if (error) {
        preview.conflicts++
        preview.conflict_items.push({
          trade: trade,
          reason: error
        })
        continue
      }

      // 检查是否已存在
      const key = `${trade.symbol}_${trade.type}_${trade.created_at}`
      if (existingMap.has(key)) {
        preview.conflicts++
        preview.conflict_items.push({
          trade: trade,
          reason: '交易记录已存在'
        })
      } else {
        preview.new_trades++
      }
    }

    return mockResponse({ success: true, preview })
  },

  // 确认导入（参考 ImportConfirmHandler）
  importConfirm(data, conflictStrategy = 'skip') {
    // 验证版本
    if (data.version !== '1.0') {
      throw new Error(`不支持的版本: ${data.version}`)
    }

    const existingTrades = getStorageData(STORAGE_KEYS.TRADES)

    // 构建现有交易映射
    const existingMap = new Map()
    for (const t of existingTrades) {
      const key = `${t.symbol}_${t.type}_${t.created_at}`
      existingMap.set(key, t)
    }

    let imported = 0
    let skipped = 0
    let overwritten = 0

    for (const trade of data.trades) {
      // 验证交易数据
      const error = validateImportTrade(trade)
      if (error) {
        skipped++
        continue
      }

      const key = `${trade.symbol}_${trade.type}_${trade.created_at}`
      const exists = existingMap.has(key)

      if (exists) {
        // 处理冲突
        switch (conflictStrategy) {
          case 'skip':
            skipped++
            continue
          case 'overwrite':
            // 删除旧记录
            const index = existingTrades.findIndex(t =>
              t.symbol === trade.symbol &&
              t.type === trade.type &&
              t.created_at === trade.created_at
            )
            if (index !== -1) {
              existingTrades.splice(index, 1)
            }
            overwritten++
            break
        }
      }

      // 创建新交易记录
      const newTrade = {
        id: generateId(),
        symbol: trade.symbol,
        type: trade.type,
        amount: trade.amount,
        price: trade.price,
        total: trade.total,
        created_at: trade.created_at
      }

      existingTrades.push(newTrade)
      imported++
    }

    // 保存交易记录
    setStorageData(STORAGE_KEYS.TRADES, existingTrades)

    return mockResponse({
      success: true,
      imported,
      skipped,
      overwritten
    })
  }
}
