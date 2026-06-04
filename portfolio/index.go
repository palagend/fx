package portfolio

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"api/db"
	"api/middleware"
	"api/models"
	"api/utils"

	"github.com/google/uuid"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"gorm.io/gorm"
)

// 支持的加密货币
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
	"AAPL":  true,
	"MSFT":  true,
	"GOOG":  true,
	"AMZN":  true,
	"TSLA":  true,
	"META":  true,
	"NVDA":  true,
	"BABA":  true,
	"ORCL":  true,
	"CRCL":  true,
	"MSTR":  true,
	"QQQI":  true,
	"TCEHY": true,
	"PURR":  true,
	"QQQ":   true,
}

// BusinessError 业务错误类型
type BusinessError struct {
	Message string
}

func (e *BusinessError) Error() string {
	return e.Message
}

// validateTradeRequest 对交易请求进行业务校验
func validateTradeRequest(req *CreateTradeRequest) error {
	switch req.Type {
	case "recharge":
		if req.AssetType != "cash" {
			return &BusinessError{"充值资产类型必须是cash"}
		}
		if req.Symbol != "USD" {
			return &BusinessError{"充值只支持USD"}
		}
		if req.Price != 1 {
			return &BusinessError{"充值价格必须为1"}
		}
	case "buy", "sell":
		switch req.AssetType {
		case "crypto":
			if !supportedCryptos[req.Symbol] {
				return &BusinessError{fmt.Sprintf("不支持的加密货币: %s", req.Symbol)}
			}
		case "us_stock":
			if !supportedUSStocks[req.Symbol] {
				return &BusinessError{fmt.Sprintf("不支持的美股: %s", req.Symbol)}
			}
		default:
			return &BusinessError{fmt.Sprintf("买卖交易不支持资产类型: %s", req.AssetType)}
		}
	}
	return nil
}

// --- 美股价格缓存 (轻量级，替代被删除的 utils/stock.go) ---

type stockPriceCache struct {
	price      float64
	updateTime time.Time
}

var (
	stockCache    = make(map[string]*stockPriceCache)
	stockCacheMux sync.RWMutex
	tencentAPIURL = "https://qt.gtimg.cn"
	cacheTTL      = 5 * time.Minute
)

func getCachedPrice(symbol string) (float64, bool) {
	stockCacheMux.RLock()
	defer stockCacheMux.RUnlock()
	if c, ok := stockCache[symbol]; ok && time.Since(c.updateTime) < cacheTTL {
		return c.price, true
	}
	return 0, false
}

func setCachedPrice(symbol string, price float64) {
	stockCacheMux.Lock()
	defer stockCacheMux.Unlock()
	stockCache[symbol] = &stockPriceCache{price: price, updateTime: time.Now()}
}

func fetchUSStockPrice(symbol string) (float64, error) {
	if price, ok := getCachedPrice(symbol); ok {
		return price, nil
	}

	url := fmt.Sprintf("%s/q=us%s", tencentAPIURL, symbol)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://gu.qq.com/")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	reader := transform.NewReader(resp.Body, simplifiedchinese.GBK.NewDecoder())
	body, _ := io.ReadAll(reader)
	text := string(body)

	// 腾讯格式: v_usAAPL="...|...|当前价|..."
	parts := strings.Split(text, "~")
	if len(parts) >= 4 {
		price, err := strconv.ParseFloat(parts[3], 64)
		if err == nil && price > 0 {
			setCachedPrice(symbol, price)
			return price, nil
		}
	}
	return 0, fmt.Errorf("获取股票价格失败: %s", symbol)
}

func fetchUSStockPrices(symbols []string) map[string]float64 {
	prices := make(map[string]float64)
	for _, symbol := range symbols {
		if price, ok := getCachedPrice(symbol); ok {
			prices[symbol] = price
			continue
		}
		if price, err := fetchUSStockPrice(symbol); err == nil {
			prices[symbol] = price
		}
	}
	return prices
}

