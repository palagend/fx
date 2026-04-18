import apiClient from './axios'

export const portfolioApi = {
  // 获取仪表盘聚合数据（价格+持仓+统计）
  getDashboard() {
    return apiClient.get('/portfolio/dashboard')
  },

  // 获取交易记录
  getTrades() {
    return apiClient.get('/portfolio/trades')
  },

  // 创建交易
  createTrade(trade) {
    return apiClient.post('/portfolio/trades', trade)
  },

  // 删除交易记录
  deleteTrade(id) {
    return apiClient.delete(`/portfolio/trades/${id}`)
  },

  // 清空交易记录
  clearTrades() {
    return apiClient.delete('/portfolio/trades')
  },

  // 获取单个资产价格（用于交易时获取最新价）
  getAssetPrice(symbol) {
    return apiClient.get(`/prices/${symbol}`)
  }
}
