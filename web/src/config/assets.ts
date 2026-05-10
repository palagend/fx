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
    currency: 'USD',
    currencyName: '美元',
    currencyIcon: 'mdi:currency-usd',
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
  CNY: 0.14,
  HKD: 0.128
}

export const CRYPTO_CONFIG = {
  COLORS: {
    BTC: '#F7931A',
    ETH: '#627EEA',
    BNB: '#F3BA2F',
    XRP: '#00A5DF',
    ADA: '#0033AD',
    SOL: '#9945FF',
    DOGE: '#C2A633',
    TRX: '#EB0029',
    AVAX: '#E84142',
    HYPE: '#89F0E6',
    POL: '#8247E5',
    DOT: '#E6007A'
  } as Record<string, string>,
  ICONS: {
    BTC: 'cryptocurrency-color:btc',
    ETH: 'cryptocurrency-color:eth',
    BNB: 'cryptocurrency-color:bnb',
    XRP: 'cryptocurrency-color:xrp',
    ADA: 'cryptocurrency-color:ada',
    SOL: 'token-branded:sol',
    DOGE: 'cryptocurrency-color:doge',
    TRX: 'cryptocurrency-color:trx',
    AVAX: 'cryptocurrency-color:avax',
    HYPE: 'token:hyper-evm',
    POL: 'token-branded:polygon-zkevm',
    DOT: 'token-branded:polkadot'
  } as Record<string, string>,
  NAMES: {
    BTC: 'Bitcoin',
    ETH: 'Ethereum',
    BNB: 'Binance Coin',
    XRP: 'Ripple',
    ADA: 'Cardano',
    SOL: 'Solana',
    DOGE: 'Dogecoin',
    TRX: 'Tron',
    AVAX: 'Avalanche',
    HYPE: 'Hyperliquid',
    POL: 'Polygon',
    DOT: 'Polkadot'
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
    'AAPL': '#C0C0C0',      // 亮银灰 - Apple 银色风格
    'MSFT': '#00BCF2',      // 亮蓝 - Microsoft 品牌色
    'GOOG': '#FBBC04',      // 金黄 - Google 四色之一，温暖明亮
    'AMZN': '#FF9500',      // 橙黄 - Amazon 温暖橙色
    'TSLA': '#E82127',      // 特斯拉红 - 品牌标志性红色
    'META': '#0668E1',      // 深蓝 - Meta 品牌蓝
    'NVDA': '#76B900',      // 亮绿 - NVIDIA 品牌绿，保持不变
    'BABA': '#FF6A00',      // 阿里橙 - 保持不变
    'ORCL': '#C74634',      // 砖红 - Oracle 品牌红，区别于特斯拉
    'CRCL': '#00D4AA',      // 青绿 - Circle 稳定币清新色
    'MSTR': '#E04403',      // 深橙 - MicroStrategy 强调色
    'QQQI': '#8E44AD'       // 紫罗兰 - QQQI 独特紫色，区别于其他
  } as Record<string, string>,
  ICONS: {
    'AAPL': 'logos:apple',
    'MSFT': 'logos:microsoft-icon',
    'GOOG': 'logos:google-icon',
    'AMZN': 'simple-icons:amazon',
    'TSLA': 'simple-icons:tesla',
    'META': 'logos:meta-icon',
    'NVDA': 'simple-icons:nvidia',
    'BABA': 'simple-icons:alibabadotcom',
    'ORCL': 'simple-icons:oracle',
    'CRCL': 'simple-icons:circle',
    'MSTR': 'simple-icons:microstrategy',
    'QQQI': 'mdi:chart-line'
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
    'MSTR': 'MicroStrategy',
    'QQQI': 'Nasdaq 100 Income'
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
  CRYPTO: ['BTC', 'ETH', 'BNB', 'SOL', 'TRX', 'HYPE', 'XRP', 'POL', 'DOGE', 'AVAX', 'ADA', 'DOT'] as const,
  A_STOCK: ['600519', '000858', '000333', '002415', '300750', '601318', '600036', '000002'] as const,
  US_STOCK: ['GOOG', 'TSLA', 'CRCL', 'QQQI', 'NVDA', 'AAPL', 'MSTR', 'AMZN', 'BABA', 'META', 'MSFT', 'ORCL'] as const,
  HK_STOCK: ['0700', '3690', '9988', '2318', '0005'] as const
}

export const AVAILABLE_SYMBOLS = AVAILABLE_ASSETS.CRYPTO

export const getAssetTypeConfig = (assetType?: string): AssetTypeConfig => {
  const key = assetType?.toUpperCase()
  return key ? ASSET_TYPES[key] || ASSET_TYPES.CRYPTO : ASSET_TYPES.CRYPTO
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
    CNY: '¥',
    HKD: 'HK$'
  }
  return symbols[currency] || '$'
}