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
	"gitee.com/palagend/fx/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	"AAPL": true,
	"MSFT": true,
	"GOOG": true,
	"AMZN": true,
	"TSLA": true,
	"META": true,
	"NVDA": true,
	"BABA": true,
	"ORCL": true,
	"CRCL": true,
	"MSTR": true,
	"QQQI": true,
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

	// 只能删除24小时内的交易
	if time.Since(trade.CreatedAt) > 24*time.Hour {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只能删除24小时内的交易记录"})
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

// recalcAllHoldings 重新计算所有持仓
func recalcAllHoldings(tx *gorm.DB, uid uint) error {
	// 删除该用户的所有持仓
	if err := tx.Where("user_id = ?", uid).Delete(&models.Holding{}).Error; err != nil {
		return err
	}

	// 重新计算持仓
	var trades []models.Trade
	if err := tx.Where("user_id = ?", uid).Order("created_at asc").Find(&trades).Error; err != nil {
		return err
	}

	holdings := make(map[string]*models.Holding) // key: assetType:symbol
	cashHolding := &models.Holding{
		UserID:    uid,
		AssetType: "cash",
		Symbol:    "USD",
		Currency:  "USD",
		Amount:    0,
	}

	for _, t := range trades {
		if t.Type == "recharge" {
			cashHolding.Amount += t.Amount
			continue
		}

		key := t.AssetType + ":" + t.Symbol
		if holdings[key] == nil {
			holdings[key] = &models.Holding{
				UserID:    uid,
				AssetType: t.AssetType,
				Symbol:    t.Symbol,
				Currency:  t.Currency,
				Amount:    0,
			}
		}

		switch t.Type {
		case "buy":
			holdings[key].Amount += t.Amount
			cashHolding.Amount -= t.Total
		case "sell":
			holdings[key].Amount -= t.Amount
			cashHolding.Amount += t.Total
		}
	}

	// 保存持仓
	for _, h := range holdings {
		if h.Amount != 0 {
			if err := tx.Create(h).Error; err != nil {
				return err
			}
		}
	}

	// 保存现金持仓
	if cashHolding.Amount != 0 {
		if err := tx.Create(cashHolding).Error; err != nil {
			return err
		}
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
	assetType := c.DefaultQuery("asset_type", "crypto")

	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "币种代码不能为空"})
		return
	}

	switch assetType {
	case "crypto":
		getCryptoPrice(c, symbol)
	case "us_stock":
		getUSStockPrice(c, symbol)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的资产类型"})
	}
}

// getCryptoPrice 获取加密货币价格
func getCryptoPrice(c *gin.Context, symbol string) {

	if !supportedCryptos[symbol] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的加密货币"})
		return
	}

	url := fmt.Sprintf("https://rest.coincap.io/v3/price/bysymbol/%s", symbol)
	req, _ := http.NewRequest("GET", url, nil)

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
		"asset_type": "crypto",
		"currency":   "USD",
		"updated_at": result.Timestamp,
	})
}

// getUSStockPrice 获取美股价格
func getUSStockPrice(c *gin.Context, symbol string) {
	if !supportedUSStocks[symbol] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的美股代码"})
		return
	}

	client := utils.GetStockClient()
	price, err := client.GetUSStockPrice(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"symbol":      symbol,
		"price":       price.Close,
		"open":        price.Open,
		"high":        price.High,
		"low":         price.Low,
		"pre_close":   price.PreClose,
		"change":      price.Change,
		"change_rate": price.PctChg,
		"asset_type":  "us_stock",
		"currency":    "USD",
		"trade_date":  price.TradeDate,
	})
}

// ========== Dashboard 接口 ==========

