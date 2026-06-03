package auth

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"api/db"
	"api/middleware"
	"api/models"
	"api/utils"

	"gorm.io/gorm"
)

// ---------- 请求结构体 ----------

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"Username"`
	Password string `json:"Password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

// Handler 统一入口 /api/auth/
func Handler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/auth")
	path = strings.TrimPrefix(path, "/")

	switch path {
	case "register":
		if r.Method == http.MethodPost {
			handleRegister(w, r)
			return
		}
	case "login":
		if r.Method == http.MethodPost {
			handleLogin(w, r)
			return
		}
	case "logout":
		if r.Method == http.MethodPost {
			middleware.AuthMiddleware(handleLogout)(w, r)
			return
		}
	case "refresh":
		if r.Method == http.MethodPost {
			handleRefresh(w, r)
			return
		}
	case "change-password":
		if r.Method == http.MethodPost {
			middleware.AuthMiddleware(handleChangePassword)(w, r)
			return
		}
	}

	utils.Error(w, http.StatusNotFound, "接口不存在")
}

// handleRegister 注册
func handleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"请求格式错误"}`, http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		http.Error(w, `{"error":"用户名、邮箱和密码不能为空"}`, http.StatusBadRequest)
		return
	}

	database := db.GetDB()

	// 检查用户名是否已存在
	var existing models.User
	if err := database.Where("username = ?", req.Username).First(&existing).Error; err == nil {
		http.Error(w, `{"error":"用户名已存在"}`, http.StatusConflict)
		return
	}

	// 检查邮箱是否已存在
	if err := database.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		http.Error(w, `{"error":"邮箱已被注册"}`, http.StatusConflict)
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, `{"error":"密码加密失败"}`, http.StatusInternalServerError)
		return
	}

	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := database.Create(&user).Error; err != nil {
		http.Error(w, `{"error":"注册失败"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true}`))
}

// handleLogin 登录
func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"请求格式错误"}`, http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, `{"error":"用户名和密码不能为空"}`, http.StatusBadRequest)
		return
	}

	database := db.GetDB()

	var user models.User
	if err := database.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, `{"error":"用户名或密码错误"}`, http.StatusUnauthorized)
			return
		}
		http.Error(w, `{"error":"登录失败"}`, http.StatusInternalServerError)
		return
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		http.Error(w, `{"error":"用户名或密码错误"}`, http.StatusUnauthorized)
		return
	}

	// 生成令牌
	tokenPair, err := utils.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		http.Error(w, `{"error":"生成令牌失败"}`, http.StatusInternalServerError)
		return
	}

	// 保存刷新令牌
	refreshToken := models.RefreshToken{
		UserID:    user.ID,
		Token:     tokenPair.RefreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	database.Create(&refreshToken)

	// 返回用户信息和令牌（前端通过 fetch 直接解析）
	resp := map[string]interface{}{
		"tokens": map[string]string{
			"access_token":  tokenPair.AccessToken,
			"refresh_token": tokenPair.RefreshToken,
		},
		"user": map[string]interface{}{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"created_at": user.CreatedAt.Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleLogout 登出（清除用户的所有刷新令牌）
func handleLogout(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, `{"error":"未授权"}`, http.StatusUnauthorized)
		return
	}

	database := db.GetDB()
	database.Where("user_id = ?", userID).Delete(&models.RefreshToken{})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true}`))
}

// handleRefresh 刷新令牌
func handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"请求格式错误"}`, http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		http.Error(w, `{"error":"refresh_token 不能为空"}`, http.StatusBadRequest)
		return
	}

	database := db.GetDB()

	// 在数据库中查找刷新令牌
	var storedToken models.RefreshToken
	if err := database.Where("token = ?", req.RefreshToken).First(&storedToken).Error; err != nil {
		http.Error(w, `{"error":"无效的刷新令牌"}`, http.StatusUnauthorized)
		return
	}

	// 检查是否过期
	if time.Now().After(storedToken.ExpiresAt) {
		database.Delete(&storedToken)
		http.Error(w, `{"error":"刷新令牌已过期"}`, http.StatusUnauthorized)
		return
	}

	// 解析旧的刷新令牌获取用户信息
	claims, err := utils.ParseToken(req.RefreshToken)
	if err != nil {
		database.Delete(&storedToken)
		http.Error(w, `{"error":"无效的刷新令牌"}`, http.StatusUnauthorized)
		return
	}

	// 删除旧的刷新令牌（一次性使用）
	database.Delete(&storedToken)

	// 生成新令牌对
	tokenPair, err := utils.GenerateTokenPair(claims.UserID, claims.Username)
	if err != nil {
		http.Error(w, `{"error":"生成令牌失败"}`, http.StatusInternalServerError)
		return
	}

	// 保存新的刷新令牌
	newRefreshToken := models.RefreshToken{
		UserID:    claims.UserID,
		Token:     tokenPair.RefreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	database.Create(&newRefreshToken)

	resp := map[string]interface{}{
		"tokens": map[string]string{
			"access_token":  tokenPair.AccessToken,
			"refresh_token": tokenPair.RefreshToken,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleChangePassword 修改密码
func handleChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, `{"error":"未授权"}`, http.StatusUnauthorized)
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"请求格式错误"}`, http.StatusBadRequest)
		return
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		http.Error(w, `{"error":"旧密码和新密码不能为空"}`, http.StatusBadRequest)
		return
	}

	if req.NewPassword == req.OldPassword {
		http.Error(w, `{"error":"新密码不能与旧密码相同"}`, http.StatusBadRequest)
		return
	}

	database := db.GetDB()

	var user models.User
	if err := database.First(&user, userID).Error; err != nil {
		http.Error(w, `{"error":"用户不存在"}`, http.StatusNotFound)
		return
	}

	if !utils.CheckPassword(req.OldPassword, user.Password) {
		http.Error(w, `{"error":"旧密码错误"}`, http.StatusUnauthorized)
		return
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, `{"error":"密码加密失败"}`, http.StatusInternalServerError)
		return
	}

	database.Model(&user).Update("password", hashedPassword)

	// 修改密码后清除所有刷新令牌，强制重新登录
	database.Where("user_id = ?", userID).Delete(&models.RefreshToken{})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true}`))
}
