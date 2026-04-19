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

// calculateAssetCost 从交易记录计算资产的累计成本和持仓
// 返回：当前成本、当前持仓、总投入、总退出
func calculateAssetCost(trades []models.Trade, symbol string) (cost, amount, totalIn, totalOut float64) {
	for _, t := range trades {
		if t.Symbol != symbol {
			continue
		}
		switch t.Type {
		case "buy":
			cost += t.Total
			amount += t.Amount
			totalIn += t.Total
		case "sell":
			if amount > 0 && t.Amount > 0 {
				// 按卖出比例计算回收的成本
				sellRatio := t.Amount / amount
				costRecovered := cost * sellRatio
				cost -= costRecovered
				amount -= t.Amount
				totalOut += costRecovered
			}
		}
	}
	return
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
		DELETE t, h FROM trades t
		LEFT JOIN holdings h ON h.user_id = t.user_id
		WHERE t.user_id = ?
	`
	if err := db.Exec(sql, uid).Error; err != nil {
		// 如果批量删除失败，回退到逐个物理删除
		tx := db.Begin()
		models := []interface{}{&models.Trade{}, &models.Holding{}}
		for _, model := range models {
			if err := tx.Unscoped().Where("user_id = ?", uid).Delete(model).Error; err != nil {
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
