// 资产配置 - 颜色与各币种官方主题色保持一致
// 参考来源：各币种官方品牌指南
export const ASSET_CONFIG = {
  COLORS: {
    // Tether - 官方绿色 #26A17B
    USDT: '#26A17B',
    // Bitcoin - 官方橙色 #F7931A
    BTC: '#F7931A',
    // Ethereum - 官方紫蓝色 #627EEA
    ETH: '#627EEA',
    // Binance - 官方金黄色 #F3BA2F
    BNB: '#F3BA2F',
    // XRP/Ripple - 官方蓝色 #00A5DF (企业蓝)
    XRP: '#00A5DF',
    // Cardano - 官方深蓝色 #0033AD
    ADA: '#0033AD',
    // Solana - 官方紫色 #9945FF (主要品牌色)
    SOL: '#9945FF',
    // Dogecoin - 官方金色 #C2A633
    DOGE: '#C2A633',
    // TRON - 官方红色 #EB0029
    TRX: '#EB0029',
    // Avalanche - 官方红色 #E84142 (品牌主色)
    AVAX: '#E84142',
    // Hyperliquid - 官方青色/蓝绿色 #89F0E6
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

// 支持的加密货币列表
export const AVAILABLE_SYMBOLS = ['BTC', 'ETH', 'BNB', 'XRP', 'ADA', 'SOL', 'DOGE', 'TRX', 'AVAX', 'HYPE']

// 获取资产颜色
export const getAssetColor = (symbol) => ASSET_CONFIG.COLORS[symbol] || '#667eea'

// 获取资产图标
export const getAssetIcon = (symbol) => ASSET_CONFIG.ICONS[symbol] || 'mdi:currency-usd'

// 获取资产名称
export const getAssetName = (symbol) => ASSET_CONFIG.NAMES[symbol] || symbol
