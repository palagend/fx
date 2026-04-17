package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gitee.com/palagend/fx/config"
	"gitee.com/palagend/fx/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 辅助函数：获取持仓
func getHolding(tx *gorm.DB, userID uint, symbol string) (*models.Holding, error) {
	var holding models.Holding
	err := tx.Where("user_id = ? AND symbol = ?", userID, symbol).First(&holding).Error
	return &holding, err
}

// 辅助函数：获取或创建持仓
func getOrCreateHolding(tx *gorm.DB, userID uint, symbol string, defaultCost float64) (*models.Holding, error) {
	holding, err := getHolding(tx, userID, symbol)
	if err != nil {
		holding = &models.Holding{UserID: userID, Symbol: symbol, AvgCost: defaultCost}
		err = tx.Create(holding).Error
	}
	return holding, err
}

// 辅助函数：更新持仓数量（正数增加，负数减少）
func updateHoldingAmount(tx *gorm.DB, holding *models.Holding, delta float64) error {
	holding.Amount += delta
	if holding.Amount <= 0 {
		return tx.Delete(holding).Error
	}
	return tx.Save(holding).Error
}

// 辅助函数：更新加密资产持仓（买入时计算新平均成本）
func updateCryptoHolding(tx *gorm.DB, userID uint, symbol string, amount, price float64) error {
	holding, err := getHolding(tx, userID, symbol)
	if err != nil {
		// 新建持仓
		return tx.Create(&models.Holding{
			UserID:  userID,
			Symbol:  symbol,
			Amount:  amount,
			AvgCost: price,
		}).Error
	}
	// 更新平均成本
	totalCost := holding.Amount*holding.AvgCost + amount*price
	totalAmount := holding.Amount + amount
	holding.Amount = totalAmount
	holding.AvgCost = totalCost / totalAmount
	return tx.Save(holding).Error
}

// CreateTradeRequest 创建交易请求
type CreateTradeRequest struct {
	Symbol string  `json:"symbol" binding:"required"`
	Type   string  `json:"type" binding:"required,oneof=buy sell recharge"`
	Amount float64 `json:"amount" binding:"required,gt=0"`
	Price  float64 `json:"price" binding:"required,gt=0"`
}

// GetTradesResponse 获取交易记录响应
type GetTradesResponse struct {
	ID         uint    `json:"id"`
	Symbol     string  `json:"symbol"`
	Type       string  `json:"type"`
	Amount     float64 `json:"amount"`
	Price      float64 `json:"price"`
	Total      float64 `json:"total"`
	RealizedPL float64 `json:"realized_pl"`
	CreatedAt  string  `json:"created_at"`
}

// HoldingResponse 持仓响应
type HoldingResponse struct {
	ID        uint    `json:"id"`
	Symbol    string  `json:"symbol"`
	Amount    float64 `json:"amount"`
	AvgCost   float64 `json:"avg_cost"`
	CreatedAt string  `json:"created_at"`
}

// PortfolioSummaryResponse 资产组合摘要响应 - 仅返回实现盈亏，其他数据由前端计算
type PortfolioSummaryResponse struct {
	RealizedProfitLoss float64           `json:"realized_profit_loss"`
	Holdings           []HoldingResponse `json:"holdings"`
}

