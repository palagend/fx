package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"

	"gitee.com/palagend/fx/config"
	"gitee.com/palagend/fx/models"
	"gorm.io/gorm"
)



// ========== 辅助函数 ==========

// getOrCreateHolding 获取或创建持仓
func getOrCreateHolding(tx *gorm.DB, userID uint, symbol string, assetType string) (*models.Holding, error) {
	var holding models.Holding
	err := tx.Where("user_id = ? AND symbol = ? AND asset_type = ?", userID, symbol, assetType).First(&holding).Error
	if err != nil {
		currency := models.GetCurrencyByAssetType(assetType)
		holding = models.Holding{
			UserID:    userID,
			Symbol:    symbol,
			AssetType: assetType,
			Amount:    0,
			Currency:  currency,
		}
		err = tx.Create(&holding).Error
		return &holding, err
	}
	return &holding, nil
}

// updateHolding 更新持仓数量
func updateHolding(tx *gorm.DB, holding *models.Holding, delta float64) error {
	holding.Amount += delta
	return tx.Save(holding).Error
}

// ========== 请求/响应结构 ==========

// 支持的加密货币列表
var supportedCryptos = map[string]bool{
	"BTC":  true,
	"ETH":  true,
	"BNB":  true,
	"XRP":  true,
	"ADA":  true,
	"SOL":  true,
	"DOGE": true,
	"TRX":  true,
	"AVAX": true,
	"HYPE": true,
	"POL":  true,
	"DOT":  true,
}

// 支持的美股列表
var supportedUSStocks = map[string]bool{
	"AAPL":   true,
	"MSFT":   true,
	"GOOG":   true,
	"AMZN":   true,
	"TSLA":   true,
	"META":   true,
	"NVDA":   true,
	"BABA":   true,
	"ORCL":   true,
	"CRCL":   true,
	"MSTR":   true,
	"QQQI":   true,
	"TCEHY":  true,
	"PURR":   true,
	"QQQ":    true,
}

// BusinessError 业务错误类型
type BusinessError struct {
	Message string
}

func (e *BusinessError) Error() string {
	return e.Message
}

type CreateTradeRequest struct {
	AssetType string  `json:"asset_type" binding:"required,oneof=crypto us_stock cash"`
	Symbol    string  `json:"symbol" binding:"required"`
	Type      string  `json:"type" binding:"required,oneof=buy sell recharge"`
	Amount    float64 `json:"amount" binding:"required,gt=0,lte=1000000000"`
	Price     float64 `json:"price" binding:"required,gt=0,lte=1000000000"`
}

// validateTradeRequest 对交易请求进行业务校验
func validateTradeRequest(req *CreateTradeRequest) error {
	switch req.Type {
	case "recharge":
		// 充值只能是USD，资产类型必须是cash
		if req.AssetType != "cash" {
			return fmt.Errorf("充值资产类型必须是cash")
		}
		if req.Symbol != "USD" {
			return fmt.Errorf("充值只支持USD")
		}
		if req.Price != 1 {
			return fmt.Errorf("充值价格必须为1")
		}
	case "buy", "sell":
		// 检查资产类型和代码
		switch req.AssetType {
		case "crypto":
			if !supportedCryptos[req.Symbol] {
				return fmt.Errorf("不支持的加密货币: %s", req.Symbol)
			}
		case "us_stock":
			if !supportedUSStocks[req.Symbol] {
				return fmt.Errorf("不支持的美股: %s", req.Symbol)
			}
		default:
			return fmt.Errorf("买卖交易不支持资产类型: %s", req.AssetType)
		}
	}
	return nil
}

type TradeResponse struct {
	ID        uint    `json:"id"`
	UUID      string  `json:"uuid"`
	AssetType string  `json:"asset_type"`
	Symbol    string  `json:"symbol"`
	Type      string  `json:"type"`
	Amount    float64 `json:"amount"`
	Price     float64 `json:"price"`
	Total     float64 `json:"total"`
	Currency  string  `json:"currency"`
	CreatedAt string  `json:"created_at"`
}

// ========== 交易接口 ==========

// CreateTrade 创建交易记录
func CreateTrade(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req CreateTradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 业务层参数校验
	if err := validateTradeRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := config.GetDB()
	uid := userID.(uint)
	total := req.Amount * req.Price

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var err error
	switch req.Type {
	case "recharge":
		// 充值USD到现金账户
		err = handleRecharge(tx, uid, req.Amount)
	case "buy":
		err = handleBuy(tx, uid, req.Symbol, req.Amount, total, req.AssetType)
	case "sell":
		err = handleSell(tx, uid, req.Symbol, req.Amount, total, req.AssetType)
	}

	if err != nil {
		tx.Rollback()
		if _, ok := err.(*BusinessError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// 创建交易记录
	currency := models.GetCurrencyByAssetType(req.AssetType)
	trade := models.Trade{
		UUID:      uuid.New().String(),
		UserID:    uid,
		AssetType: req.AssetType,
		Symbol:    req.Symbol,
		Type:      req.Type,
		Amount:    req.Amount,
		Price:     req.Price,
		Total:     total,
		Currency:  currency,
	}
	if err := tx.Create(&trade).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建交易记录失败"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交交易失败"})
		return
	}

	c.JSON(http.StatusCreated, TradeResponse{
		ID:        trade.ID,
		UUID:      trade.UUID,
		AssetType: trade.AssetType,
		Symbol:    trade.Symbol,
		Type:      trade.Type,
		Amount:    trade.Amount,
		Price:     trade.Price,
		Total:     trade.Total,
		Currency:  trade.Currency,
		CreatedAt: trade.CreatedAt.Format("2006-01-02 15:04:05"),
	})
}

