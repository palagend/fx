package api

import (
	"net/http"

	"api/utils"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	utils.Success(w, map[string]interface{}{
		"status": "healthy",
	})
}
