export interface Asset {
  symbol: string
  name: string
  amount: number
  price: number
  market_value: number
  cost_basis: number
  profit_loss: number
  profit_loss_percent: number
  asset_type: 'crypto' | 'us_stock' | 'cash'
}

export interface Trade {
  id: string
  symbol: string
  asset_type: 'crypto' | 'us_stock'
  type: 'buy' | 'sell'
  amount: number
  price: number
  total: number
  fee: number
  timestamp: number
  cost_price?: number
}

export interface PortfolioAllocation {
  name: string
  value: number
  percentage: number
  color: string
}

export interface User {
  id: string
  username: string
  email: string
  settings: UserSettings
}

export interface UserSettings {
  theme: 'light' | 'dark' | 'system'
  currency: string
}

export interface PasswordEntry {
  id: string
  website: string
  username: string
  password: string
  notes?: string
  createdAt: number
}

export type TransactionType = 'buy' | 'sell'