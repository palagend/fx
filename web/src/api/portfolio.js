import axios from 'axios'

const API_BASE_URL = '/api'

export const portfolioApi = {
  // 获取仪表盘聚合数据（价格+持仓+统计）
  getDashboard() {
    return axios.get(`${API_BASE_URL}/portfolio/dashboard`)
  },

  // 获取交易记录
  getTrades() {
    return axios.get(`${API_BASE_URL}/portfolio/trades`)
  },

  // 创建交易
  createTrade(trade) {
    return axios.post(`${API_BASE_URL}/portfolio/trades`, trade)
  },

  // 删除交易记录
  deleteTrade(id) {
    return axios.delete(`${API_BASE_URL}/portfolio/trades/${id}`)
  },

  // 清空交易记录
  clearTrades() {
    return axios.delete(`${API_BASE_URL}/portfolio/trades`)
  },

  // 获取单个资产价格（用于交易时获取最新价）
  getAssetPrice(symbol) {
    return axios.get(`${API_BASE_URL}/prices/${symbol}`)
  }
}
