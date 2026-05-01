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
  fee?: number
  timestamp?: number
  cost_price?: number
  currency?: string
  created_at?: string
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
  settings?: UserSettings
  created_at?: string
}

export interface UserSettings {
  theme: 'light' | 'dark' | 'system'
  currency: string
}

export interface PasswordEntry {
  id: string
  website?: string
  title?: string
  username: string
  password: string
  url?: string
  tags?: string[]
  notes?: string
  createdAt?: number
  updatedAt?: number
  useCount?: number
  lastUsedAt?: number
  strength?: number
  length?: number
}