// handleRecharge 处理充值（统一为USD）
func handleRecharge(tx *gorm.DB, uid uint, amount float64) error {
	cashHolding, err := getOrCreateHolding(tx, uid, "USD", "cash")
	if err != nil {
		return fmt.Errorf("创建现金持仓失败")
	}

	if err := updateHolding(tx, cashHolding, amount); err != nil {
		return fmt.Errorf("更新现金持仓失败")
	}
	return nil
}

// handleBuy 处理买入（统一使用USD现金）
func handleBuy(tx *gorm.DB, uid uint, symbol string, amount, total float64, assetType string) error {
	// 检查USD现金余额
	cashHolding, err := getOrCreateHolding(tx, uid, "USD", "cash")
	if err != nil {
		return fmt.Errorf("获取现金持仓失败")
	}

	if cashHolding.Amount < total {
		return &BusinessError{Message: "USD现金余额不足"}
	}

	// 减少USD现金
	if err := updateHolding(tx, cashHolding, -total); err != nil {
		return fmt.Errorf("更新现金持仓失败")
	}

	// 增加资产
	assetHolding, err := getOrCreateHolding(tx, uid, symbol, assetType)
	if err != nil {
		return fmt.Errorf("创建持仓失败")
	}

	if err := updateHolding(tx, assetHolding, amount); err != nil {
		return fmt.Errorf("更新持仓失败")
	}
	return nil
}

// handleSell 处理卖出（统一得到USD现金）
func handleSell(tx *gorm.DB, uid uint, symbol string, amount, total float64, assetType string) error {
	// 检查持仓
	assetHolding, err := getOrCreateHolding(tx, uid, symbol, assetType)
	if err != nil {
		return fmt.Errorf("获取持仓失败")
	}

	if assetHolding.Amount < amount {
		return &BusinessError{Message: "持仓不足"}
	}

	// 减少资产
	if err := updateHolding(tx, assetHolding, -amount); err != nil {
		return fmt.Errorf("更新持仓失败")
	}

	// 增加USD现金
	cashHolding, err := getOrCreateHolding(tx, uid, "USD", "cash")
	if err != nil {
		return fmt.Errorf("获取现金持仓失败")
	}

	if err := updateHolding(tx, cashHolding, total); err != nil {
		return fmt.Errorf("更新现金持仓失败")
	}
	return nil
}

// GetTrades 获取交易记录
func GetTrades(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	db := config.GetDB()
	uid := userID.(uint)

	var trades []models.Trade
	if err := db.Where("user_id = ?", uid).Order("created_at DESC").Find(&trades).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取交易记录失败"})
		return
	}

	response := make([]TradeResponse, len(trades))
	for i, t := range trades {
		response[i] = TradeResponse{
			ID:        t.ID,
			UUID:      t.UUID,
			AssetType: t.AssetType,
			Symbol:    t.Symbol,
			Type:      t.Type,
			Amount:    t.Amount,
			Price:     t.Price,
			Total:     t.Total,
			Currency:  t.Currency,
			CreatedAt: t.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(http.StatusOK, gin.H{"trades": response})
}

// DeleteTrade 删除交易记录
func DeleteTrade(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	tradeID := c.Param("id")
	db := config.GetDB()
	uid := userID.(uint)

	var trade models.Trade
	if err := db.Where("id = ? AND user_id = ?", tradeID, uid).First(&trade).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "交易记录不存在"})
		return
	}

	tx := db.Begin()

	// 获取该用户所有交易（排除要删除的这条）
	var remainingTrades []models.Trade
	if err := tx.Where("user_id = ? AND id != ?", uid, tradeID).Order("created_at asc").Find(&remainingTrades).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取交易记录失败"})
		return
	}

	// 计算删除后的持仓状态
	simulatedHoldings := make(map[string]map[string]float64) // assetType -> symbol -> amount
	cashBalance := 0.0

	for _, t := range remainingTrades {
		if t.Type == "recharge" {
			cashBalance += t.Amount
			continue
		}
		if simulatedHoldings[t.AssetType] == nil {
			simulatedHoldings[t.AssetType] = make(map[string]float64)
		}
		switch t.Type {
		case "buy":
			simulatedHoldings[t.AssetType][t.Symbol] += t.Amount
			cashBalance -= t.Total
		case "sell":
			simulatedHoldings[t.AssetType][t.Symbol] -= t.Amount
			cashBalance += t.Total
		}
	}

	// 校验删除后不能导致负持仓
	for assetType, holdings := range simulatedHoldings {
		for symbol, amount := range holdings {
			if amount < 0 {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("删除该交易会导致 %s(%s) 持仓为负数(%.8f)，无法删除", symbol, assetType, amount),
				})
				return
			}
		}
	}

	// 校验删除后不能导致现金余额为负
	if cashBalance < 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("删除该交易会导致 USD 现金余额为负数(%.2f)，无法删除", cashBalance),
		})
		return
	}

	// 执行物理删除
	if err := tx.Unscoped().Delete(&trade).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除交易记录失败"})
		return
	}

	// 重新计算所有持仓
	if err := recalcAllHoldings(tx, uid); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重新计算持仓失败"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交删除操作失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "交易记录已删除",
		"deleted_trade": gin.H{
			"id":         trade.ID,
			"asset_type": trade.AssetType,
			"symbol":     trade.Symbol,
			"type":       trade.Type,
			"amount":     trade.Amount,
			"price":      trade.Price,
			"created_at": trade.CreatedAt,
		},
	})
}
