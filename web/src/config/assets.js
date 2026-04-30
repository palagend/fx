// 资产配置 - 支持多资产类型（加密货币、A股、美股、港股）
// 参考来源：各币种/股票官方品牌指南

// 资产类型定义
export const ASSET_TYPES = {
  CRYPTO: {
    id: 'crypto',
    name: '加密货币',
    currency: 'USDT',
    currencyName: '泰达币',
    currencyIcon: 'cryptocurrency-color:usdt',
    priceApi: 'coincap',
    defaultColor: '#26A17B'
  },
  A_STOCK: {
    id: 'a_stock',
    name: 'A股',
    currency: 'CNY',
    currencyName: '人民币',
    currencyIcon: 'mdi:currency-cny',
    priceApi: 'tushare',
    defaultColor: '#E60012'
  },
  US_STOCK: {
    id: 'us_stock',
    name: '美股',
    currency: 'USD',
    currencyName: '美元',
    currencyIcon: 'mdi:currency-usd',
    priceApi: 'alphavantage',
    defaultColor: '#1E88E5'
  },
  HK_STOCK: {
    id: 'hk_stock',
    name: '港股',
    currency: 'HKD',
    currencyName: '港币',
    currencyIcon: 'mdi:currency-hkd',
    priceApi: 'tushare',
    defaultColor: '#8E24AA'
  }
}

// 货币汇率基准（以USD为基准）
export const CURRENCY_RATES = {
  USD: 1,
  USDT: 1,  // 稳定币近似等于USD
  CNY: 0.14,  // 1 CNY ≈ 0.14 USD (汇率会实时更新)
  HKD: 0.128  // 1 HKD ≈ 0.128 USD
}

