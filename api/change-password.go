package api

import (
	"encoding/json"
	"net/http"

	"api/db"
	"api/middleware"
	"api/models"
	"api/utils"
)

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
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

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.BadRequest(w, "无效的请求体")
		return
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		utils.BadRequest(w, "旧密码和新密码不能为空")
		return
	}

	if len(req.NewPassword) < 6 {
		utils.BadRequest(w, "新密码长度至少6位")
		return
	}

	database := db.GetDB()

	var user models.User
	if err := database.First(&user, userID).Error; err != nil {
		utils.NotFound(w, "用户不存在")
		return
	}

	if !utils.CheckPassword(req.OldPassword, user.Password) {
		utils.Unauthorized(w, "旧密码错误")
		return
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		utils.InternalError(w, "密码加密失败")
		return
	}

	user.Password = hashedPassword
	if err := database.Save(&user).Error; err != nil {
		utils.InternalError(w, "密码更新失败")
		return
	}

	utils.Success(w, map[string]interface{}{
		"message": "密码修改成功",
	})
}
