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

// DeleteTrade 删除交易记录（带安全校验）
// 限制：只能删除最近24小时内的交易，且删除后不能导致负持仓
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

	// 校验1：只能删除24小时内的交易
	if time.Since(trade.CreatedAt) > 24*time.Hour {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只能删除24小时内的交易记录"})
		return
	}

	// 校验2：模拟删除后的持仓状态
	tx := db.Begin()

	// 获取该用户所有交易（排除要删除的这条）
	var remainingTrades []models.Trade
	if err := tx.Where("user_id = ? AND id != ?", uid, tradeID).Order("created_at asc").Find(&remainingTrades).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取交易记录失败"})
		return
	}

	// 计算删除后的持仓状态
	simulatedHoldings := make(map[string]float64)
	usdtBalance := 0.0

	for _, t := range remainingTrades {
		switch t.Type {
		case "recharge":
			usdtBalance += t.Amount
		case "buy":
			simulatedHoldings[t.Symbol] += t.Amount
			usdtBalance -= t.Total
		case "sell":
			simulatedHoldings[t.Symbol] -= t.Amount
			usdtBalance += t.Total
		}
	}

	// 校验3：删除后不能导致任何资产负持仓
	for symbol, amount := range simulatedHoldings {
		if amount < 0 {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("删除该交易会导致 %s 持仓为负数(%.8f)，无法删除", symbol, amount),
			})
			return
		}
	}

	// 校验4：删除后不能导致 USDT 余额为负
	if usdtBalance < 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("删除该交易会导致 USDT 余额为负数(%.2f)，无法删除", usdtBalance),
		})
		return
	}

	// 执行物理删除（永久删除，不使用软删除）
	if err := tx.Unscoped().Delete(&trade).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除交易记录失败"})
		return
	}

	// 重新计算所有相关持仓
	affectedSymbols := make(map[string]bool)
	affectedSymbols["USDT"] = true
	if trade.Symbol != "USDT" {
		affectedSymbols[trade.Symbol] = true
	}

	// 对于卖出交易，还需要检查是否影响了其他资产的买入成本计算
	// 这里简化处理：重新计算所有持仓
	for symbol := range simulatedHoldings {
		if symbol != "USDT" {
			if err := recalcAsset(tx, uid, symbol); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("重新计算 %s 持仓失败: %v", symbol, err)})
				return
			}
		}
	}

	// 重新计算 USDT 持仓
	if err := recalcUSDT(tx, uid); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重新计算 USDT 持仓失败"})
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
			"symbol":     trade.Symbol,
			"type":       trade.Type,
			"amount":     trade.Amount,
			"price":      trade.Price,
			"created_at": trade.CreatedAt,
		},
	})
}

