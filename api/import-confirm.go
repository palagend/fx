package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"api/db"
	"api/middleware"
	"api/models"
	"api/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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

type ImportConfirmRequest struct {
	Data             ExportData `json:"data"`
	ConflictStrategy string     `json:"conflict_strategy"`
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

func recalcHoldingsAfterImport(tx *gorm.DB, uid uint) error {
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

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.Error(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

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

	if err := recalcHoldingsAfterImport(tx, userID); err != nil {
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
