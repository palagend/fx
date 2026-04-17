import axios from 'axios'

const API_BASE_URL = '/api'

export const portfolioApi = {
  // 获取资产组合摘要
  getSummary() {
    return axios.get(`${API_BASE_URL}/portfolio/summary`)
  },

  // 获取持仓列表
  getHoldings() {
    return axios.get(`${API_BASE_URL}/portfolio/holdings`)
  },

  // 获取交易记录
  getTrades(params = {}) {
    const { page = 1, page_size = 50, type } = params
    let url = `${API_BASE_URL}/portfolio/trades?page=${page}&page_size=${page_size}`
    if (type) {
      url += `&type=${type}`
    }
    return axios.get(url)
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
  },

  // 导入数据（覆盖式）
  importData(data) {
    return axios.post(`${API_BASE_URL}/portfolio/import`, data)
  },

  // 导出数据
  exportData() {
    return axios.get(`${API_BASE_URL}/portfolio/export`)
  }
}