// PortfolioItem 投资组合项
type PortfolioItem struct {
	AssetType      string  `json:"asset_type"`
	Symbol         string  `json:"symbol"`
	Amount         float64 `json:"amount"`
	CurrentPrice   float64 `json:"current_price"`
	AvgCost        float64 `json:"avg_cost"`
	MarketValue    float64 `json:"market_value"`
	Cost           float64 `json:"cost"`
	ProfitLoss     float64 `json:"profit_loss"`
	PLRate         float64 `json:"pl_rate"`
	RealizedPL     float64 `json:"realized_pl"`
	RealizedPLRate float64 `json:"realized_pl_rate"`
	Currency       string  `json:"currency"`
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
	cryptoPrices, cryptoChanges, cryptoUpdatedAt := fetchCryptoPrices()
	usStockPrices := fetchUSStockPrices()

	// 获取持仓和交易记录
	var holdings []models.Holding
	db.Where("user_id = ?", uid).Find(&holdings)

	var trades []models.Trade
	db.Where("user_id = ?", uid).Order("created_at asc").Find(&trades)

	// 计算统计数据
	stats := calculatePortfolioStats(holdings, cryptoPrices, cryptoChanges, usStockPrices, trades)

	c.JSON(http.StatusOK, gin.H{
		"prices":                      cryptoPrices,
		"us_stock_prices":             usStockPrices,
		"price_changes":               cryptoChanges,
		"crypto_updated_at":           cryptoUpdatedAt,
		"btc_price":                   cryptoPrices["BTC"], // 用于BTC本位计算
		"portfolio":                   stats.portfolio,
		"crypto_value":                stats.cryptoValue,
		"us_stock_value":              stats.usStockValue,
		"cash_balance":                stats.cashBalance,
		"total_assets_value":          stats.totalAssetsValue,
		"unrealized_profit_loss":      stats.totalUnrealizedPL,
		"unrealized_profit_loss_rate": stats.totalUnrealizedPLRate,
		"realized_profit_loss":        stats.totalRealizedPL,
		"realized_profit_loss_rate":   stats.totalRealizedPLRate,
		"total_profit_loss":           stats.totalUnrealizedPL + stats.totalRealizedPL,
		"value_change_24h":            stats.totalValueChange24h,
	})
}

