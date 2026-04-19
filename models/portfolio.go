package models

import (
	"time"

	"gorm.io/gorm"
)

// Trade 交易记录 - 采用借贷记账思想
// 每一笔交易记录资金的流入流出
// 买入：加密资产增加（+Amount），USDT减少（-Total）
// 卖出：加密资产减少（-Amount），USDT增加（+Total）
// 充值：USDT增加（+Amount）
type Trade struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	Symbol    string         `gorm:"size:20;not null" json:"symbol"`            // 交易对，如 BTC、ETH、USDT
	Type      string         `gorm:"size:10;not null" json:"type"`              // buy:买入, sell:卖出, recharge:充值
	Amount    float64        `gorm:"type:decimal(20,8);not null" json:"amount"` // 数量（买入/卖出/充值的数量）
	Price     float64        `gorm:"type:decimal(20,8);not null" json:"price"`  // 单价
	Total     float64        `gorm:"type:decimal(20,8);not null" json:"total"`  // 总额 = Amount * Price
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Holding 资产持仓 - 由交易记录间接计算得出
// 不存储成本价，成本价通过 USDT净投入/持仓量 计算
type Holding struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	Symbol    string         `gorm:"size:20;not null" json:"symbol"`
	Amount    float64        `gorm:"type:decimal(20,8);not null" json:"amount"` // 当前持仓量
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}


