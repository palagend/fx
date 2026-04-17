import axios from 'axios'

const API_BASE_URL = '/api'

export const portfolioApi = {
  // 获取完整资产组合（持仓 + 实现盈亏）
  getPortfolio() {
    return axios.get(`${API_BASE_URL}/portfolio`)
  },

  // 获取持仓列表（包含成本价）
  getHoldings() {
    return axios.get(`${API_BASE_URL}/portfolio/holdings`)
  },

  // 获取投资记录（实现盈亏）
  getInvestments() {
    return axios.get(`${API_BASE_URL}/portfolio/investments`)
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

  // 获取所有资产价格
  getAllPrices() {
    return axios.get(`${API_BASE_URL}/prices`)
  },

  // 获取单个资产价格
  getAssetPrice(symbol) {
    return axios.get(`${API_BASE_URL}/prices/${symbol}`)
  }
}
