package api

import (
	"encoding/json"
	"fmt"
	"net/http"

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
