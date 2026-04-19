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

// ========== 辅助函数 ==========

// getOrCreateHolding 获取或创建持仓
func getOrCreateHolding(tx *gorm.DB, userID uint, symbol string) (*models.Holding, error) {
	var holding models.Holding
	err := tx.Where("user_id = ? AND symbol = ?", userID, symbol).First(&holding).Error
	if err != nil {
		holding = models.Holding{UserID: userID, Symbol: symbol, Amount: 0}
		err = tx.Create(&holding).Error
		return &holding, err
	}
	return &holding, nil
}

// getOrCreateInvestment 获取或创建投资记录
func getOrCreateInvestment(tx *gorm.DB, userID uint, symbol string) (*models.Investment, error) {
	var inv models.Investment
	err := tx.Where("user_id = ? AND symbol = ?", userID, symbol).First(&inv).Error
	if err != nil {
		inv = models.Investment{UserID: userID, Symbol: symbol, TotalIn: 0, TotalOut: 0}
		err = tx.Create(&inv).Error
		return &inv, err
	}
	return &inv, nil
}

// updateHolding 更新持仓数量
func updateHolding(tx *gorm.DB, holding *models.Holding, delta float64) error {
	holding.Amount += delta
	return tx.Save(holding).Error
}

// ========== 请求/响应结构 ==========

// 支持的加密货币列表（不含USDT）
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
}

// BusinessError 业务错误类型
type BusinessError struct {
	Message string
}

func (e *BusinessError) Error() string {
	return e.Message
}

type CreateTradeRequest struct {
	Symbol string  `json:"symbol" binding:"required"`
	Type   string  `json:"type" binding:"required,oneof=buy sell recharge"`
	Amount float64 `json:"amount" binding:"required,gt=0,lte=1000000000"`
	Price  float64 `json:"price" binding:"required,gt=0,lte=1000000000"`
}