// recalcAsset 重新计算资产持仓
// 从交易记录计算当前持仓量
func recalcAsset(tx *gorm.DB, uid uint, symbol string) error {
	var trades []models.Trade
	if err := tx.Where("user_id = ? AND symbol = ?", uid, symbol).Order("created_at asc").Find(&trades).Error; err != nil {
		return fmt.Errorf("获取交易记录失败")
	}

	var amount float64
	for _, t := range trades {
		switch t.Type {
		case "buy":
			amount += t.Amount
		case "sell":
			amount -= t.Amount
		}
	}

	// UPSERT Holding: 存在则更新，不存在则插入（包括持仓为0的情况）
	holding := models.Holding{UserID: uid, Symbol: symbol, Amount: amount}
	if err := tx.Where("user_id = ? AND symbol = ?", uid, symbol).Assign(holding).FirstOrCreate(&holding).Error; err != nil {
		return fmt.Errorf("更新持仓失败")
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
	// UPSERT: 存在则更新，不存在则插入（包括余额为0的情况）
	holding := models.Holding{UserID: uid, Symbol: "USDT", Amount: balance}
	return tx.Where("user_id = ? AND symbol = ?", uid, "USDT").Assign(holding).FirstOrCreate(&holding).Error
}

// ClearTrades 清空所有交易记录（物理删除）
func ClearTrades(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	db := config.GetDB()
	uid := userID.(uint)

	// 使用事务批量物理删除
	tx := db.Begin()
	if err := tx.Unscoped().Where("user_id = ?", uid).Delete(&models.Trade{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清空交易记录失败"})
		return
	}
	if err := tx.Unscoped().Where("user_id = ?", uid).Delete(&models.Holding{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "清空持仓记录失败"})
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

	// 从配置获取 API Key
	cfg := config.GetConfig()
	if cfg.API.CoinCapKey != "" {
		req.Header.Add("Authorization", "Bearer "+cfg.API.CoinCapKey)
	}
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
	Symbol         string  `json:"symbol"`
	Amount         float64 `json:"amount"`
	CurrentPrice   float64 `json:"current_price"`
	AvgCost        float64 `json:"avg_cost"`
	MarketValue    float64 `json:"market_value"`
	Cost           float64 `json:"cost"`
	ProfitLoss     float64 `json:"profit_loss"`
	PLRate         float64 `json:"pl_rate"`
	RealizedPL     float64 `json:"realized_pl"`      // 实现盈亏（该币种已卖出部分的盈亏）
	RealizedPLRate float64 `json:"realized_pl_rate"` // 实现盈亏率
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

	// 获取持仓和交易记录
	var holdings []models.Holding
	db.Where("user_id = ?", uid).Find(&holdings)

	var trades []models.Trade
	db.Where("user_id = ?", uid).Order("created_at asc").Find(&trades)

	// 计算统计数据（从 Trade 实时计算成本和实现盈亏）
	stats := calculatePortfolioStats(holdings, prices, priceChanges, trades)

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

	// 从配置获取 API Key
	cfg := config.GetConfig()
	if cfg.API.CoinCapKey != "" {
		req.Header.Add("Authorization", "Bearer "+cfg.API.CoinCapKey)
	}

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
// 从 Trade 记录实时计算成本和实现盈亏
func calculatePortfolioStats(holdings []models.Holding, prices, priceChanges map[string]float64, trades []models.Trade) PortfolioStats {
	portfolio := make([]PortfolioItem, 0, len(holdings))
	var totalValue, totalAssetsValue, usdtBalance, totalUnrealizedPL, totalCost, weightedChange float64
	var totalRealizedPL, totalHistoricalCost float64

	// 按时间顺序遍历交易，计算各币种的成本和实现盈亏
	assetData := make(map[string]struct {
		amount     float64 // 当前持仓量
		cost       float64 // 当前成本
		totalIn    float64 // 累计买入金额
		realizedPL float64 // 实现盈亏
	})

	for _, t := range trades {
		if t.Symbol == "USDT" {
			continue
		}

		d := assetData[t.Symbol]

		switch t.Type {
		case "buy":
			// 增加持仓和成本
			d.amount += t.Amount
			d.cost += t.Total
			d.totalIn += t.Total

		case "sell":
			// 按比例回收成本，计算实现盈亏
			if d.amount > 0 && t.Amount > 0 {
				sellRatio := t.Amount / d.amount
				costRecovered := d.cost * sellRatio
				realizedPL := t.Total - costRecovered

				d.realizedPL += realizedPL
				d.cost -= costRecovered
				d.amount -= t.Amount
			}
		}

		assetData[t.Symbol] = d
	}

	// 先计算所有有交易记录的币种的总实现盈亏（包括已清仓的）
	for symbol, d := range assetData {
		if symbol != "USDT" {
			totalRealizedPL += d.realizedPL
			totalHistoricalCost += d.totalIn
		}
	}

	for _, h := range holdings {
		if h.Symbol == "USDT" {
			usdtBalance = h.Amount
			portfolio = append(portfolio, PortfolioItem{
				Symbol:       h.Symbol,
				Amount:       h.Amount,
				CurrentPrice: 1.00,
				AvgCost:      0,
				MarketValue:  usdtBalance,
				Cost:         0,
				ProfitLoss:   0,
				PLRate:       0,
			})
			continue
		}
		price := prices[h.Symbol]
		marketValue := h.Amount * price
		totalAssetsValue += marketValue
		// 从预计算数据获取成本和实现盈亏
		d := assetData[h.Symbol]
		cost := d.cost
		realizedPL := d.realizedPL

		// 持仓为0的资产只展示实现盈亏，不参与总市值计算
		if h.Amount == 0 && realizedPL != 0 {
			portfolio = append(portfolio, PortfolioItem{
				Symbol:         h.Symbol,
				Amount:         0,
				CurrentPrice:   price,
				AvgCost:        0,
				MarketValue:    0,
				Cost:           0,
				ProfitLoss:     0,
				PLRate:         0,
				RealizedPL:     realizedPL,
				RealizedPLRate: 0,
			})
			continue
		}

		totalValue += marketValue

		avgCost := 0.0
		if h.Amount != 0 {
			avgCost = cost / h.Amount
		}

		profitLoss := marketValue - cost
		plRate := 0.0
		if cost != 0 {
			plRate = (profitLoss / cost) * 100
		}

		totalUnrealizedPL += profitLoss
		totalCost += cost
		weightedChange += marketValue * priceChanges[h.Symbol]

		// 计算实现盈亏率
		realizedPLRate := 0.0
		if d.totalIn != 0 {
			realizedPLRate = (realizedPL / d.totalIn) * 100
		}

		portfolio = append(portfolio, PortfolioItem{
			Symbol:         h.Symbol,
			Amount:         h.Amount,
			CurrentPrice:   price,
			AvgCost:        avgCost,
			MarketValue:    marketValue,
			Cost:           cost,
			ProfitLoss:     profitLoss,
			PLRate:         plRate,
			RealizedPL:     realizedPL,
			RealizedPLRate: realizedPLRate,
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
	totalRealizedPLRate := 0.0
	if totalHistoricalCost != 0 {
		totalRealizedPLRate = (totalRealizedPL / totalHistoricalCost) * 100
	}

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

// ========== 数据导出/导入接口 ==========

// ExportTrade 导出交易记录结构
type ExportTrade struct {
	Symbol    string  `json:"symbol"`
	Type      string  `json:"type"`
	Amount    float64 `json:"amount"`
	Price     float64 `json:"price"`
	Total     float64 `json:"total"`
	CreatedAt string  `json:"created_at"`
	Notes     string  `json:"notes,omitempty"`
}

// ExportData 导出数据结构
type ExportData struct {
	Version    string        `json:"version"`
	ExportTime string        `json:"export_time"`
	AppName    string        `json:"app_name"`
	Trades     []ExportTrade `json:"trades"`
}

// ImportPreview 导入预览结果
type ImportPreview struct {
	TotalTrades   int            `json:"total_trades"`
	NewTrades     int            `json:"new_trades"`
	Conflicts     int            `json:"conflicts"`
	ConflictItems []ConflictItem `json:"conflict_items,omitempty"`
}

// ConflictItem 冲突项
type ConflictItem struct {
	Trade      ExportTrade `json:"trade"`
	Reason     string      `json:"reason"`
	ExistingID uint        `json:"existing_id,omitempty"`
}

// ExportDataHandler 导出用户数据
func ExportDataHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	db := config.GetDB()
	uid := userID.(uint)

	// 获取所有交易记录
	var trades []models.Trade
	if err := db.Where("user_id = ?", uid).Order("created_at asc").Find(&trades).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取交易记录失败"})
		return
	}

	// 转换为导出格式
	exportTrades := make([]ExportTrade, len(trades))
	for i, t := range trades {
		exportTrades[i] = ExportTrade{
			Symbol:    t.Symbol,
			Type:      t.Type,
			Amount:    t.Amount,
			Price:     t.Price,
			Total:     t.Total,
			CreatedAt: t.CreatedAt.Format(time.RFC3339),
		}
	}

	data := ExportData{
		Version:    "1.0",
		ExportTime: time.Now().UTC().Format(time.RFC3339),
		AppName:    "fx-portfolio",
		Trades:     exportTrades,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// ImportPreviewHandler 导入预览
func ImportPreviewHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req struct {
		Data ExportData `json:"data" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的导入数据"})
		return
	}

	// 验证版本
	if req.Data.Version != "1.0" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的版本: " + req.Data.Version})
		return
	}

	db := config.GetDB()
	uid := userID.(uint)

	// 获取现有交易记录用于冲突检测
	var existingTrades []models.Trade
	db.Where("user_id = ?", uid).Find(&existingTrades)

	// 构建现有交易的时间戳集合（用于快速查找）
	existingMap := make(map[string]bool)
	for _, t := range existingTrades {
		key := fmt.Sprintf("%s_%s_%s", t.Symbol, t.Type, t.CreatedAt.Format(time.RFC3339))
		existingMap[key] = true
	}

	preview := ImportPreview{
		TotalTrades: len(req.Data.Trades),
	}

	// 检查冲突
	for _, trade := range req.Data.Trades {
		// 验证交易数据
		if err := validateImportTrade(&trade); err != nil {
			preview.Conflicts++
			preview.ConflictItems = append(preview.ConflictItems, ConflictItem{
				Trade:  trade,
				Reason: err.Error(),
			})
			continue
		}

		// 检查是否已存在
		key := fmt.Sprintf("%s_%s_%s", trade.Symbol, trade.Type, trade.CreatedAt)
		if existingMap[key] {
			preview.Conflicts++
			preview.ConflictItems = append(preview.ConflictItems, ConflictItem{
				Trade:  trade,
				Reason: "交易记录已存在",
			})
		} else {
			preview.NewTrades++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"preview": preview,
	})
}

// ImportConfirmHandler 确认导入
func ImportConfirmHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req struct {
		Data             ExportData `json:"data" binding:"required"`
		ConflictStrategy string     `json:"conflict_strategy" binding:"required,oneof=skip overwrite"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 验证版本
	if req.Data.Version != "1.0" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的版本: " + req.Data.Version})
		return
	}

	db := config.GetDB()
	uid := userID.(uint)

	// 获取现有交易记录
	var existingTrades []models.Trade
	db.Where("user_id = ?", uid).Find(&existingTrades)
	existingMap := make(map[string]uint) // key -> trade ID
	for _, t := range existingTrades {
		key := fmt.Sprintf("%s_%s_%s", t.Symbol, t.Type, t.CreatedAt.Format(time.RFC3339))
		existingMap[key] = t.ID
	}

	tx := db.Begin()
	imported := 0
	skipped := 0
	overwritten := 0

	for _, trade := range req.Data.Trades {
		// 验证交易数据
		if err := validateImportTrade(&trade); err != nil {
			skipped++
			continue
		}

		key := fmt.Sprintf("%s_%s_%s", trade.Symbol, trade.Type, trade.CreatedAt)
		existingID, exists := existingMap[key]

		if exists {
			// 处理冲突
			switch req.ConflictStrategy {
			case "skip":
				skipped++
				continue
			case "overwrite":
				// 删除旧记录，插入新记录
				if err := tx.Delete(&models.Trade{}, existingID).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "删除旧记录失败"})
					return
				}
				overwritten++
			}
		}

		// 解析时间
		createdAt, _ := time.Parse(time.RFC3339, trade.CreatedAt)
		if createdAt.IsZero() {
			createdAt = time.Now()
		}

		// 创建新交易记录
		newTrade := models.Trade{
			UserID:    uid,
			Symbol:    trade.Symbol,
			Type:      trade.Type,
			Amount:    trade.Amount,
			Price:     trade.Price,
			Total:     trade.Total,
			CreatedAt: createdAt,
		}

		if err := tx.Create(&newTrade).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建交易记录失败"})
			return
		}
		imported++
	}

	// 重新计算所有持仓
	if err := recalcAllHoldings(tx, uid); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重新计算持仓失败"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交导入失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"imported":    imported,
		"skipped":     skipped,
		"overwritten": overwritten,
	})
}

