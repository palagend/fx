package models

import (
	"time"

	"gorm.io/gorm"
)

// Trade 交易记录
type Trade struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	UserID       uint           `gorm:"index;not null" json:"user_id"`
	Symbol       string         `gorm:"size:20;not null" json:"symbol"`
	Type         string         `gorm:"size:10;not null" json:"type"` // buy, sell, recharge
	Amount       float64        `gorm:"type:decimal(20,8);not null" json:"amount"`
	Price        float64        `gorm:"type:decimal(20,8);not null" json:"price"`
	Total        float64        `gorm:"type:decimal(20,8);not null" json:"total"`
	RealizedPL   float64        `gorm:"type:decimal(20,8);default:0" json:"realized_pl"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// Holding 资产持仓
type Holding struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	UserID       uint           `gorm:"index;not null" json:"user_id"`
	Symbol       string         `gorm:"size:20;not null" json:"symbol"`
	Amount       float64        `gorm:"type:decimal(20,8);not null" json:"amount"`
	AvgCost      float64        `gorm:"type:decimal(20,8);not null" json:"avg_cost"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// PortfolioSummary 资产组合摘要
type PortfolioSummary struct {
	UserID            uint    `json:"user_id"`
	TotalValue        float64 `json:"total_value"`
	USDTBalance       float64 `json:"usdt_balance"`
	RealizedProfitLoss float64 `json:"realized_profit_loss"`
}

func init() {
	// 注册模型到自动迁移
	autoMigrateFuncs = append(autoMigrateFuncs, func(db *gorm.DB) error {
		return db.AutoMigrate(&Trade{}, &Holding{})
	})
}

var autoMigrateFuncs []func(*gorm.DB) error

// RunAutoMigrations 执行所有模型的自动迁移
func RunAutoMigrations(db *gorm.DB) error {
	for _, fn := range autoMigrateFuncs {
		if err := fn(db); err != nil {
			return err
		}
	}
	return nil
}
