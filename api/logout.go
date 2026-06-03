package api

import (
	"encoding/json"
	"net/http"

	"api/db"
	"api/models"
	"api/utils"
)

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.Error(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "无效的请求体")
		return
	}

	if req.RefreshToken != "" {
		db.GetDB().Where("token = ?", req.RefreshToken).Delete(&models.RefreshToken{})
	}

	utils.Success(w, map[string]interface{}{
		"message": "登出成功",
	})
}