// 加密货币配置
export const CRYPTO_CONFIG = {
  COLORS: {
    USDT: '#26A17B',
    BTC: '#F7931A',
    ETH: '#627EEA',
    BNB: '#F3BA2F',
    XRP: '#00A5DF',
    ADA: '#0033AD',
    SOL: '#9945FF',
    DOGE: '#C2A633',
    TRX: '#EB0029',
    AVAX: '#E84142',
    HYPE: '#89F0E6'
  },
  ICONS: {
    USDT: 'cryptocurrency-color:usdt',
    BTC: 'cryptocurrency-color:btc',
    ETH: 'cryptocurrency-color:eth',
    BNB: 'cryptocurrency-color:bnb',
    XRP: 'cryptocurrency-color:xrp',
    ADA: 'cryptocurrency-color:ada',
    SOL: 'token-branded:sol',
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

// A股配置（示例股票）
export const ASTOCK_CONFIG = {
  COLORS: {
    '600519': '#E60012',  // 茅台
    '000858': '#FF6B00',  // 五粮液
    '000333': '#0066CC',  // 美的
    '002415': '#00A86B',  // 海康威视
    '300750': '#FF1744',  // 宁德时代
    '601318': '#9C27B0',  // 中国平安
    '600036': '#3F51B5',  // 招商银行
    '000002': '#795548'   // 万科
  },
  ICONS: {
    '600519': 'mdi:glass-wine',
    '000858': 'mdi:glass-mug',
    '000333': 'mdi:air-conditioner',
    '002415': 'mdi:cctv',
    '300750': 'mdi:battery-charging',
    '601318': 'mdi:shield-check',
    '600036': 'mdi:bank',
    '000002': 'mdi:home-city'
  },
  NAMES: {
    '600519': '贵州茅台',
    '000858': '五粮液',
    '000333': '美的集团',
    '002415': '海康威视',
    '300750': '宁德时代',
    '601318': '中国平安',
    '600036': '招商银行',
    '000002': '万科A'
  }
}

// 美股配置
export const USSTOCK_CONFIG = {
  COLORS: {
    'AAPL': '#555555',  // 苹果
    'MSFT': '#00A4EF',  // 微软
    'GOOG': '#4285F4', // 谷歌
    'AMZN': '#FF9900',  // 亚马逊
    'TSLA': '#CC0000',  // 特斯拉
    'META': '#0081FB',  // Meta
    'NVDA': '#76B900',  // 英伟达
    'BABA': '#FF6A00',  // 阿里
    'ORCL': '#F80000',  // Oracle
    'CRCL': '#00A86B',  // Circle
    'MSTR': '#1A1A1A'   // MicroStrategy
  },
  ICONS: {
    'AAPL': 'simple-icons:apple',
    'MSFT': 'simple-icons:microsoft',
    'GOOG': 'simple-icons:google',
    'AMZN': 'simple-icons:amazon',
    'TSLA': 'simple-icons:tesla',
    'META': 'simple-icons:meta',
    'NVDA': 'simple-icons:nvidia',
    'BABA': 'simple-icons:alibabadotcom',
    'ORCL': 'simple-icons:oracle',
    'CRCL': 'simple-icons:circle',
    'MSTR': 'simple-icons:microstrategy'
  },
  NAMES: {
    'AAPL': 'Apple',
    'MSFT': 'Microsoft',
    'GOOG': 'Alphabet',
    'AMZN': 'Amazon',
    'TSLA': 'Tesla',
    'META': 'Meta',
    'NVDA': 'NVIDIA',
    'BABA': 'Alibaba',
    'ORCL': 'Oracle',
    'CRCL': 'Circle',
    'MSTR': 'MicroStrategy'
  }
}

// 港股配置（示例股票）
export const HKSTOCK_CONFIG = {
  COLORS: {
    '0700': '#00A1E0',  // 腾讯
    '3690': '#FFD100',  // 美团
    '9988': '#FF6A00',  // 阿里健康
    '2318': '#9C27B0',  // 平安
    '0005': '#E60012'   // 汇丰
  },
  ICONS: {
    '0700': 'mdi:message-text',
    '3690': 'mdi:food-delivery',
    '9988': 'mdi:hospital-box',
    '2318': 'mdi:shield-check',
    '0005': 'mdi:bank'
  },
  NAMES: {
    '0700': '腾讯控股',
    '3690': '美团',
    '9988': '阿里健康',
    '2318': '中国平安',
    '0005': '汇丰控股'
  }
}

// 支持的资产列表
export const AVAILABLE_ASSETS = {
  CRYPTO: ['BTC', 'ETH', 'BNB', 'XRP', 'ADA', 'SOL', 'DOGE', 'TRX', 'AVAX', 'HYPE'],
  A_STOCK: ['600519', '000858', '000333', '002415', '300750', '601318', '600036', '000002'],
  US_STOCK: ['AAPL', 'MSFT', 'GOOG', 'AMZN', 'TSLA', 'META', 'NVDA', 'BABA', 'ORCL', 'CRCL', 'MSTR'],
  HK_STOCK: ['0700', '3690', '9988', '2318', '0005']
}

// 向后兼容：默认导出加密货币列表
export const AVAILABLE_SYMBOLS = AVAILABLE_ASSETS.CRYPTO

// 获取资产类型配置
export const getAssetTypeConfig = (assetType) => {
  return ASSET_TYPES[assetType?.toUpperCase()] || ASSET_TYPES.CRYPTO
}

// 获取资产配置（根据资产类型和代码）
export const getAssetConfig = (assetType, symbol) => {
  switch (assetType) {
    case 'a_stock':
      return {
        color: ASTOCK_CONFIG.COLORS[symbol] || '#E60012',
        icon: ASTOCK_CONFIG.ICONS[symbol] || 'mdi:chart-line',
        name: ASTOCK_CONFIG.NAMES[symbol] || symbol
      }
    case 'us_stock':
      return {
        color: USSTOCK_CONFIG.COLORS[symbol] || '#1E88E5',
        icon: USSTOCK_CONFIG.ICONS[symbol] || 'mdi:chart-line',
        name: USSTOCK_CONFIG.NAMES[symbol] || symbol
      }
    case 'hk_stock':
      return {
        color: HKSTOCK_CONFIG.COLORS[symbol] || '#8E24AA',
        icon: HKSTOCK_CONFIG.ICONS[symbol] || 'mdi:chart-line',
        name: HKSTOCK_CONFIG.NAMES[symbol] || symbol
      }
    case 'crypto':
    default:
      return {
        color: CRYPTO_CONFIG.COLORS[symbol] || '#667eea',
        icon: CRYPTO_CONFIG.ICONS[symbol] || 'mdi:currency-usd',
        name: CRYPTO_CONFIG.NAMES[symbol] || symbol
      }
  }
}

// 获取资产颜色（向后兼容：如果只传一个参数，默认为加密货币）
export const getAssetColor = (assetTypeOrSymbol, symbol) => {
  const assetType = symbol ? assetTypeOrSymbol : 'crypto'
  const sym = symbol || assetTypeOrSymbol
  return getAssetConfig(assetType, sym).color
}

// 获取资产图标（向后兼容：如果只传一个参数，默认为加密货币）
export const getAssetIcon = (assetTypeOrSymbol, symbol) => {
  const assetType = symbol ? assetTypeOrSymbol : 'crypto'
  const sym = symbol || assetTypeOrSymbol
  return getAssetConfig(assetType, sym).icon
}

// 获取资产名称（向后兼容：如果只传一个参数，默认为加密货币）
export const getAssetName = (assetTypeOrSymbol, symbol) => {
  const assetType = symbol ? assetTypeOrSymbol : 'crypto'
  const sym = symbol || assetTypeOrSymbol
  return getAssetConfig(assetType, sym).name
}

// 货币换算（统一到目标货币）
export const convertCurrency = (amount, fromCurrency, toCurrency = 'USD', rates = CURRENCY_RATES) => {
  if (fromCurrency === toCurrency) return amount
  const fromRate = rates[fromCurrency] || 1
  const toRate = rates[toCurrency] || 1
  return amount * (fromRate / toRate)
}

// 获取货币符号
export const getCurrencySymbol = (currency) => {
  const symbols = {
    USD: '$',
    USDT: '$',
    CNY: '¥',
    HKD: 'HK$'
  }
  return symbols[currency] || '$'
}