// validateTradeRequest 对交易请求进行业务校验
func validateTradeRequest(req *CreateTradeRequest) error {
	switch req.Type {
	case "recharge":
		// 充值必须是USDT
		if req.Symbol != "USDT" {
			return fmt.Errorf("充值只支持USDT")
		}
		// 充值时价格和数量应该一致（1:1）
		if req.Price != 1 {
			return fmt.Errorf("USDT充值价格必须为1")
		}
	case "buy", "sell":
		// 买卖不能是USDT
		if req.Symbol == "USDT" {
			return fmt.Errorf("不能直接买卖USDT，请使用充值功能")
		}
		// 检查是否是支持的加密货币
		if !supportedCryptos[req.Symbol] {
			return fmt.Errorf("不支持的加密货币: %s", req.Symbol)
		}
	}
	return nil
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
		err = handleRecharge(tx, uid, req.Amount)
	case "buy":
		err = handleBuy(tx, uid, req.Symbol, req.Amount, total)
	case "sell":
		err = handleSell(tx, uid, req.Symbol, req.Amount, total)
	}

	if err != nil {
		tx.Rollback()
		// 区分业务错误和系统错误
		if _, ok := err.(*BusinessError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
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

// handleRecharge 处理USDT充值
func handleRecharge(tx *gorm.DB, uid uint, amount float64) error {
	usdtHolding, err := getOrCreateHolding(tx, uid, "USDT")
	if err != nil {
		return fmt.Errorf("创建USDT持仓失败")
	}

	if err := updateHolding(tx, usdtHolding, amount); err != nil {
		return fmt.Errorf("更新USDT持仓失败")
	}
	return nil
}

// handleBuy 处理买入
func handleBuy(tx *gorm.DB, uid uint, symbol string, amount, total float64) error {
	// 检查USDT余额
	usdtHolding, err := getOrCreateHolding(tx, uid, "USDT")
	if err != nil {
		return fmt.Errorf("获取USDT持仓失败")
	}

	if usdtHolding.Amount < total {
		return &BusinessError{Message: "USDT余额不足"}
	}

	// 减少USDT
	if err := updateHolding(tx, usdtHolding, -total); err != nil {
		return fmt.Errorf("更新USDT持仓失败")
	}

	// 增加加密资产
	cryptoHolding, err := getOrCreateHolding(tx, uid, symbol)
	if err != nil {
		return fmt.Errorf("创建持仓失败")
	}

	if err := updateHolding(tx, cryptoHolding, amount); err != nil {
		return fmt.Errorf("更新持仓失败")
	}

	// 记录USDT投入
	inv, err := getOrCreateInvestment(tx, uid, symbol)
	if err != nil {
		return fmt.Errorf("创建投资记录失败")
	}
	inv.TotalIn += total
	if err := tx.Save(inv).Error; err != nil {
		return fmt.Errorf("保存投资记录失败")
	}
	return nil
}

// handleSell 处理卖出
func handleSell(tx *gorm.DB, uid uint, symbol string, amount, total float64) error {
	// 检查持仓
	cryptoHolding, err := getOrCreateHolding(tx, uid, symbol)
	if err != nil {
		return fmt.Errorf("获取持仓失败")
	}

	if cryptoHolding.Amount < amount {
		return &BusinessError{Message: "持仓不足"}
	}

	// 减少加密资产
	if err := updateHolding(tx, cryptoHolding, -amount); err != nil {
		return fmt.Errorf("更新持仓失败")
	}

	// 增加USDT
	usdtHolding, err := getOrCreateHolding(tx, uid, "USDT")
	if err != nil {
		return fmt.Errorf("获取USDT持仓失败")
	}

	if err := updateHolding(tx, usdtHolding, total); err != nil {
		return fmt.Errorf("更新USDT持仓失败")
	}

	// 记录投资退出（借贷记账法：按卖出比例回收成本）
	inv, err := getOrCreateInvestment(tx, uid, symbol)
	if err != nil {
		return fmt.Errorf("获取投资记录失败")
	}

	// 计算卖出比例和回收的成本
	// 卖出前持仓 = 当前持仓 + 本次卖出数量
	holdingBeforeSell := cryptoHolding.Amount + amount
	if holdingBeforeSell > 0 {
		sellRatio := amount / holdingBeforeSell
		// 当前累计成本 = TotalIn - TotalOut
		currentCost := inv.TotalIn - inv.TotalOut
		costRecovered := currentCost * sellRatio
		inv.TotalOut += costRecovered
	}

	if err := tx.Save(inv).Error; err != nil {
		return fmt.Errorf("保存投资记录失败")
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
			Symbol:    t.Symbol,
			Type:      t.Type,
			Amount:    t.Amount,
			Price:     t.Price,
			Total:     t.Total,
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

	// 删除交易记录
	if err := tx.Delete(&trade).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除交易记录失败"})
		return
	}

	// 重新计算持仓和投资记录
	if trade.Symbol != "USDT" {
		if err := recalcAsset(tx, uid, trade.Symbol); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

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
// 使用借贷记账法：TotalIn=累计买入投入, TotalOut=累计卖出按比例回收的成本
func recalcAsset(tx *gorm.DB, uid uint, symbol string) error {
	var trades []models.Trade
	if err := tx.Where("user_id = ? AND symbol = ?", uid, symbol).Order("created_at asc").Find(&trades).Error; err != nil {
		return fmt.Errorf("获取交易记录失败")
	}

	var amount, totalIn, totalOut float64
	var currentCost float64 // 当前累计投入成本

	for _, t := range trades {
		switch t.Type {
		case "buy":
			amount += t.Amount
			totalIn += t.Total
			currentCost += t.Total // 买入增加成本
		case "sell":
			if amount > 0 {
				// 按卖出比例计算回收的成本（借贷记账法）
				sellRatio := t.Amount / amount
				costRecovered := currentCost * sellRatio
				totalOut += costRecovered
				currentCost -= costRecovered
			}
			amount -= t.Amount
		}
	}

	// UPSERT Holding: 存在则更新，不存在则插入
	if amount > 0 {
		holding := models.Holding{UserID: uid, Symbol: symbol, Amount: amount}
		if err := tx.Where("user_id = ? AND symbol = ?", uid, symbol).Assign(holding).FirstOrCreate(&holding).Error; err != nil {
			return fmt.Errorf("更新持仓失败")
		}
	} else {
		// 持仓为0，删除记录
		tx.Where("user_id = ? AND symbol = ?", uid, symbol).Delete(&models.Holding{})
	}

	// UPSERT Investment
	if totalIn > 0 || totalOut > 0 {
		inv := models.Investment{
			UserID:   uid,
			Symbol:   symbol,
			TotalIn:  totalIn,
			TotalOut: totalOut,
		}
		if err := tx.Where("user_id = ? AND symbol = ?", uid, symbol).Assign(inv).FirstOrCreate(&inv).Error; err != nil {
			return fmt.Errorf("更新投资记录失败")
		}
	} else {
		// 无投资记录，删除
		tx.Where("user_id = ? AND symbol = ?", uid, symbol).Delete(&models.Investment{})
	}

	return nil
}

// recalcUSDT 重新计算USDT持仓
// 使用UPSERT优化：避免DELETE+INSERT，直接UPDATE或INSERT
func recalcUSDT(tx *gorm.DB, uid uint) error {
	var recharge, buyTotal, sellTotal float64
	tx.Model(&models.Trade{}).Where("user_id = ? AND type = ?", uid, "recharge").Select("COALESCE(SUM(amount), 0)").Scan(&recharge)
	tx.Model(&models.Trade{}).Where("user_id = ? AND type = ?", uid, "buy").Select("COALESCE(SUM(total), 0)").Scan(&buyTotal)
	tx.Model(&models.Trade{}).Where("user_id = ? AND type = ?", uid, "sell").Select("COALESCE(SUM(total), 0)").Scan(&sellTotal)

	balance := recharge - buyTotal + sellTotal
	if balance > 0 {
		// UPSERT: 存在则更新，不存在则插入
		holding := models.Holding{UserID: uid, Symbol: "USDT", Amount: balance}
		return tx.Where("user_id = ? AND symbol = ?", uid, "USDT").Assign(holding).FirstOrCreate(&holding).Error
	}
	// 余额为0，删除记录
	return tx.Where("user_id = ? AND symbol = ?", uid, "USDT").Delete(&models.Holding{}).Error
}

// ClearTrades 清空所有交易记录
// 使用批量DELETE优化，减少往返次数
func ClearTrades(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	db := config.GetDB()
	uid := userID.(uint)

	// 使用原生SQL批量删除，一条SQL删除所有相关表数据
	// 注意：表名使用复数形式，与GORM默认一致
	sql := `
		DELETE t, h, i FROM trades t
		LEFT JOIN holdings h ON h.user_id = t.user_id
		LEFT JOIN investments i ON i.user_id = t.user_id
		WHERE t.user_id = ?
	`
	if err := db.Exec(sql, uid).Error; err != nil {
		// 如果批量删除失败，回退到逐个删除
		tx := db.Begin()
		models := []interface{}{&models.Trade{}, &models.Holding{}, &models.Investment{}}
		for _, model := range models {
			if err := tx.Where("user_id = ?", uid).Delete(model).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "清空数据失败"})
				return
			}
		}
		if err := tx.Commit().Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "提交清空操作失败"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "所有数据已清空"})
}

