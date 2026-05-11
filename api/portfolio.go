package api

import (
	"fmt"
	"strconv"
	"time"

	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"

	"gitee.com/palagend/fx/config"
	"gitee.com/palagend/fx/models"
	"gitee.com/palagend/fx/utils"
	"gorm.io/gorm"
)


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
