package models

import (
	"time"

	"gorm.io/gorm"
)

// AssetType 资产类型
const (
	AssetTypeCrypto  = "crypto"   // 加密货币
	AssetTypeAStock  = "a_stock"  // A股
	AssetTypeUSStock = "us_stock" // 美股
	AssetTypeHKStock = "hk_stock" // 港股
	AssetTypeCash    = "cash"     // 现金
)

// Currency 货币类型
const (
	CurrencyUSD = "USD" // 美元（统一计价货币）
	CurrencyCNY = "CNY" // 人民币
	CurrencyHKD = "HKD" // 港币
)

// Trade 交易记录
type Trade struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	UUID      string         `gorm:"size:36;uniqueIndex" json:"uuid"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	AssetType string         `gorm:"size:20;not null;default:'crypto'" json:"asset_type"`
	Symbol    string         `gorm:"size:20;not null" json:"symbol"`
	Type      string         `gorm:"size:10;not null" json:"type"`
	Amount    float64        `gorm:"type:decimal(20,8);not null" json:"amount"`
	Price     float64        `gorm:"type:decimal(20,8);not null" json:"price"`
	Total     float64        `gorm:"type:decimal(20,8);not null" json:"total"`
	Currency  string         `gorm:"size:10;not null;default:'USD'" json:"currency"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Holding 资产持仓
type Holding struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	AssetType string         `gorm:"size:20;not null;default:'crypto'" json:"asset_type"`
	Symbol    string         `gorm:"size:20;not null" json:"symbol"`
	Amount    float64        `gorm:"type:decimal(20,8);not null" json:"amount"`
	Currency  string         `gorm:"size:10;not null;default:'USD'" json:"currency"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ExchangeRate 汇率记录
type ExchangeRate struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	From      string    `gorm:"size:10;not null" json:"from"`
	To        string    `gorm:"size:10;not null" json:"to"`
	Rate      float64   `gorm:"type:decimal(20,8);not null" json:"rate"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetCurrencyByAssetType 根据资产类型获取默认货币
func GetCurrencyByAssetType(assetType string) string {
	switch assetType {
	case AssetTypeCrypto:
		return CurrencyUSD
	case AssetTypeAStock:
		return CurrencyCNY
	case AssetTypeUSStock:
		return CurrencyUSD
	case AssetTypeHKStock:
		return CurrencyHKD
	case AssetTypeCash:
		return CurrencyUSD
	default:
		return CurrencyUSD
	}
}