type CreateTradeRequest struct {
	AssetType string  `json:"asset_type"`
	Symbol    string  `json:"symbol"`
	Type      string  `json:"type"`
	Amount    float64 `json:"amount"`
	Price     float64 `json:"price"`
	Total     float64 `json:"total"`
	Currency  string  `json:"currency"`
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

type ExportData struct {
	Version     string        `json:"version"`
	Exported    string        `json:"exported"`
	Trades      []TradeExport `json:"trades"`
	Fingerprint string        `json:"fingerprint"`
}

type ImportPreviewRequest struct {
	Data ExportData `json:"data"`
}

type ConflictItem struct {
	Trade  TradeExport `json:"trade"`
	Reason string      `json:"reason"`
}

type ImportPreviewResponse struct {
	TotalTrades   int            `json:"total_trades"`
	NewTrades     int            `json:"new_trades"`
	Conflicts     int            `json:"conflicts"`
	ConflictItems []ConflictItem `json:"conflict_items"`
}

type ImportConfirmRequest struct {
	Data             ExportData `json:"data"`
	ConflictStrategy string     `json:"conflict_strategy"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// /api/prices/{symbol}?asset_type=... — 不经过 AuthMiddleware
	if strings.HasPrefix(path, "/api/prices/") {
		if r.Method == http.MethodGet {
			handleAssetPrice(w, r)
			return
		}
		utils.Error(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	// /api/portfolio/... 路由
	path = strings.TrimPrefix(path, "/api/portfolio")
	path = strings.TrimPrefix(path, "/")

	// 健康检查不需要认证
	if path == "health" {
		if r.Method == http.MethodGet {
			utils.Success(w, map[string]interface{}{
				"status": "healthy",
			})
			return
		}
		utils.Error(w, http.StatusNotFound, "接口不存在")
		return
	}

	// 所有其他路由都需要 JWT 认证
	middleware.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// 根据路径和方法分发
		if path == "" || path == "dashboard" {
			if r.Method == http.MethodGet {
				handleDashboard(w, r)
				return
			}
		}

		if path == "trades" || strings.HasPrefix(path, "trades/") {
			handleTrades(w, r)
			return
		}

		if path == "export" {
			if r.Method == http.MethodGet {
				handleExport(w, r)
				return
			}
		}

		if path == "import/preview" {
			if r.Method == http.MethodPost {
				handleImportPreview(w, r)
				return
			}
		}

		if path == "import/confirm" {
			if r.Method == http.MethodPost {
				handleImportConfirm(w, r)
				return
			}
		}

		utils.Error(w, http.StatusNotFound, "接口不存在")
	})(w, r)
}

// PortfolioItem 投资组合单项
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

// fetchCryptoPrices 获取加密货币价格 (CoinGecko API - 免费免Key)
// CoinGecko API文档: https://www.coingecko.com/api/documentation
func fetchCryptoPrices() (map[string]float64, map[string]float64, int64) {
	// CoinGecko ID 映射表 (币种 symbol -> CoinGecko ID)
	symbolToGeckoID := map[string]string{
		"BTC": "bitcoin", "ETH": "ethereum", "BNB": "binancecoin",
		"XRP": "ripple", "ADA": "cardano", "SOL": "solana",
		"DOGE": "dogecoin", "TRX": "tron", "AVAX": "avalanche-2",
		"HYPE": "hyperliquid", "POL": "polygon-ecosystem-token", "DOT": "polkadot",
	}

	// 构建 ids 参数
	ids := make([]string, 0, len(symbolToGeckoID))
	for _, id := range symbolToGeckoID {
		ids = append(ids, id)
	}
	idsParam := strings.Join(ids, ",")

	// CoinGecko 免费公共 API (无需 API Key)
	url := fmt.Sprintf(
		"https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd&include_24hr_change=true",
		idsParam,
	)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/json")
	// 添加 User-Agent 避免被限流
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; PortfolioApp/1.0)")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return map[string]float64{}, map[string]float64{}, 0
	}
	defer resp.Body.Close()

	prices := map[string]float64{}
	priceChanges := map[string]float64{}
	updatedAt := time.Now().Unix()

	if resp.StatusCode == http.StatusOK {
		// CoinGecko 返回格式: {"bitcoin": {"usd": 50000, "usd_24h_change": 2.5}, ...}
		var result map[string]struct {
			Usd          float64 `json:"usd"`
			Usd24hChange float64 `json:"usd_24h_change"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
			// 反向映射: geckoID -> symbol
			geckoIDToSymbol := make(map[string]string, len(symbolToGeckoID))
			for sym, gid := range symbolToGeckoID {
				geckoIDToSymbol[gid] = sym
			}

			for geckoID, data := range result {
				if symbol, ok := geckoIDToSymbol[geckoID]; ok {
					prices[symbol] = data.Usd
					// CoinGecko 返回的是百分比数值 (如 2.5 表示 +2.5%)
					// 转换为小数形式 (如 0.025) 保持与原有逻辑一致
					priceChanges[symbol] = data.Usd24hChange / 100
				}
			}
		}
	}

	return prices, priceChanges, updatedAt
}

