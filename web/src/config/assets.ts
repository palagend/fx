export interface AssetTypeConfig {
  id: string
  name: string
  currency: string
  currencyName: string
  currencyIcon: string
  priceApi: string
  defaultColor: string
}

export const ASSET_TYPES: Record<string, AssetTypeConfig> = {
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

export const CURRENCY_RATES: Record<string, number> = {
  USD: 1,
  USDT: 1,
  CNY: 0.14,
  HKD: 0.128
}

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
  } as Record<string, string>,
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
  } as Record<string, string>,
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
  } as Record<string, string>
}

export const ASTOCK_CONFIG = {
  COLORS: {
    '600519': '#E60012',
    '000858': '#FF6B00',
    '000333': '#0066CC',
    '002415': '#00A86B',
    '300750': '#FF1744',
    '601318': '#9C27B0',
    '600036': '#3F51B5',
    '000002': '#795548'
  } as Record<string, string>,
  ICONS: {
    '600519': 'mdi:glass-wine',
    '000858': 'mdi:glass-mug',
    '000333': 'mdi:air-conditioner',
    '002415': 'mdi:cctv',
    '300750': 'mdi:battery-charging',
    '601318': 'mdi:shield-check',
    '600036': 'mdi:bank',
    '000002': 'mdi:home-city'
  } as Record<string, string>,
  NAMES: {
    '600519': '贵州茅台',
    '000858': '五粮液',
    '000333': '美的集团',
    '002415': '海康威视',
    '300750': '宁德时代',
    '601318': '中国平安',
    '600036': '招商银行',
    '000002': '万科A'
  } as Record<string, string>
}

export const USSTOCK_CONFIG = {
  COLORS: {
    'AAPL': '#555555',
    'MSFT': '#00A4EF',
    'GOOG': '#4285F4',
    'AMZN': '#FF9900',
    'TSLA': '#CC0000',
    'META': '#0081FB',
    'NVDA': '#76B900',
    'BABA': '#FF6A00',
    'ORCL': '#F80000',
    'CRCL': '#00A86B',
    'MSTR': '#1A1A1A'
  } as Record<string, string>,
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
  } as Record<string, string>,
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
  } as Record<string, string>
}

export const HKSTOCK_CONFIG = {
  COLORS: {
    '0700': '#00A1E0',
    '3690': '#FFD100',
    '9988': '#FF6A00',
    '2318': '#9C27B0',
    '0005': '#E60012'
  } as Record<string, string>,
  ICONS: {
    '0700': 'mdi:message-text',
    '3690': 'mdi:food-delivery',
    '9988': 'mdi:hospital-box',
    '2318': 'mdi:shield-check',
    '0005': 'mdi:bank'
  } as Record<string, string>,
  NAMES: {
    '0700': '腾讯控股',
    '3690': '美团',
    '9988': '阿里健康',
    '2318': '中国平安',
    '0005': '汇丰控股'
  } as Record<string, string>
}

export const AVAILABLE_ASSETS = {
  CRYPTO: ['BTC', 'ETH', 'BNB', 'XRP', 'ADA', 'SOL', 'DOGE', 'TRX', 'AVAX', 'HYPE'] as const,
  A_STOCK: ['600519', '000858', '000333', '002415', '300750', '601318', '600036', '000002'] as const,
  US_STOCK: ['AAPL', 'MSFT', 'GOOG', 'AMZN', 'TSLA', 'META', 'NVDA', 'BABA', 'ORCL', 'CRCL', 'MSTR'] as const,
  HK_STOCK: ['0700', '3690', '9988', '2318', '0005'] as const
}

export const AVAILABLE_SYMBOLS = AVAILABLE_ASSETS.CRYPTO

export const getAssetTypeConfig = (assetType?: string): AssetTypeConfig => {
  return ASSET_TYPES[assetType?.toUpperCase()] || ASSET_TYPES.CRYPTO
}

export interface AssetConfig {
  color: string
  icon: string
  name: string
}

export const getAssetConfig = (assetType: string, symbol: string): AssetConfig => {
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

export const getAssetColor = (assetTypeOrSymbol: string, symbol?: string): string => {
  const assetType = symbol ? assetTypeOrSymbol : 'crypto'
  const sym = symbol || assetTypeOrSymbol
  return getAssetConfig(assetType, sym).color
}

export const getAssetIcon = (assetTypeOrSymbol: string, symbol?: string): string => {
  const assetType = symbol ? assetTypeOrSymbol : 'crypto'
  const sym = symbol || assetTypeOrSymbol
  return getAssetConfig(assetType, sym).icon
}

export const getAssetName = (assetTypeOrSymbol: string, symbol?: string): string => {
  const assetType = symbol ? assetTypeOrSymbol : 'crypto'
  const sym = symbol || assetTypeOrSymbol
  return getAssetConfig(assetType, sym).name
}

export const convertCurrency = (amount: number, fromCurrency: string, toCurrency: string = 'USD', rates: Record<string, number> = CURRENCY_RATES): number => {
  if (fromCurrency === toCurrency) return amount
  const fromRate = rates[fromCurrency] || 1
  const toRate = rates[toCurrency] || 1
  return amount * (fromRate / toRate)
}

export const getCurrencySymbol = (currency: string): string => {
  const symbols: Record<string, string> = {
    USD: '$',
    USDT: '$',
    CNY: '¥',
    HKD: 'HK$'
  }
  return symbols[currency] || '$'
}