// ========== 价格接口 ==========

// GetAssetPrice 获取单个资产价格
func GetAssetPrice(c *gin.Context) {
	symbol := c.Param("symbol")

	// 校验symbol参数
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "币种代码不能为空"})
		return
	}

	// 检查是否是USDT或支持的加密货币
	if symbol != "USDT" && !supportedCryptos[symbol] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的加密货币"})
		return
	}

	if symbol == "USDT" {
		c.JSON(http.StatusOK, gin.H{"symbol": symbol, "price": 1.0})
		return
	}

	url := fmt.Sprintf("https://rest.coincap.io/v3/price/bysymbol/%s", symbol)
	req, _ := http.NewRequest("GET", url, nil)
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

// ========== Dashboard 接口 ==========

// PortfolioItem 投资组合项
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

// GetDashboard 获取仪表盘聚合数据
func GetDashboard(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	db := config.GetDB()
	uid := userID.(uint)

	// 获取价格数据
	prices, priceChanges, updatedAt := fetchPrices()

	// 获取持仓和投资记录
	var holdings []models.Holding
	db.Where("user_id = ?", uid).Find(&holdings)

	var investments []models.Investment
	db.Where("user_id = ?", uid).Find(&investments)

	// 计算实现盈亏 = 所有卖出收入 - 所有卖出对应的成本
	var totalRealizedPL float64
	var trades []models.Trade
	db.Where("user_id = ?", uid).Order("created_at asc").Find(&trades)

	// 按币种累计成本和持仓，计算实现盈亏
	assetCost := make(map[string]float64)   // 各币种的当前累计成本
	assetAmount := make(map[string]float64) // 各币种的当前持仓量
	for _, t := range trades {
		switch t.Type {
		case "buy":
			assetCost[t.Symbol] += t.Total
			assetAmount[t.Symbol] += t.Amount
		case "sell":
			if assetAmount[t.Symbol] > 0 && t.Amount > 0 {
				// 按卖出比例计算回收的成本（借贷记账法）
				sellRatio := t.Amount / assetAmount[t.Symbol]
				costRecovered := assetCost[t.Symbol] * sellRatio
				totalRealizedPL += t.Total - costRecovered // 卖出收入 - 成本回收
				assetCost[t.Symbol] -= costRecovered
				assetAmount[t.Symbol] -= t.Amount
			}
		}
	}

	// 计算统计数据
	stats := calculatePortfolioStats(holdings, investments, prices, priceChanges, totalRealizedPL)

	c.JSON(http.StatusOK, gin.H{
		"prices":                      prices,
		"price_changes":               priceChanges,
		"updated_at":                  updatedAt,
		"portfolio":                   stats.portfolio,
		"crypto_value":                stats.totalValue,                                // 加密资产市值（不含USDT）
		"total_assets_value":          stats.totalAssetsValue,                          // 总资产价值（加密资产+USDT）
		"usdt_balance":                stats.usdtBalance,                               // USDT余额
		"unrealized_profit_loss":      stats.totalUnrealizedPL,                         // 浮动盈亏（未实现）
		"unrealized_profit_loss_rate": stats.totalUnrealizedPLRate,                     // 浮动盈亏率
		"realized_profit_loss":        stats.totalRealizedPL,                           // 实现盈亏（已卖出部分的盈亏）
		"realized_profit_loss_rate":   stats.totalRealizedPLRate,                       // 实现盈亏率
		"total_profit_loss":           stats.totalUnrealizedPL + stats.totalRealizedPL, // 总盈亏
		"value_change_24h":            stats.totalValueChange24h,                       // 24小时价值变化率
	})
}