// CreateTrade 创建交易记录 - 支持USDT余额管理
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

	db := config.GetDB()
	uid := userID.(uint)
	total := req.Amount * req.Price

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	usdtHolding, usdtErr := getOrCreateHolding(tx, uid, "USDT", 1.0)

	switch req.Type {
	case "recharge":
		if err := updateHoldingAmount(tx, usdtHolding, req.Amount); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

	case "buy":
		if usdtErr != nil || usdtHolding.Amount < total {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "USDT余额不足"})
			return
		}

		if err := updateHoldingAmount(tx, usdtHolding, -total); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := updateCryptoHolding(tx, uid, req.Symbol, req.Amount, req.Price); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

	case "sell":
		holding, err := getHolding(tx, uid, req.Symbol)
		if err != nil || holding.Amount < req.Amount {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "持仓不足"})
			return
		}

		realizedPL := (req.Price - holding.AvgCost) * req.Amount

		if err := updateHoldingAmount(tx, holding, -req.Amount); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := updateHoldingAmount(tx, usdtHolding, total); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		_ = realizedPL
	}

	trade := models.Trade{
		UserID: uid, Symbol: req.Symbol, Type: req.Type,
		Amount: req.Amount, Price: req.Price, Total: total,
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

	c.JSON(http.StatusCreated, gin.H{
		"message": "交易成功",
		"trade": gin.H{
			"id": trade.ID, "symbol": trade.Symbol, "type": trade.Type,
			"amount": trade.Amount, "price": trade.Price, "total": trade.Total,
			"created_at": trade.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	})
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

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	// 交易类型过滤
	tradeType := c.Query("type")

	var trades []models.Trade
	var total int64

	query := db.Model(&models.Trade{}).Where("user_id = ?", uid)
	if tradeType != "" {
		query = query.Where("type = ?", tradeType)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&trades).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取交易记录失败"})
		return
	}

	// 格式化响应
	var response []GetTradesResponse
	for _, trade := range trades {
		response = append(response, GetTradesResponse{
			ID:         trade.ID,
			Symbol:     trade.Symbol,
			Type:       trade.Type,
			Amount:     trade.Amount,
			Price:      trade.Price,
			Total:      trade.Total,
			RealizedPL: trade.RealizedPL,
			CreatedAt:  trade.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"trades": response,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetHoldings 获取持仓列表
func GetHoldings(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	db := config.GetDB()
	uid := userID.(uint)

	var holdings []models.Holding
	if err := db.Where("user_id = ?", uid).Find(&holdings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取持仓失败"})
		return
	}

	var response []HoldingResponse
	for _, holding := range holdings {
		response = append(response, HoldingResponse{
			ID:        holding.ID,
			Symbol:    holding.Symbol,
			Amount:    holding.Amount,
			AvgCost:   holding.AvgCost,
			CreatedAt: holding.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"holdings": response,
	})
}

// GetPortfolioSummary 获取资产组合摘要 - 仅返回实现盈亏和持仓列表
func GetPortfolioSummary(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	db := config.GetDB()
	uid := userID.(uint)

	// 获取所有持仓
	var holdings []models.Holding
	if err := db.Where("user_id = ?", uid).Find(&holdings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取持仓失败"})
		return
	}

	// 计算实现盈亏（仅由后端计算，因为涉及历史交易记录）
	var totalRealizedPL float64
	db.Model(&models.Trade{}).Where("user_id = ? AND type = ?", uid, "sell").Select("COALESCE(SUM(realized_pl), 0)").Scan(&totalRealizedPL)

	// 格式化持仓响应
	var holdingsResponse []HoldingResponse
	for _, holding := range holdings {
		holdingsResponse = append(holdingsResponse, HoldingResponse{
			ID:        holding.ID,
			Symbol:    holding.Symbol,
			Amount:    holding.Amount,
			AvgCost:   holding.AvgCost,
			CreatedAt: holding.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, PortfolioSummaryResponse{
		RealizedProfitLoss: totalRealizedPL,
		Holdings:           holdingsResponse,
	})
}

// DeleteTrade 删除交易记录 - 会重新计算相关资产的持仓和USDT余额
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

	if err := tx.Delete(&trade).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除交易记录失败"})
		return
	}

	// 重新计算该资产持仓
	if trade.Symbol != "USDT" {
		if err := recalcHolding(tx, uid, trade.Symbol); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// 重新计算USDT余额
	if err := recalcUSDT(tx, uid); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交删除操作失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "交易记录已删除，持仓已重新计算"})
}

// 重新计算资产持仓
func recalcHolding(tx *gorm.DB, uid uint, symbol string) error {
	tx.Where("user_id = ? AND symbol = ?", uid, symbol).Delete(&models.Holding{})

	var trades []models.Trade
	if err := tx.Where("user_id = ? AND symbol = ? AND type IN ?", uid, symbol, []string{"buy", "sell"}).
		Order("created_at ASC").Find(&trades).Error; err != nil {
		return fmt.Errorf("获取交易记录失败")
	}

	var amount, cost float64
	for _, t := range trades {
		if t.Type == "buy" {
			cost += t.Amount * t.Price
			amount += t.Amount
		} else {
			if amount > 0 {
				cost = cost * (amount - t.Amount) / amount
			}
			amount -= t.Amount
		}
	}

	if amount > 0 {
		return tx.Create(&models.Holding{UserID: uid, Symbol: symbol, Amount: amount, AvgCost: cost / amount}).Error
	}
	return nil
}

// 重新计算USDT余额
func recalcUSDT(tx *gorm.DB, uid uint) error {
	tx.Where("user_id = ? AND symbol = ?", uid, "USDT").Delete(&models.Holding{})

	var recharge, buyTotal, sellTotal float64
	tx.Model(&models.Trade{}).Where("user_id = ? AND symbol = ? AND type = ?", uid, "USDT", "recharge").
		Select("COALESCE(SUM(amount), 0)").Scan(&recharge)
	tx.Model(&models.Trade{}).Where("user_id = ? AND type = ?", uid, "buy").
		Select("COALESCE(SUM(total), 0)").Scan(&buyTotal)
	tx.Model(&models.Trade{}).Where("user_id = ? AND type = ?", uid, "sell").
		Select("COALESCE(SUM(total), 0)").Scan(&sellTotal)

	balance := recharge - buyTotal + sellTotal
	if balance > 0 {
		return tx.Create(&models.Holding{UserID: uid, Symbol: "USDT", Amount: balance, AvgCost: 1.0}).Error
	}
	return nil
}

// ClearTrades 清空所有交易记录
func ClearTrades(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	db := config.GetDB()
	uid := userID.(uint)

	// 删除该用户的所有交易记录
	if err := db.Where("user_id = ?", uid).Delete(&models.Trade{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清空交易记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "交易记录已清空"})
}

// GetAssetPrice 获取资产价格（模拟接口，实际应该对接交易所API）
func GetAssetPrice(c *gin.Context) {
	symbol := c.Param("symbol")

	// 模拟价格数据，实际应该对接交易所API
	prices := map[string]float64{
		"BTC":  65000.0,
		"ETH":  3500.0,
		"BNB":  600.0,
		"XRP":  0.6,
		"ADA":  0.5,
		"SOL":  150.0,
		"DOGE": 0.15,
		"TRX":  0.12,
		"AVAX": 35.0,
		"HYPE": 10.0,
		"USDT": 1.0,
	}

	price, exists := prices[symbol]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "不支持的资产"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"symbol":     symbol,
		"price":      price,
		"updated_at": time.Now().Format("2006-01-02 15:04:05"),
	})
}

// GetAllPrices 获取所有资产价格 - 从CoinCap获取实时价格
func GetAllPrices(c *gin.Context) {
	// 支持的资产列表
	symbols := []string{"BTC", "ETH", "BNB", "XRP", "ADA", "SOL", "DOGE", "TRX", "AVAX"}
	symbolsParam := strings.Join(symbols, ",")

	// 调用CoinCap API
	url := fmt.Sprintf("https://rest.coincap.io/v3/assets?ids=%s", symbolsParam)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建请求失败"})
		return
	}

	// 添加Authorization头
	req.Header.Add("Authorization", "Bearer b617d9cf029dbb40f02b058a0e74919176b768cf36fd1ea6fae55a13a1610f41")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取价格失败"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "价格服务暂时不可用"})
		return
	}

	var result struct {
		Data []struct {
			Symbol            string `json:"symbol"`
			PriceUsd          string `json:"priceUsd"`
			ChangePercent24Hr string `json:"changePercent24Hr"`
		} `json:"data"`
		Timestamp int64 `json:"timestamp"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解析价格数据失败"})
		return
	}

	// 构建价格映射
	prices := make(map[string]float64)
	priceChanges := make(map[string]float64)

	for _, item := range result.Data {
		price, _ := strconv.ParseFloat(item.PriceUsd, 64)
		change24h, _ := strconv.ParseFloat(item.ChangePercent24Hr, 64)
		prices[item.Symbol] = price
		priceChanges[item.Symbol] = change24h
	}

	// USDT固定为1.0
	prices["USDT"] = 1.0
	priceChanges["USDT"] = 0

	c.JSON(http.StatusOK, gin.H{
		"prices":        prices,
		"price_changes": priceChanges,
		"updated_at":    time.Now().Format("2006-01-02 15:04:05"),
	})
}

// ImportDataRequest 导入数据请求
type ImportDataRequest struct {
	Portfolio          []HoldingData `json:"portfolio" binding:"required"`
	Trades             []TradeData   `json:"trades" binding:"required"`
	RealizedProfitLoss float64       `json:"realized_profit_loss"`
}

// HoldingData 导入的持仓数据
type HoldingData struct {
	Symbol  string  `json:"symbol" binding:"required"`
	Amount  float64 `json:"amount" binding:"required,gte=0"`
	AvgCost float64 `json:"avg_cost" binding:"required,gte=0"`
}

// TradeData 导入的交易数据
type TradeData struct {
	Symbol     string  `json:"symbol" binding:"required"`
	Type       string  `json:"type" binding:"required,oneof=buy sell recharge"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Price      float64 `json:"price" binding:"required,gt=0"`
	Total      float64 `json:"total"`
	RealizedPL float64 `json:"realized_pl"`
	CreatedAt  string  `json:"created_at"`
}

// ImportData 导入数据（覆盖式）
func ImportData(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req ImportDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := config.GetDB()
	uid := userID.(uint)

	// 开启事务
	tx := db.Begin()

	// 1. 删除该用户的所有持仓
	if err := tx.Where("user_id = ?", uid).Delete(&models.Holding{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清空持仓失败"})
		return
	}

	// 2. 删除该用户的所有交易记录
	if err := tx.Where("user_id = ?", uid).Delete(&models.Trade{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清空交易记录失败"})
		return
	}

	// 3. 导入持仓数据
	for _, h := range req.Portfolio {
		holding := models.Holding{
			UserID:  uid,
			Symbol:  h.Symbol,
			Amount:  h.Amount,
			AvgCost: h.AvgCost,
		}
		if err := tx.Create(&holding).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "导入持仓失败"})
			return
		}
	}

	// 4. 导入交易数据
	for _, t := range req.Trades {
		// 解析时间
		createdAt := time.Now()
		if t.CreatedAt != "" {
			if parsedTime, err := time.Parse("2006-01-02 15:04:05", t.CreatedAt); err == nil {
				createdAt = parsedTime
			} else if parsedTime, err := time.Parse(time.RFC3339, t.CreatedAt); err == nil {
				createdAt = parsedTime
			}
		}

		// 计算总额
		total := t.Total
		if total == 0 {
			total = t.Amount * t.Price
		}

		trade := models.Trade{
			UserID:     uid,
			Symbol:     t.Symbol,
			Type:       t.Type,
			Amount:     t.Amount,
			Price:      t.Price,
			Total:      total,
			RealizedPL: t.RealizedPL,
			CreatedAt:  createdAt,
		}
		if err := tx.Create(&trade).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "导入交易记录失败"})
			return
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交数据失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "数据导入成功",
		"stats": gin.H{
			"holdings_imported": len(req.Portfolio),
			"trades_imported":   len(req.Trades),
		},
	})
}

