package api

import (
	"encoding/json"
	"net/http"

	"api/db"
	"api/models"
	"api/utils"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.Error(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "无效的请求体")
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		utils.BadRequest(w, "用户名、邮箱和密码不能为空")
		return
	}

	if len(req.Password) < 6 {
		utils.BadRequest(w, "密码长度至少6位")
		return
	}

	database := db.GetDB()

	var existingUser models.User
	if err := database.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		utils.Error(w, http.StatusConflict, "用户名已存在")
		return
	}

	if err := database.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		utils.Error(w, http.StatusConflict, "邮箱已被注册")
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.InternalError(w, "密码加密失败")
		return
	}

	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := database.Create(&user).Error; err != nil {
		utils.InternalError(w, "用户创建失败")
		return
	}

	tokenPair, err := utils.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		utils.InternalError(w, "令牌生成失败")
		return
	}

	utils.Success(w, map[string]interface{}{
		"message": "注册成功",
		"user": map[string]interface{}{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"created_at": user.CreatedAt.Format("2006-01-02 15:04:05"),
		},
		"tokens": tokenPair,
	})
}