// fetchPrices 获取价格数据
func fetchPrices() (map[string]float64, map[string]float64, int64) {
	ids := []string{"bitcoin", "ethereum", "binance-coin", "xrp", "cardano", "solana", "dogecoin", "tron", "avalanche", "hyperliquid"}
	idsParam := strings.Join(ids, ",")
	url := fmt.Sprintf("https://rest.coincap.io/v3/assets?ids=%s", idsParam)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer b617d9cf029dbb40f02b058a0e74919176b768cf36fd1ea6fae55a13a1610f41")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return map[string]float64{"USDT": 1.0}, map[string]float64{"USDT": 0}, 0
	}
	defer resp.Body.Close()

	prices := map[string]float64{"USDT": 1.0}
	priceChanges := map[string]float64{"USDT": 0}
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
				change24hPercent, _ := strconv.ParseFloat(item.ChangePercent24Hr, 64)
				prices[item.Symbol] = price
				priceChanges[item.Symbol] = change24hPercent / 100
			}
			updatedAt = result.Timestamp
		}
	}

	return prices, priceChanges, updatedAt
}

// PortfolioStats 投资组合统计
type PortfolioStats struct {
	portfolio             []PortfolioItem
	totalValue            float64
	totalAssetsValue      float64
	usdtBalance           float64
	totalUnrealizedPL     float64
	totalUnrealizedPLRate float64
	totalRealizedPL       float64
	totalRealizedPLRate   float64 // 实现盈亏率
	totalValueChange24h   float64
}