// validateImportTrade 验证导入的交易数据
func validateImportTrade(trade *ExportTrade) error {
	// 必填字段
	if trade.Symbol == "" {
		return fmt.Errorf("币种代码不能为空")
	}
	if trade.Type == "" {
		return fmt.Errorf("交易类型不能为空")
	}

	// 类型有效性
	if trade.Type != "buy" && trade.Type != "sell" && trade.Type != "recharge" {
		return fmt.Errorf("无效的交易类型: %s", trade.Type)
	}

	// 币种有效性
	if trade.Symbol != "USDT" && !supportedCryptos[trade.Symbol] {
		return fmt.Errorf("不支持的币种: %s", trade.Symbol)
	}

	// 数值有效性
	if trade.Amount <= 0 {
		return fmt.Errorf("交易数量必须大于0")
	}
	if trade.Price <= 0 {
		return fmt.Errorf("交易价格必须大于0")
	}

	// 一致性检查（允许0.01的误差）
	expectedTotal := trade.Amount * trade.Price
	if abs(expectedTotal-trade.Total) > 0.01 {
		return fmt.Errorf("交易金额计算不一致: %.2f != %.2f", expectedTotal, trade.Total)
	}

	// 类型特定校验
	if trade.Type == "recharge" && trade.Symbol != "USDT" {
		return fmt.Errorf("充值必须是USDT")
	}
	if trade.Type == "recharge" && trade.Price != 1 {
		return fmt.Errorf("USDT充值价格必须为1")
	}
	if (trade.Type == "buy" || trade.Type == "sell") && trade.Symbol == "USDT" {
		return fmt.Errorf("不能直接买卖USDT")
	}

	return nil
}

// abs 绝对值
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// recalcAllHoldings 重新计算所有持仓
func recalcAllHoldings(tx *gorm.DB, uid uint) error {
	// 删除现有持仓
	if err := tx.Where("user_id = ?", uid).Delete(&models.Holding{}).Error; err != nil {
		return err
	}

	// 获取所有交易记录
	var trades []models.Trade
	if err := tx.Where("user_id = ?", uid).Order("created_at asc").Find(&trades).Error; err != nil {
		return err
	}

	// 计算各币种持仓
	holdings := make(map[string]float64)
	for _, t := range trades {
		switch t.Type {
		case "buy":
			holdings[t.Symbol] += t.Amount
			holdings["USDT"] -= t.Total
		case "sell":
			holdings[t.Symbol] -= t.Amount
			holdings["USDT"] += t.Total
		case "recharge":
			holdings["USDT"] += t.Amount
		}
	}

	// 保存持仓
	for symbol, amount := range holdings {
		holding := models.Holding{
			UserID: uid,
			Symbol: symbol,
			Amount: amount,
		}
		if err := tx.Create(&holding).Error; err != nil {
			return err
		}
	}

	return nil
}
