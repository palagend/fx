import { apiClient } from './axios'

export interface TradeRequest {
  asset_type: string
  symbol: string
  type: string
  amount: number
  price: number
}

export interface Trade {
  id: number
  uuid: string
  asset_type: string
  symbol: string
  type: string
  amount: number
  price: number
  total: number
  currency: string
  created_at: string
}

export interface PortfolioItem {
  asset_type: string
  symbol: string
  amount: number
  current_price: number
  avg_cost: number
  market_value: number
  cost: number
  profit_loss: number
  pl_rate: number
  realized_pl: number
  realized_pl_rate: number
  currency: string
}

export interface DashboardResponse {
  prices: Record<string, number>
  us_stock_prices: Record<string, number>
  price_changes: Record<string, number>
  crypto_updated_at: number
  btc_price: number
  portfolio: PortfolioItem[]
  crypto_value: number
  us_stock_value: number
  cash_balance: number
  total_assets_value: number
  unrealized_profit_loss: number
  unrealized_profit_loss_rate: number
  realized_profit_loss: number
  realized_profit_loss_rate: number
  total_profit_loss: number
  value_change_24h: number
}

export interface TradesResponse {
  trades: Trade[]
}

export interface CreateTradeResponse {
  id: string
  asset_type: string
  symbol: string
  type: string
  amount: number
  price: number
  total: number
  currency: string
  created_at: string
}

export interface DeleteTradeResponse {
  message: string
  deleted_trade: {
    id: string
    asset_type: string
    symbol: string
    type: string
    amount: number
    price: number
    created_at: string
  }
}

export interface AssetPriceResponse {
  symbol: string
  price: number
  asset_type: string
  currency: string
  updated_at: string
}

export interface ExportDataResponse {
  data: {
    version: string
    exported: string
    trades: Trade[]
    fingerprint?: string
  }
}

export interface ImportPreviewResponse {
  preview: {
    total_trades: number
    new_trades: number
    conflicts: number
    conflict_items: {
      trade: Trade
      reason: string
    }[]
  }
}

export interface ImportConfirmResponse {
  imported: number
  skipped: number
  overwritten: number
}

export const backendPortfolioApi = {
  getDashboard(): Promise<AxiosResponse<DashboardResponse>> {
    return apiClient.get('/portfolio/dashboard')
  },

  getTrades(): Promise<AxiosResponse<TradesResponse>> {
    return apiClient.get('/portfolio/trades')
  },

  createTrade(trade: TradeRequest): Promise<AxiosResponse<CreateTradeResponse>> {
    return apiClient.post('/portfolio/trades', trade)
  },

  deleteTrade(id: number): Promise<AxiosResponse<DeleteTradeResponse>> {
    return apiClient.delete(`/portfolio/trades/${id}`)
  },

  clearTrades(): Promise<AxiosResponse<{ message: string }>> {
    return apiClient.delete('/portfolio/trades')
  },

  getAssetPrice(symbol: string, assetType: string = 'crypto'): Promise<AxiosResponse<AssetPriceResponse>> {
    return apiClient.get(`/prices/${symbol}?asset_type=${assetType}`)
  },

  exportData(): Promise<AxiosResponse<ExportDataResponse>> {
    return apiClient.get('/portfolio/export')
  },

  importPreview(data: unknown): Promise<AxiosResponse<ImportPreviewResponse>> {
    return apiClient.post('/portfolio/import/preview', { data })
  },

  importConfirm(data: unknown, conflictStrategy: string = 'skip'): Promise<AxiosResponse<ImportConfirmResponse>> {
    return apiClient.post('/portfolio/import/confirm', {
      data,
      conflict_strategy: conflictStrategy
    })
  }
}

import type { AxiosResponse } from 'axios'