// calculatePortfolioStats 计算投资组合统计
// totalRealizedPL: 已实现盈亏（所有卖出收入 - 所有卖出对应的成本）
func calculatePortfolioStats(holdings []models.Holding, investments []models.Investment, prices, priceChanges map[string]float64, totalRealizedPL float64) PortfolioStats {
	invMap := make(map[string]*models.Investment, len(investments))
	for i := range investments {
		invMap[investments[i].Symbol] = &investments[i]
	}

	portfolio := make([]PortfolioItem, 0, len(holdings))
	var totalValue, totalAssetsValue, usdtBalance, totalUnrealizedPL, totalCost, weightedChange float64

	for _, h := range holdings {
		isUSDT := h.Symbol == "USDT"
		price := prices[h.Symbol]
		if isUSDT {
			price = 1
		}

		marketValue := h.Amount * price
		totalAssetsValue += marketValue

		if isUSDT {
			usdtBalance = h.Amount
			portfolio = append(portfolio, PortfolioItem{
				Symbol:       h.Symbol,
				Amount:       h.Amount,
				CurrentPrice: 1,
				AvgCost:      1,
				MarketValue:  marketValue,
				Cost:         h.Amount,
				ProfitLoss:   0,
				PLRate:       0,
			})
			continue
		}

		// 跳过持仓为0的加密资产
		if h.Amount == 0 {
			continue
		}

		totalValue += marketValue

		inv := invMap[h.Symbol]
		avgCost := calcAvgCost(inv, &h)
		cost := inv.TotalIn - inv.TotalOut
		profitLoss := marketValue - cost
		plRate := 0.0
		if cost != 0 {
			plRate = (profitLoss / cost) * 100
		}

		totalUnrealizedPL += profitLoss
		totalCost += cost
		weightedChange += marketValue * priceChanges[h.Symbol]

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

	totalUnrealizedPLRate := 0.0
	if totalCost != 0 {
		totalUnrealizedPLRate = (totalUnrealizedPL / totalCost) * 100
	}

	totalValueChange24h := 0.0
	if totalValue != 0 {
		totalValueChange24h = (weightedChange / totalValue) * 100
	}

	// 计算实现盈亏率 = 实现盈亏 / 历史总投入成本
	// 历史总投入成本 = 各币种 TotalIn 之和
	var totalHistoricalCost float64
	for _, inv := range investments {
		if inv.Symbol != "USDT" {
			totalHistoricalCost += inv.TotalIn
		}
	}
	totalRealizedPLRate := 0.0
	if totalHistoricalCost != 0 {
		totalRealizedPLRate = (totalRealizedPL / totalHistoricalCost) * 100
	}

	// totalRealizedPL 已由调用方计算好（所有卖出收入 - 所有卖出对应的成本）

	return PortfolioStats{
		portfolio:             portfolio,
		totalValue:            totalValue,
		totalAssetsValue:      totalAssetsValue,
		usdtBalance:           usdtBalance,
		totalUnrealizedPL:     totalUnrealizedPL,
		totalUnrealizedPLRate: totalUnrealizedPLRate,
		totalRealizedPL:       totalRealizedPL,
		totalRealizedPLRate:   totalRealizedPLRate,
		totalValueChange24h:   totalValueChange24h,
	}
}

// calcAvgCost 计算平均成本 = USDT净投入 / 持仓量
func calcAvgCost(inv *models.Investment, holding *models.Holding) float64 {
	if holding.Amount == 0 {
		return 0
	}
	netInvestment := inv.TotalIn - inv.TotalOut
	return netInvestment / holding.Amount
}
