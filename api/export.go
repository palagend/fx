package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"api/db"
	"api/middleware"
	"api/models"
	"api/utils"
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

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.Error(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

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
