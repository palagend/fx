package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"api/db"
	"api/middleware"
	"api/models"
	"api/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateTradeRequest struct {
	AssetType string  `json:"asset_type"`
	Symbol    string  `json:"symbol"`
	Type      string  `json:"type"`
	Amount    float64 `json:"amount"`
	Price     float64 `json:"price"`
	Total     float64 `json:"total"`
	Currency  string  `json:"currency"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTrades(w, r)
	case http.MethodPost:
		createTrade(w, r)
	case http.MethodDelete:
		// 检查是否是清空操作
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

	utils.Success(w, trades)
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

	database := db.GetDB()
	if err := database.Create(&trade).Error; err != nil {
		utils.InternalError(w, "创建交易失败")
		return
	}

	if err := recalcHoldings(database, userID); err != nil {
		utils.InternalError(w, "更新持仓失败")
		return
	}

	utils.Success(w, trade)
}

func deleteTrade(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		utils.Unauthorized(w, "未授权")
		return
	}

	// 从 URL 获取 ID: /api/trades/123
	path := strings.TrimPrefix(r.URL.Path, "/api/trades/")
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
