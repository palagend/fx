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

	var user models.User
	if err := db.GetDB().First(&user, userID).Error; err != nil {
		utils.NotFound(w, "用户不存在")
		return
	}

	utils.Success(w, map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"created_at": user.CreatedAt.Format("2006-01-02 15:04:05"),
	})
}