// fetchCryptoPrices 获取加密货币价格
func fetchCryptoPrices() (map[string]float64, map[string]float64, int64) {
	ids := []string{"bitcoin", "ethereum", "binance-coin", "xrp", "cardano", "solana", "dogecoin", "tron", "avalanche", "hyperliquid", "tether"}
	idsParam := strings.Join(ids, ",")
	url := fmt.Sprintf("https://rest.coincap.io/v3/assets?ids=%s", idsParam)

	req, _ := http.NewRequest("GET", url, nil)

	cfg := config.GetConfig()
	if cfg.API.CoinCapKey != "" {
		req.Header.Add("Authorization", "Bearer "+cfg.API.CoinCapKey)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return map[string]float64{}, map[string]float64{}, 0
	}
	defer resp.Body.Close()

	prices := map[string]float64{}
	priceChanges := map[string]float64{}
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

// fetchUSStockPrices 获取美股价格
func fetchUSStockPrices() map[string]float64 {
	prices := make(map[string]float64)
	client := utils.GetStockClient()

	// 构建股票代码列表
	symbols := make([]string, 0, len(supportedUSStocks))
	for symbol := range supportedUSStocks {
		symbols = append(symbols, symbol)
	}

	// 使用批量获取方法（带缓存）
	results := client.GetUSStockPricesBatch(symbols)
	for symbol, price := range results {
		if price != nil {
			prices[symbol] = price.Close
		}
	}

	return prices
}

// PortfolioStats 投资组合统计
type PortfolioStats struct {
	portfolio             []PortfolioItem
	cryptoValue           float64
	usStockValue          float64
	cashBalance           float64
	totalAssetsValue      float64
	totalUnrealizedPL     float64
	totalUnrealizedPLRate float64
	totalRealizedPL       float64
	totalRealizedPLRate   float64
	totalValueChange24h   float64
}

// calculatePortfolioStats 计算投资组合统计
func calculatePortfolioStats(holdings []models.Holding, cryptoPrices, cryptoChanges map[string]float64,
	usStockPrices map[string]float64, trades []models.Trade) PortfolioStats {

	portfolio := make([]PortfolioItem, 0, len(holdings))
	var cryptoValue, usStockValue, cashBalance, totalUnrealizedPL float64
	var totalRealizedPL, totalHistoricalCost float64

	// 按资产类型和代码分组计算
	assetData := make(map[string]map[string]struct {
		amount     float64
		cost       float64
		totalIn    float64
		realizedPL float64
	})

	for _, t := range trades {
		if t.Type == "recharge" {
			continue
		}

		if assetData[t.AssetType] == nil {
			assetData[t.AssetType] = make(map[string]struct {
				amount     float64
				cost       float64
				totalIn    float64
				realizedPL float64
			})
		}

		d := assetData[t.AssetType][t.Symbol]

		switch t.Type {
		case "buy":
			d.amount += t.Amount
			d.cost += t.Total
			d.totalIn += t.Total
		case "sell":
			if d.amount > 0 && t.Amount > 0 {
				sellRatio := t.Amount / d.amount
				costRecovered := d.cost * sellRatio
				realizedPL := t.Total - costRecovered

				d.realizedPL += realizedPL
				d.cost -= costRecovered
				d.amount -= t.Amount
			}
		}

		assetData[t.AssetType][t.Symbol] = d
	}

	// 计算实现盈亏
	for _, symbols := range assetData {
		for symbol, d := range symbols {
			if symbol != "USD" {
				totalRealizedPL += d.realizedPL
				totalHistoricalCost += d.totalIn
			}
		}
	}

	for _, h := range holdings {
		// 现金持仓
		if h.AssetType == "cash" && h.Symbol == "USD" {
			cashBalance = h.Amount
			continue
		}

		// 获取价格
		var price float64
		switch h.AssetType {
		case "crypto":
			price = cryptoPrices[h.Symbol]
		case "us_stock":
			price = usStockPrices[h.Symbol]
		}

		marketValue := h.Amount * price

		// 从预计算数据获取成本
		d := assetData[h.AssetType][h.Symbol]
		cost := d.cost
		realizedPL := d.realizedPL

		// 累加到对应资产类型
		switch h.AssetType {
		case "crypto":
			cryptoValue += marketValue
		case "us_stock":
			usStockValue += marketValue
		}

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

		// 计算实现盈亏率
		realizedPLRate := 0.0
		if d.totalIn != 0 {
			realizedPLRate = (realizedPL / d.totalIn) * 100
		}

		portfolio = append(portfolio, PortfolioItem{
			AssetType:      h.AssetType,
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
			Currency:       h.Currency,
		})
	}

	totalAssetsValue := cryptoValue + usStockValue + cashBalance

	// 计算总盈亏率
	totalUnrealizedPLRate := 0.0
	totalCost := 0.0
	for _, h := range holdings {
		if h.AssetType != "cash" {
			d := assetData[h.AssetType][h.Symbol]
			totalCost += d.cost
		}
	}
	if totalCost != 0 {
		totalUnrealizedPLRate = (totalUnrealizedPL / totalCost) * 100
	}

	// 计算总实现盈亏率
	totalRealizedPLRate := 0.0
	if totalHistoricalCost != 0 {
		totalRealizedPLRate = (totalRealizedPL / totalHistoricalCost) * 100
	}

	return PortfolioStats{
		portfolio:             portfolio,
		cryptoValue:           cryptoValue,
		usStockValue:          usStockValue,
		cashBalance:           cashBalance,
		totalAssetsValue:      totalAssetsValue,
		totalUnrealizedPL:     totalUnrealizedPL,
		totalUnrealizedPLRate: totalUnrealizedPLRate,
		totalRealizedPL:       totalRealizedPL,
		totalRealizedPLRate:   totalRealizedPLRate,
		totalValueChange24h:   0,
	}
}

// ========== 导入/导出接口 ==========

// ExportData 导出用户数据
type ExportData struct {
	Version  string        `json:"version"`
	Exported string        `json:"exported"`
	Trades   []TradeExport `json:"trades"`
}

type TradeExport struct {
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

// ExportDataHandler 导出数据接口
func ExportDataHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	db := config.GetDB()
	uid := userID.(uint)

	var trades []models.Trade
	if err := db.Where("user_id = ?", uid).Order("created_at asc").Find(&trades).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取交易记录失败"})
		return
	}

	tradeExports := make([]TradeExport, len(trades))
	for i, t := range trades {
		tradeExports[i] = TradeExport{
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

	exportData := ExportData{
		Version:  "1.0",
		Exported: time.Now().Format("2006-01-02 15:04:05"),
		Trades:   tradeExports,
	}

	c.JSON(http.StatusOK, gin.H{"data": exportData})
}

// ImportPreviewRequest 导入预览请求
type ImportPreviewRequest struct {
	Data ExportData `json:"data"`
}

// ImportPreviewResponse 导入预览响应
type ImportPreviewResponse struct {
	TotalTrades   int            `json:"total_trades"`
	NewTrades     int            `json:"new_trades"`
	Conflicts     int            `json:"conflicts"`
	ConflictItems []ConflictItem `json:"conflict_items"`
}

type ConflictItem struct {
	Trade  TradeExport `json:"trade"`
	Reason string      `json:"reason"`
}

// ImportPreviewHandler 导入预览接口
func ImportPreviewHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req ImportPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	db := config.GetDB()
	uid := userID.(uint)

	// 获取用户现有交易记录的UUID集合（用于检测冲突）
	var existingTrades []models.Trade
	db.Where("user_id = ?", uid).Find(&existingTrades)

	existingUUIDs := make(map[string]bool)
	for _, t := range existingTrades {
		if t.UUID != "" {
			existingUUIDs[t.UUID] = true
		}
	}

	preview := ImportPreviewResponse{
		TotalTrades:   len(req.Data.Trades),
		NewTrades:     0,
		Conflicts:     0,
		ConflictItems: []ConflictItem{},
	}

	for _, trade := range req.Data.Trades {
		// 使用UUID检测冲突，如果没有UUID则使用时间戳作为备用
		var key string
		if trade.UUID != "" {
			key = trade.UUID
		} else {
			key = fmt.Sprintf("%s_%s_%s_%s", trade.AssetType, trade.Symbol, trade.Type, trade.CreatedAt)
		}

		if existingUUIDs[key] {
			preview.Conflicts++
			preview.ConflictItems = append(preview.ConflictItems, ConflictItem{
				Trade:  trade,
				Reason: "与现有记录UUID相同",
			})
		} else {
			preview.NewTrades++
		}
	}

	c.JSON(http.StatusOK, gin.H{"preview": preview})
}

// ImportConfirmRequest 导入确认请求
type ImportConfirmRequest struct {
	Data             ExportData `json:"data"`
	ConflictStrategy string     `json:"conflict_strategy"` // skip 或 overwrite
}

// ImportConfirmHandler 导入确认接口
func ImportConfirmHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req ImportConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	if req.ConflictStrategy != "skip" && req.ConflictStrategy != "overwrite" {
		req.ConflictStrategy = "skip"
	}

	db := config.GetDB()
	uid := userID.(uint)

	// 获取用户现有交易记录的UUID集合
	var existingTrades []models.Trade
	db.Where("user_id = ?", uid).Find(&existingTrades)

	existingUUIDs := make(map[string]uint) // UUID -> trade ID
	for _, t := range existingTrades {
		if t.UUID != "" {
			existingUUIDs[t.UUID] = t.ID
		}
	}

	var imported, skipped, overwritten int

	tx := db.Begin()

	for _, trade := range req.Data.Trades {
		// 使用UUID检测冲突，如果没有UUID则使用时间戳作为备用
		var key string
		if trade.UUID != "" {
			key = trade.UUID
		} else {
			key = fmt.Sprintf("%s_%s_%s_%s", trade.AssetType, trade.Symbol, trade.Type, trade.CreatedAt)
		}

		if existingID, exists := existingUUIDs[key]; exists {
			// 存在冲突
			if req.ConflictStrategy == "overwrite" {
				// 删除旧记录
				if err := tx.Unscoped().Delete(&models.Trade{}, existingID).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "删除旧记录失败"})
					return
				}
				overwritten++
			} else {
				// 跳过
				skipped++
				continue
			}
		}

		// 解析时间
		createdAt, _ := time.Parse("2006-01-02 15:04:05", trade.CreatedAt)
		if createdAt.IsZero() {
			createdAt = time.Now()
		}

		// 创建新记录，保留原始UUID（如果有），否则生成新的UUID
		newTrade := models.Trade{
			UUID:      trade.UUID,
			UserID:    uid,
			AssetType: trade.AssetType,
			Symbol:    trade.Symbol,
			Type:      trade.Type,
			Amount:    trade.Amount,
			Price:     trade.Price,
			Total:     trade.Total,
			Currency:  trade.Currency,
		}
		// 如果导入的数据没有UUID，生成一个新的
		if newTrade.UUID == "" {
			newTrade.UUID = uuid.New().String()
		}
		newTrade.CreatedAt = createdAt

		if err := tx.Create(&newTrade).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建交易记录失败"})
			return
		}

		imported++
	}

	// 重新计算持仓
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
		"imported":    imported,
		"skipped":     skipped,
		"overwritten": overwritten,
	})
}