// handleAssetPrice 获取单个资产价格 (/api/prices/{symbol}?asset_type=crypto|us_stock)
func handleAssetPrice(w http.ResponseWriter, r *http.Request) {
	symbol := strings.TrimPrefix(r.URL.Path, "/api/prices/")
	symbol = strings.TrimSuffix(symbol, "/")
	if symbol == "" {
		utils.BadRequest(w, "资产代码不能为空")
		return
	}

	assetType := r.URL.Query().Get("asset_type")
	if assetType == "" {
		assetType = "crypto"
	}

	switch assetType {
	case "crypto":
		if !supportedCryptos[symbol] {
			utils.BadRequest(w, fmt.Sprintf("不支持的加密货币: %s", symbol))
			return
		}
		prices, _, updatedAt := fetchCryptoPrices()
		price := prices[symbol]
		utils.Success(w, map[string]interface{}{
			"symbol":     symbol,
			"price":      price,
			"asset_type": "crypto",
			"currency":   "USD",
			"updated_at": updatedAt,
		})
	case "us_stock":
		if !supportedUSStocks[symbol] {
			utils.BadRequest(w, fmt.Sprintf("不支持的美股: %s", symbol))
			return
		}
		price, err := fetchUSStockPrice(symbol)
		if err != nil {
			utils.InternalError(w, fmt.Sprintf("获取股票价格失败: %v", err))
			return
		}
		utils.Success(w, map[string]interface{}{
			"symbol":     symbol,
			"price":      price,
			"asset_type": "us_stock",
			"currency":   "USD",
			"updated_at": time.Now().Unix(),
		})
	default:
		utils.BadRequest(w, fmt.Sprintf("不支持的资产类型: %s", assetType))
	}
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		utils.Unauthorized(w, "未授权")
		return
	}

	database := db.GetDB()

	var holdings []models.Holding
	database.Where("user_id = ?", userID).Find(&holdings)

	var trades []models.Trade
	database.Where("user_id = ?", userID).Order("created_at asc").Find(&trades)

	// 获取价格数据
	cryptoPrices, cryptoChanges, cryptoUpdatedAt := fetchCryptoPrices()

	usStockSymbols := make([]string, 0, len(supportedUSStocks))
	for sym := range supportedUSStocks {
		usStockSymbols = append(usStockSymbols, sym)
	}
	usStockPrices := fetchUSStockPrices(usStockSymbols)

	// 计算统计数据
	portfolio, cryptoValue, usStockValue, cashBalance, totalAssetsValue,
		totalUnrealizedPL, totalUnrealizedPLRate, totalRealizedPL, totalRealizedPLRate :=
		calculatePortfolioStats(holdings, cryptoPrices, usStockPrices, trades)

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"prices":                      cryptoPrices,
		"us_stock_prices":             usStockPrices,
		"price_changes":               cryptoChanges,
		"exchange_rates":              map[string]interface{}{},
		"crypto_updated_at":           cryptoUpdatedAt,
		"btc_price":                   cryptoPrices["BTC"],
		"portfolio":                   portfolio,
		"crypto_value":                cryptoValue,
		"us_stock_value":              usStockValue,
		"cash_balance":                cashBalance,
		"total_assets_value":          totalAssetsValue,
		"unrealized_profit_loss":      totalUnrealizedPL,
		"unrealized_profit_loss_rate": totalUnrealizedPLRate,
		"realized_profit_loss":        totalRealizedPL,
		"realized_profit_loss_rate":   totalRealizedPLRate,
		"total_profit_loss":           totalUnrealizedPL + totalRealizedPL,
		"value_change_24h":            0.0,
	})
}

