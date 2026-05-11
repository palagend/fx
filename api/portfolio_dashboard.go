package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"

	"gitee.com/palagend/fx/config"
	"gitee.com/palagend/fx/models"
	"gitee.com/palagend/fx/utils"
)


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
