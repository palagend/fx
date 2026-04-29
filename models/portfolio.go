package models

import (
	"time"

	"gorm.io/gorm"
)

// AssetType 资产类型
const (
	AssetTypeCrypto   = "crypto"    // 加密货币
	AssetTypeAStock   = "a_stock"   // A股
	AssetTypeUSStock  = "us_stock"  // 美股
	AssetTypeHKStock  = "hk_stock"  // 港股
)

// Currency 货币类型
const (
	CurrencyUSDT = "USDT" // 加密货币计价
	CurrencyCNY  = "CNY"  // 人民币
	CurrencyUSD  = "USD"  // 美元
	CurrencyHKD  = "HKD"  // 港币
)

// Trade 交易记录 - 采用借贷记账思想
// 每一笔交易记录资金的流入流出
// 买入：资产增加（+Amount），现金减少（-Total）
// 卖出：资产减少（-Amount），现金增加（+Total）
// 充值：现金增加（+Amount）
type Trade struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	UserID     uint           `gorm:"index;not null" json:"user_id"`
	AssetType  string         `gorm:"size:20;not null;default:'crypto'" json:"asset_type"` // crypto/a_stock/us_stock/hk_stock
	Symbol     string         `gorm:"size:20;not null" json:"symbol"`                      // 交易代码，如 BTC、600519、AAPL
	Type       string         `gorm:"size:10;not null" json:"type"`                        // buy:买入, sell:卖出, recharge:充值
	Amount     float64        `gorm:"type:decimal(20,8);not null" json:"amount"`           // 数量（买入/卖出/充值的数量）
	Price      float64        `gorm:"type:decimal(20,8);not null" json:"price"`            // 单价（原始货币）
	Total      float64        `gorm:"type:decimal(20,8);not null" json:"total"`            // 总额 = Amount * Price
	Currency   string         `gorm:"size:10;not null;default:'USDT'" json:"currency"`     // 计价货币：USDT/CNY/USD/HKD
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// Holding 资产持仓 - 由交易记录间接计算得出
type Holding struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	AssetType string         `gorm:"size:20;not null;default:'crypto'" json:"asset_type"` // 资产类型
	Symbol    string         `gorm:"size:20;not null" json:"symbol"`
	Amount    float64        `gorm:"type:decimal(20,8);not null" json:"amount"`           // 当前持仓量
	Currency  string         `gorm:"size:10;not null;default:'USDT'" json:"currency"`     // 计价货币
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ExchangeRate 汇率记录
type ExchangeRate struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	From      string    `gorm:"size:10;not null" json:"from"`                        // 源货币
	To        string    `gorm:"size:10;not null" json:"to"`                          // 目标货币
	Rate      float64   `gorm:"type:decimal(20,8);not null" json:"rate"`             // 汇率
	UpdatedAt time.Time `json:"updated_at"`
}

// AssetConfig 资产配置
type AssetConfig struct {
	Symbol      string  `json:"symbol"`       // 资产代码
	Name        string  `json:"name"`         // 资产名称
	AssetType   string  `json:"asset_type"`   // 资产类型
	Currency    string  `json:"currency"`     // 计价货币
	Color       string  `json:"color"`        // 显示颜色
	Icon        string  `json:"icon"`         // 图标
	PriceSource string  `json:"price_source"` // 价格数据源
}

// GetCurrencyByAssetType 根据资产类型获取默认货币
func GetCurrencyByAssetType(assetType string) string {
	switch assetType {
	case AssetTypeCrypto:
		return CurrencyUSDT
	case AssetTypeAStock:
		return CurrencyCNY
	case AssetTypeUSStock:
		return CurrencyUSD
	case AssetTypeHKStock:
		return CurrencyHKD
	default:
		return CurrencyUSDT
	}
}