// calculatePortfolioStats 计算投资组合统计
func calculatePortfolioStats(holdings []models.Holding, cryptoPrices, usStockPrices map[string]float64,
	trades []models.Trade) (portfolio []PortfolioItem, cryptoValue, usStockValue, cashBalance, totalAssetsValue,
	totalUnrealizedPL, totalUnrealizedPLRate, totalRealizedPL, totalRealizedPLRate float64) {

	portfolio = make([]PortfolioItem, 0, len(holdings))
	var totalUnrealizedPLVal, totalHistoricalCost float64

	// 按资产类型和代码分组计算成本跟踪
	type assetData struct {
		amount     float64
		cost       float64
		totalIn    float64
		realizedPL float64
	}
	assetDataMap := make(map[string]map[string]*assetData)

	// 遍历交易记录计算每项资产的成本
	for _, t := range trades {
		if t.Type == "recharge" {
			continue
		}

		if assetDataMap[t.AssetType] == nil {
			assetDataMap[t.AssetType] = make(map[string]*assetData)
		}
		if assetDataMap[t.AssetType][t.Symbol] == nil {
			assetDataMap[t.AssetType][t.Symbol] = &assetData{}
		}
		d := assetDataMap[t.AssetType][t.Symbol]

		switch t.Type {
		case "buy":
			d.amount += t.Amount
			d.cost += t.Total
			d.totalIn += t.Total
		case "sell":
			if d.amount > 0 && t.Amount > 0 {
				sellRatio := t.Amount / d.amount
				costRecovered := d.cost * sellRatio
				pl := t.Total - costRecovered

				d.realizedPL += pl
				d.cost -= costRecovered
				d.amount -= t.Amount
			}
		}
	}

	// 计算已实现盈亏
	for _, symbols := range assetDataMap {
		for _, d := range symbols {
			totalRealizedPL += d.realizedPL
			totalHistoricalCost += d.totalIn
		}
	}

	// 计算总实现盈亏率
	if totalHistoricalCost != 0 {
		totalRealizedPLRate = (totalRealizedPL / totalHistoricalCost) * 100
	}

	// 遍历持仓构建组合项
	for _, h := range holdings {
		if h.AssetType == models.AssetTypeCash && h.Symbol == "USD" {
			cashBalance = h.Amount
			continue
		}

		// 获取当前价格
		var price float64
		switch h.AssetType {
		case models.AssetTypeCrypto:
			price = cryptoPrices[h.Symbol]
		case models.AssetTypeUSStock:
			price = usStockPrices[h.Symbol]
		}

		marketValue := h.Amount * price

		// 从成本跟踪数据获取
		d := assetDataMap[h.AssetType][h.Symbol]
		var cost float64
		var realizedPL float64
		if d != nil {
			cost = d.cost
			realizedPL = d.realizedPL
		}

		// 累加到对应资产类型的总市值
		switch h.AssetType {
		case models.AssetTypeCrypto:
			cryptoValue += marketValue
		case models.AssetTypeUSStock:
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

		totalUnrealizedPLVal += profitLoss

		// 计算单项已实现盈亏率
		realizedPLRate := 0.0
		if d != nil && d.totalIn != 0 {
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

	totalAssetsValue = cryptoValue + usStockValue + cashBalance

	// 计算总未实现盈亏率
	totalCost := 0.0
	for _, h := range holdings {
		if h.AssetType != models.AssetTypeCash {
			if d := assetDataMap[h.AssetType][h.Symbol]; d != nil {
				totalCost += d.cost
			}
		}
	}
	if totalCost != 0 {
		totalUnrealizedPLRate = (totalUnrealizedPLVal / totalCost) * 100
	}

	totalUnrealizedPL = totalUnrealizedPLVal
	return
}

func handleTrades(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTrades(w, r)
	case http.MethodPost:
		createTrade(w, r)
	case http.MethodDelete:
		if strings.HasSuffix(r.URL.Path, "/trades") {
			clearTrades(w, r)
		} else {
			deleteTrade(w, r)
		}
	default:
		utils.Error(w, http.StatusMethodNotAllowed, "方法不允许")
	}
}

func getTrades(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		utils.Unauthorized(w, "未授权")
		return
	}

	var trades []models.Trade
	db.GetDB().Where("user_id = ?", userID).Order("created_at desc").Find(&trades)

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"trades": trades,
	})
}

