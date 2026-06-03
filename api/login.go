package api

import (
	"encoding/json"
	"net/http"
	"time"

	"api/db"
	"api/models"
	"api/utils"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.Error(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "无效的请求体")
		return
	}

	if req.Username == "" || req.Password == "" {
		utils.BadRequest(w, "用户名和密码不能为空")
		return
	}

	database := db.GetDB()

	var user models.User
	if err := database.Where("username = ?", req.Username).First(&user).Error; err != nil {
		utils.Unauthorized(w, "用户名或密码错误")
		return
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		utils.Unauthorized(w, "用户名或密码错误")
		return
	}

	tokenPair, err := utils.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		utils.InternalError(w, "令牌生成失败")
		return
	}

	refreshToken := models.RefreshToken{
		UserID:    user.ID,
		Token:     tokenPair.RefreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := database.Create(&refreshToken).Error; err != nil {
		utils.InternalError(w, "刷新令牌保存失败")
		return
	}

	utils.Success(w, map[string]interface{}{
		"message": "登录成功",
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
		"tokens": tokenPair,
	})
}
