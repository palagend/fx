package api

import (
	"encoding/json"
	"net/http"
	"time"

	"api/db"
	"api/models"
	"api/utils"
)

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.Error(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "无效的请求体")
		return
	}

	if req.RefreshToken == "" {
		utils.BadRequest(w, "刷新令牌不能为空")
		return
	}

	claims, err := utils.ParseToken(req.RefreshToken)
	if err != nil {
		utils.Unauthorized(w, "无效的刷新令牌")
		return
	}

	if claims.TokenType != "refresh" {
		utils.Unauthorized(w, "无效的令牌类型")
		return
	}

	database := db.GetDB()

	var storedToken models.RefreshToken
	if err := database.Where("token = ? AND expires_at > ?", req.RefreshToken, time.Now()).First(&storedToken).Error; err != nil {
		utils.Unauthorized(w, "刷新令牌已过期或不存在")
		return
	}

	tokenPair, err := utils.GenerateTokenPair(claims.UserID, claims.Username)
	if err != nil {
		utils.InternalError(w, "令牌生成失败")
		return
	}

	database.Where("token = ?", req.RefreshToken).Delete(&models.RefreshToken{})

	newRefreshToken := models.RefreshToken{
		UserID:    claims.UserID,
		Token:     tokenPair.RefreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := database.Create(&newRefreshToken).Error; err != nil {
		utils.InternalError(w, "刷新令牌保存失败")
		return
	}

	utils.Success(w, map[string]interface{}{
		"tokens": tokenPair,
	})
}
