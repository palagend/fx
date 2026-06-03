package middleware

import (
	"context"
	"net/http"
	"strings"

	"api/utils"
)

// AuthMiddleware 验证 JWT 令牌
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.Unauthorized(w, "缺少授权头")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.Unauthorized(w, "授权头格式错误")
			return
		}

		tokenString := parts[1]
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			utils.Unauthorized(w, "无效的令牌")
			return
		}

		if claims.TokenType != "access" {
			utils.Unauthorized(w, "无效的令牌类型")
			return
		}

		// 将用户信息存入上下文
		ctx := context.WithValue(r.Context(), "userID", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)
		next(w, r.WithContext(ctx))
	}
}

// GetUserID 从上下文中获取用户ID
func GetUserID(r *http.Request) (uint, bool) {
	userID, ok := r.Context().Value("userID").(uint)
	return userID, ok
}
