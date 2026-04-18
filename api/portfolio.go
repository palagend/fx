package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"log"

	"gitee.com/palagend/fx/config"
	"gitee.com/palagend/fx/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ========== 辅助函数 ==========

// getOrCreateHolding 获取或创建持仓
func getOrCreateHolding(tx *gorm.DB, userID uint, symbol string) (*models.Holding, error) {
	var holding models.Holding
	err := tx.Where("user_id = ? AND symbol = ?", userID, symbol).First(&holding).Error
	if err != nil {
		holding = models.Holding{UserID: userID, Symbol: symbol, Amount: 0}
		err = tx.Create(&holding).Error
		log.Printf("创建持仓失败: %v", err)
		return &holding, err
	}
	return &holding, nil
}

// getOrCreateInvestment 获取或创建投资记录
func getOrCreateInvestment(tx *gorm.DB, userID uint, symbol string) (*models.Investment, error) {
	var inv models.Investment
	err := tx.Where("user_id = ? AND symbol = ?", userID, symbol).First(&inv).Error
	if err != nil {
		inv = models.Investment{UserID: userID, Symbol: symbol, TotalIn: 0, TotalOut: 0, RealizedPL: 0}
		err = tx.Create(&inv).Error
		return &inv, err
	}
	return &inv, nil
}

// updateHolding 更新持仓数量
func updateHolding(tx *gorm.DB, holding *models.Holding, delta float64) error {
	holding.Amount += delta
	if holding.Amount < 0 {
		return tx.Delete(holding).Error
	}
	return tx.Save(holding).Error
}

// calcAvgCost 计算平均成本 = USDT净投入 / 持仓量
func calcAvgCost(inv *models.Investment, holding *models.Holding) float64 {
	if holding.Amount <= 0 {
		return 0
	}
	netInvestment := inv.TotalIn - inv.TotalOut
	if netInvestment <= 0 {
		return 0
	}
	return netInvestment / holding.Amount
}

// ========== 请求/响应结构 ==========

type CreateTradeRequest struct {
	Symbol string  `json:"symbol" binding:"required"`
	Type   string  `json:"type" binding:"required,oneof=buy sell recharge"`
	Amount float64 `json:"amount" binding:"required,gt=0"`
	Price  float64 `json:"price" binding:"required,gt=0"`
}

type TradeResponse struct {
	ID        uint    `json:"id"`
	Symbol    string  `json:"symbol"`
	Type      string  `json:"type"`
	Amount    float64 `json:"amount"`
	Price     float64 `json:"price"`
	Total     float64 `json:"total"`
	CreatedAt string  `json:"created_at"`
}

type HoldingResponse struct {
	ID      uint    `json:"id"`
	Symbol  string  `json:"symbol"`
	Amount  float64 `json:"amount"`
	AvgCost float64 `json:"avg_cost"` // 实时计算
	Value   float64 `json:"value"`    // 实时计算
	Profit  float64 `json:"profit"`   // 实时计算
}

type InvestmentResponse struct {
	Symbol     string  `json:"symbol"`
	TotalIn    float64 `json:"total_in"`
	TotalOut   float64 `json:"total_out"`
	RealizedPL float64 `json:"realized_pl"`
}

// ========== 交易接口 ==========