func createTrade(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		utils.Unauthorized(w, "未授权")
		return
	}

	var req CreateTradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "无效的请求体")
		return
	}

	if req.Symbol == "" || req.Type == "" || req.Amount <= 0 || req.Price <= 0 {
		utils.BadRequest(w, "缺少必填字段或数值无效")
		return
	}

	if req.AssetType == "" {
		req.AssetType = models.AssetTypeCrypto
	}
	if req.Currency == "" {
		req.Currency = models.GetCurrencyByAssetType(req.AssetType)
	}
	if req.Total == 0 {
		req.Total = req.Amount * req.Price
	}

	// 业务层参数校验（资产白名单、充值规则）
	if err := validateTradeRequest(&req); err != nil {
		if _, ok := err.(*BusinessError); ok {
			utils.BadRequest(w, err.Error())
		} else {
			utils.InternalError(w, err.Error())
		}
		return
	}

	database := db.GetDB()

	// 使用事务保证数据一致性
	tx := database.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	trade := models.Trade{
		UUID:      uuid.New().String(),
		UserID:    userID,
		AssetType: req.AssetType,
		Symbol:    req.Symbol,
		Type:      req.Type,
		Amount:    req.Amount,
		Price:     req.Price,
		Total:     req.Total,
		Currency:  req.Currency,
	}

	if err := tx.Create(&trade).Error; err != nil {
		tx.Rollback()
		utils.InternalError(w, "创建交易失败")
		return
	}

	if err := recalcHoldings(tx, userID); err != nil {
		tx.Rollback()
		utils.InternalError(w, "更新持仓失败")
		return
	}

	if err := tx.Commit().Error; err != nil {
		utils.InternalError(w, "提交事务失败")
		return
	}

	utils.JSON(w, http.StatusOK, trade)
}

func deleteTrade(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		utils.Unauthorized(w, "未授权")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/portfolio/trades/")
	id, err := strconv.ParseUint(path, 10, 32)
	if err != nil {
		utils.BadRequest(w, "无效的交易ID")
		return
	}

	database := db.GetDB()

	result := database.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Trade{})
	if result.Error != nil {
		utils.InternalError(w, "删除交易失败")
		return
	}
	if result.RowsAffected == 0 {
		utils.NotFound(w, "交易记录不存在")
		return
	}

	if err := recalcHoldings(database, userID); err != nil {
		utils.InternalError(w, "更新持仓失败")
		return
	}

	utils.Success(w, map[string]interface{}{
		"message": "删除成功",
	})
}

func clearTrades(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		utils.Unauthorized(w, "未授权")
		return
	}

	database := db.GetDB()

	if err := database.Unscoped().Where("user_id = ?", userID).Delete(&models.Trade{}).Error; err != nil {
		utils.InternalError(w, "清空交易失败")
		return
	}

	if err := database.Unscoped().Where("user_id = ?", userID).Delete(&models.Holding{}).Error; err != nil {
		utils.InternalError(w, "清空持仓失败")
		return
	}

	utils.Success(w, map[string]interface{}{
		"message": "清空成功",
	})
}

func recalcHoldings(tx *gorm.DB, uid uint) error {
	if err := tx.Where("user_id = ?", uid).Delete(&models.Holding{}).Error; err != nil {
		return err
	}

	var trades []models.Trade
	if err := tx.Where("user_id = ?", uid).Order("created_at asc").Find(&trades).Error; err != nil {
		return err
	}

	holdings := make(map[string]*models.Holding)
	var cashHolding *models.Holding

	for _, t := range trades {
		if t.Type == "recharge" {
			if cashHolding == nil {
				cashHolding = &models.Holding{
					UserID:    uid,
					AssetType: models.AssetTypeCash,
					Symbol:    "USD",
					Currency:  models.CurrencyUSD,
				}
			}
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
			}
		}

		switch t.Type {
		case "buy":
			holdings[key].Amount += t.Amount
			if cashHolding == nil {
				cashHolding = &models.Holding{
					UserID:    uid,
					AssetType: models.AssetTypeCash,
					Symbol:    "USD",
					Currency:  models.CurrencyUSD,
				}
			}
			cashHolding.Amount -= t.Total
		case "sell":
			holdings[key].Amount -= t.Amount
			if cashHolding == nil {
				cashHolding = &models.Holding{
					UserID:    uid,
					AssetType: models.AssetTypeCash,
					Symbol:    "USD",
					Currency:  models.CurrencyUSD,
				}
			}
			cashHolding.Amount += t.Total
		}
	}

	for _, h := range holdings {
		if h.Amount != 0 {
			if err := tx.Create(h).Error; err != nil {
				return err
			}
		}
	}

	if cashHolding != nil && cashHolding.Amount != 0 {
		if err := tx.Create(cashHolding).Error; err != nil {
			return err
		}
	}

	return nil
}