// ExportData 导出所有数据
func ExportData(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	db := config.GetDB()
	uid := userID.(uint)

	// 获取所有持仓
	var holdings []models.Holding
	if err := db.Where("user_id = ?", uid).Find(&holdings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取持仓失败"})
		return
	}

	// 获取所有交易记录
	var trades []models.Trade
	if err := db.Where("user_id = ?", uid).Order("created_at DESC").Find(&trades).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取交易记录失败"})
		return
	}

	// 计算实现盈亏
	var totalRealizedPL float64
	db.Model(&models.Trade{}).Where("user_id = ? AND type = ?", uid, "sell").Select("COALESCE(SUM(realized_pl), 0)").Scan(&totalRealizedPL)

	// 格式化持仓数据
	var holdingsData []HoldingData
	for _, h := range holdings {
		holdingsData = append(holdingsData, HoldingData{
			Symbol:  h.Symbol,
			Amount:  h.Amount,
			AvgCost: h.AvgCost,
		})
	}

	// 格式化交易数据
	var tradesData []TradeData
	for _, t := range trades {
		tradesData = append(tradesData, TradeData{
			Symbol:     t.Symbol,
			Type:       t.Type,
			Amount:     t.Amount,
			Price:      t.Price,
			Total:      t.Total,
			RealizedPL: t.RealizedPL,
			CreatedAt:  t.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"version":              "2.0",
		"export_time":          time.Now().Format("2006-01-02 15:04:05"),
		"portfolio":            holdingsData,
		"trades":               tradesData,
		"realized_profit_loss": totalRealizedPL,
	})
}
