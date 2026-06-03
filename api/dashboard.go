package api

import (
	"net/http"

	"api/db"
	"api/middleware"
	"api/models"
	"api/utils"
)

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

	database := db.GetDB()

	var holdings []models.Holding
	database.Where("user_id = ?", userID).Find(&holdings)

	var recentTrades []models.Trade
	database.Where("user_id = ?", userID).Order("created_at desc").Limit(5).Find(&recentTrades)

	var totalTrades int64
	database.Model(&models.Trade{}).Where("user_id = ?", userID).Count(&totalTrades)

	utils.Success(w, map[string]interface{}{
		"holdings":      holdings,
		"recent_trades": recentTrades,
		"stats": map[string]interface{}{
			"total_trades": totalTrades,
		},
	})
}