// CreateTrade 创建交易记录
// 借贷记账思想：
// - 充值：USDT增加
// - 买入：加密资产增加，USDT减少（投入）
// - 卖出：加密资产减少，USDT增加（退出），实现盈亏 = 退出 - 投入
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

	switch req.Type {
	case "recharge":
		// USDT充值：USDT持仓增加
		usdtHolding, err := getOrCreateHolding(tx, uid, "USDT")
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建USDT持仓失败"})
			return
		}
		if err := updateHolding(tx, usdtHolding, req.Amount); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新USDT持仓失败"})
			return
		}

	case "buy":
		// 买入：检查USDT余额 -> 减少USDT -> 增加加密资产 -> 记录投入
		usdtHolding, err := getOrCreateHolding(tx, uid, "USDT")
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取USDT持仓失败"})
			return
		}
		if usdtHolding.Amount < total {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "USDT余额不足"})
			return
		}

		// 减少USDT
		if err := updateHolding(tx, usdtHolding, -total); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新USDT持仓失败"})
			return
		}

		// 增加加密资产
		cryptoHolding, err := getOrCreateHolding(tx, uid, req.Symbol)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建持仓失败"})
			return
		}
		if err := updateHolding(tx, cryptoHolding, req.Amount); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新持仓失败"})
			return
		}

		// 记录USDT投入
		inv, err := getOrCreateInvestment(tx, uid, req.Symbol)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建投资记录失败"})
			return
		}
		inv.TotalIn += total
		if err := tx.Save(inv).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "记录投入失败"})
			return
		}

	case "sell":
		// 卖出：检查持仓 -> 减少加密资产 -> 增加USDT -> 记录退出 -> 计算实现盈亏
		cryptoHolding, err := getOrCreateHolding(tx, uid, req.Symbol)
		if err != nil || cryptoHolding.Amount < req.Amount {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "持仓不足"})
			return
		}

		// 减少加密资产
		if err := updateHolding(tx, cryptoHolding, -req.Amount); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新持仓失败"})
			return
		}

		// 增加USDT
		usdtHolding, err := getOrCreateHolding(tx, uid, "USDT")
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取USDT持仓失败"})
			return
		}
		if err := updateHolding(tx, usdtHolding, total); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新USDT持仓失败"})
			return
		}

		// 记录USDT退出并计算实现盈亏
		inv, _ := getOrCreateInvestment(tx, uid, req.Symbol)
		inv.TotalOut += total
		inv.RealizedPL = inv.TotalOut - inv.TotalIn
		if err := tx.Save(inv).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "记录退出失败"})
			return
		}
	}

	// 创建交易记录
	trade := models.Trade{
		UserID: uid,
		Symbol: req.Symbol,
		Type:   req.Type,
		Amount: req.Amount,
		Price:  req.Price,
		Total:  total,
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
		Symbol:    trade.Symbol,
		Type:      trade.Type,
		Amount:    trade.Amount,
		Price:     trade.Price,
		Total:     trade.Total,
		CreatedAt: trade.CreatedAt.Format("2006-01-02 15:04:05"),
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

	var trades []models.Trade
	if err := db.Where("user_id = ?", uid).Order("created_at DESC").Find(&trades).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取交易记录失败"})
		return
	}

	var response []TradeResponse
	for _, t := range trades {
		response = append(response, TradeResponse{
			ID:        t.ID,
			Symbol:    t.Symbol,
			Type:      t.Type,
			Amount:    t.Amount,
			Price:     t.Price,
			Total:     t.Total,
			CreatedAt: t.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{"trades": response})
}

// DeleteTrade 删除交易记录 - 重新计算持仓和投资记录
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

	// 删除交易记录
	if err := tx.Delete(&trade).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除交易记录失败"})
		return
	}

	// 重新计算该资产持仓和投资记录
	if trade.Symbol != "USDT" {
		if err := recalcAsset(tx, uid, trade.Symbol); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// 重新计算USDT持仓
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

// recalcAsset 重新计算资产持仓和投资记录
func recalcAsset(tx *gorm.DB, uid uint, symbol string) error {
	// 清除持仓和投资记录
	tx.Where("user_id = ? AND symbol = ?", uid, symbol).Delete(&models.Holding{})
	tx.Where("user_id = ? AND symbol = ?", uid, symbol).Delete(&models.Investment{})

	// 获取所有交易记录
	var trades []models.Trade
	if err := tx.Where("user_id = ? AND symbol = ?", uid, symbol).Order("created_at ASC").Find(&trades).Error; err != nil {
		return fmt.Errorf("获取交易记录失败")
	}

	var amount, totalIn, totalOut float64
	for _, t := range trades {
		switch t.Type {
		case "buy":
			amount += t.Amount
			totalIn += t.Total
		case "sell":
			amount -= t.Amount
			totalOut += t.Total
		}
	}

	// 创建持仓记录
	if amount > 0 {
		if err := tx.Create(&models.Holding{UserID: uid, Symbol: symbol, Amount: amount}).Error; err != nil {
			return fmt.Errorf("创建持仓失败")
		}
	}

	// 创建投资记录
	if totalIn > 0 || totalOut > 0 {
		inv := models.Investment{
			UserID:     uid,
			Symbol:     symbol,
			TotalIn:    totalIn,
			TotalOut:   totalOut,
			RealizedPL: totalOut - totalIn,
		}
		if err := tx.Create(&inv).Error; err != nil {
			return fmt.Errorf("创建投资记录失败")
		}
	}

	return nil
}

// recalcUSDT 重新计算USDT持仓
func recalcUSDT(tx *gorm.DB, uid uint) error {
	tx.Where("user_id = ? AND symbol = ?", uid, "USDT").Delete(&models.Holding{})

	var recharge, buyTotal, sellTotal float64
	tx.Model(&models.Trade{}).Where("user_id = ? AND type = ?", uid, "recharge").Select("COALESCE(SUM(amount), 0)").Scan(&recharge)
	tx.Model(&models.Trade{}).Where("user_id = ? AND type = ?", uid, "buy").Select("COALESCE(SUM(total), 0)").Scan(&buyTotal)
	tx.Model(&models.Trade{}).Where("user_id = ? AND type = ?", uid, "sell").Select("COALESCE(SUM(total), 0)").Scan(&sellTotal)

	balance := recharge - buyTotal + sellTotal
	if balance > 0 {
		return tx.Create(&models.Holding{UserID: uid, Symbol: "USDT", Amount: balance}).Error
	}
	return nil
}

// ========== 持仓和投资接口 ==========

// GetHoldings 获取所有持仓
// ClearTrades 清空所有交易记录
func ClearTrades(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	db := config.GetDB()
	uid := userID.(uint)

	tx := db.Begin()

	// 删除交易记录
	if err := tx.Where("user_id = ?", uid).Delete(&models.Trade{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清空交易记录失败"})
		return
	}

	// 删除持仓
	if err := tx.Where("user_id = ?", uid).Delete(&models.Holding{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清空持仓失败"})
		return
	}

	// 删除投资记录
	if err := tx.Where("user_id = ?", uid).Delete(&models.Investment{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清空投资记录失败"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交清空操作失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "所有数据已清空"})
}

// ========== 价格接口 ==========

// GetAssetPrice 获取单个资产价格
func GetAssetPrice(c *gin.Context) {
	symbol := c.Param("symbol")

	// USDT 固定价格
	if symbol == "USDT" {
		c.JSON(http.StatusOK, gin.H{
			"symbol": symbol,
			"price":  1.0,
		})
		return
	}

	url := fmt.Sprintf("https://rest.coincap.io/v3/price/bysymbol/%s", symbol)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建请求失败"})
		return
	}

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
		Timestamp int64    `json:"timestamp"`
		Data      []string `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解析价格数据失败"})
		return
	}

	if len(result.Data) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "价格数据为空"})
		return
	}

	price, _ := strconv.ParseFloat(result.Data[0], 64)

	c.JSON(http.StatusOK, gin.H{
		"symbol":     symbol,
		"price":      price,
		"updated_at": result.Timestamp,
	})
}

// GetDashboard 获取仪表盘聚合数据（价格+持仓+统计）
func GetDashboard(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	db := config.GetDB()
	uid := userID.(uint)

	// 获取价格数据
	ids := []string{"bitcoin", "ethereum", "binance-coin", "xrp", "cardano", "solana", "dogecoin", "tron", "avalanche"}
	idsParam := strings.Join(ids, ",")
	url := fmt.Sprintf("https://rest.coincap.io/v3/assets?ids=%s", idsParam)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建请求失败"})
		return
	}
	req.Header.Add("Authorization", "Bearer b617d9cf029dbb40f02b058a0e74919176b768cf36fd1ea6fae55a13a1610f41")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取价格失败"})
		return
	}
	defer resp.Body.Close()

	prices := make(map[string]float64)
	priceChanges := make(map[string]float64)
	var updatedAt int64

	if resp.StatusCode == http.StatusOK {
		var result struct {
			Timestamp int64 `json:"timestamp"`
			Data      []struct {
				Symbol            string `json:"symbol"`
				PriceUsd          string `json:"priceUsd"`
				ChangePercent24Hr string `json:"changePercent24Hr"`
			} `json:"data"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
			for _, item := range result.Data {
				price, _ := strconv.ParseFloat(item.PriceUsd, 64)
				change24h, _ := strconv.ParseFloat(item.ChangePercent24Hr, 64)
				prices[item.Symbol] = price
				priceChanges[item.Symbol] = change24h
			}
			updatedAt = result.Timestamp
		}
	}
	prices["USDT"] = 1.0
	priceChanges["USDT"] = 0

	// 获取持仓
	var holdings []models.Holding
	db.Where("user_id = ?", uid).Find(&holdings)

	// 获取投资记录
	var investments []models.Investment
	db.Where("user_id = ?", uid).Find(&investments)
	invMap := make(map[string]*models.Investment)
	for i := range investments {
		invMap[investments[i].Symbol] = &investments[i]
	}

	// 计算统计数据
	type PortfolioItem struct {
		Symbol       string  `json:"symbol"`
		Amount       float64 `json:"amount"`
		CurrentPrice float64 `json:"current_price"`
		AvgCost      float64 `json:"avg_cost"`
		MarketValue  float64 `json:"market_value"`
		Cost         float64 `json:"cost"`
		ProfitLoss   float64 `json:"profit_loss"`
		PLRate       float64 `json:"pl_rate"`
	}

	var portfolio []PortfolioItem
	var totalValue, usdtBalance, unrealizedPL, totalCost float64
	var totalRealizedPL float64
	var weightedChange float64

	for _, h := range holdings {
		isUSDT := h.Symbol == "USDT"
		price := prices[h.Symbol]
		if isUSDT {
			price = 1
		}

		marketValue := h.Amount * price
		totalValue += marketValue

		avgCost := 0.0
		cost := 0.0
		profitLoss := 0.0
		plRate := 0.0

		if isUSDT {
			usdtBalance = h.Amount
			avgCost = 1
			cost = h.Amount
		} else {
			if inv, ok := invMap[h.Symbol]; ok && h.Amount > 0 {
				netInvestment := inv.TotalIn - inv.TotalOut
				if netInvestment > 0 {
					avgCost = netInvestment / h.Amount
				}
				totalRealizedPL += inv.RealizedPL
			}
			cost = h.Amount * avgCost
			profitLoss = marketValue - cost
			if cost > 0 {
				plRate = (profitLoss / cost) * 100
			}
			unrealizedPL += profitLoss
			totalCost += cost
			change24h := priceChanges[h.Symbol]
			weightedChange += marketValue * change24h
		}

		portfolio = append(portfolio, PortfolioItem{
			Symbol:       h.Symbol,
			Amount:       h.Amount,
			CurrentPrice: price,
			AvgCost:      avgCost,
			MarketValue:  marketValue,
			Cost:         cost,
			ProfitLoss:   profitLoss,
			PLRate:       plRate,
		})
	}

	unrealizedPLRate := 0.0
	if totalCost > 0 {
		unrealizedPLRate = (unrealizedPL / totalCost) * 100
	}

	totalValueChange24h := 0.0
	if totalValue > 0 {
		totalValueChange24h = weightedChange / totalValue
	}

	c.JSON(http.StatusOK, gin.H{
		"prices":                 prices,
		"price_changes":          priceChanges,
		"updated_at":             updatedAt,
		"portfolio":              portfolio,
		"total_value":            totalValue,
		"usdt_balance":           usdtBalance,
		"unrealized_pl":          unrealizedPL,
		"unrealized_pl_rate":     unrealizedPLRate,
		"realized_pl":            totalRealizedPL,
		"total_pl":               unrealizedPL + totalRealizedPL,
		"total_value_change_24h": totalValueChange24h,
	})
}