func handleExport(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		utils.Unauthorized(w, "未授权")
		return
	}

	var trades []models.Trade
	if err := db.GetDB().Where("user_id = ?", userID).Order("created_at asc").Find(&trades).Error; err != nil {
		utils.InternalError(w, "获取交易记录失败")
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

	exported := time.Now().Format("2006-01-02 15:04:05")
	exportData := ExportData{
		Version:     "1.0",
		Exported:    exported,
		Trades:      tradeExports,
		Fingerprint: calculateFingerprint("1.0", exported, tradeExports),
	}

	utils.Success(w, map[string]interface{}{
		"data": exportData,
	})
}

func calculateFingerprint(version, exported string, trades []TradeExport) string {
	data := struct {
		Version  string        `json:"version"`
		Exported string        `json:"exported"`
		Trades   []TradeExport `json:"trades"`
	}{
		Version:  version,
		Exported: exported,
		Trades:   trades,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}

func handleImportPreview(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		utils.Unauthorized(w, "未授权")
		return
	}

	var req ImportPreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "无效的请求数据")
		return
	}

	database := db.GetDB()

	var existingTrades []models.Trade
	database.Where("user_id = ?", userID).Find(&existingTrades)

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
		key := trade.UUID
		if key == "" {
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

	utils.Success(w, map[string]interface{}{
		"preview": preview,
	})
}

func handleImportConfirm(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		utils.Unauthorized(w, "未授权")
		return
	}

	var req ImportConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "无效的请求数据")
		return
	}

	if req.Data.Fingerprint != "" {
		expectedFingerprint := calculateFingerprint(req.Data.Version, req.Data.Exported, req.Data.Trades)
		if expectedFingerprint != req.Data.Fingerprint {
			utils.BadRequest(w, "数据指纹校验失败，文件可能已被篡改")
			return
		}
	}

	if req.ConflictStrategy != "skip" && req.ConflictStrategy != "overwrite" {
		req.ConflictStrategy = "skip"
	}

	database := db.GetDB()

	var existingTrades []models.Trade
	database.Where("user_id = ?", userID).Find(&existingTrades)

	existingUUIDs := make(map[string]uint)
	for _, t := range existingTrades {
		if t.UUID != "" {
			existingUUIDs[t.UUID] = t.ID
		}
	}

	var imported, skipped, overwritten int

	tx := database.Begin()

	for _, trade := range req.Data.Trades {
		key := trade.UUID
		if key == "" {
			key = fmt.Sprintf("%s_%s_%s_%s", trade.AssetType, trade.Symbol, trade.Type, trade.CreatedAt)
		}

		if existingID, exists := existingUUIDs[key]; exists {
			if req.ConflictStrategy == "overwrite" {
				if err := tx.Unscoped().Delete(&models.Trade{}, existingID).Error; err != nil {
					tx.Rollback()
					utils.InternalError(w, "删除旧记录失败")
					return
				}
				overwritten++
			} else {
				skipped++
				continue
			}
		}

		createdAt, _ := time.Parse("2006-01-02 15:04:05", trade.CreatedAt)
		if createdAt.IsZero() {
			createdAt = time.Now()
		}

		newTrade := models.Trade{
			UUID:      trade.UUID,
			UserID:    userID,
			AssetType: trade.AssetType,
			Symbol:    trade.Symbol,
			Type:      trade.Type,
			Amount:    trade.Amount,
			Price:     trade.Price,
			Total:     trade.Total,
			Currency:  trade.Currency,
		}
		if newTrade.UUID == "" {
			newTrade.UUID = uuid.New().String()
		}
		newTrade.CreatedAt = createdAt

		if err := tx.Create(&newTrade).Error; err != nil {
			tx.Rollback()
			utils.InternalError(w, "创建交易记录失败")
			return
		}

		imported++
	}

	if err := recalcHoldings(tx, userID); err != nil {
		tx.Rollback()
		utils.InternalError(w, "重新计算持仓失败")
		return
	}

	tx.Commit()

	utils.Success(w, map[string]interface{}{
		"imported":    imported,
		"skipped":     skipped,
		"overwritten": overwritten,
	})
}
