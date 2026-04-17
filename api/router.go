package api

import (
	"net/http"

	"gitee.com/palagend/fx/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	apiGroup := r.Group("/api")
	{

		apiGroup.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "healthy",
			})
		})

		authGroup := apiGroup.Group("/auth")
		{
			authGroup.POST("/register", Register)
			authGroup.POST("/login", Login)
			authGroup.POST("/refresh", RefreshToken)
			authGroup.POST("/logout", Logout)

			protected := authGroup.Group("/")
			protected.Use(middleware.AuthMiddleware())
			{
				protected.GET("/me", GetMe)
				protected.POST("/logout-all", LogoutAll)
			}
		}
	}
}
