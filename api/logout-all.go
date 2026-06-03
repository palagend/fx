package api

import (
	"net/http"

	"api/db"
	"api/middleware"
	"api/models"
	"api/utils"
)

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

	if err := db.GetDB().Where("user_id = ?", userID).Delete(&models.RefreshToken{}).Error; err != nil {
		utils.InternalError(w, "登出失败")
		return
	}

	utils.Success(w, map[string]interface{}{
		"message": "已从所有设备登出",
	})
}
