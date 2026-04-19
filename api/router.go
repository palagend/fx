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
				protected.POST("/change-password", ChangePassword)
			}
		}

		// 资产组合相关接口
		portfolioGroup := apiGroup.Group("/portfolio")
		portfolioGroup.Use(middleware.AuthMiddleware())
		{
			portfolioGroup.GET("/dashboard", GetDashboard)    // 获取仪表盘聚合数据
			portfolioGroup.GET("/trades", GetTrades)          // 获取交易记录
			portfolioGroup.POST("/trades", CreateTrade)       // 创建交易
			portfolioGroup.DELETE("/trades/:id", DeleteTrade) // 删除交易
			portfolioGroup.DELETE("/trades", ClearTrades)     // 清空交易
			portfolioGroup.GET("/export", ExportDataHandler)  // 导出数据
			portfolioGroup.POST("/import/preview", ImportPreviewHandler) // 导入预览
			portfolioGroup.POST("/import/confirm", ImportConfirmHandler) // 确认导入
		}

		// 价格相关接口（公开访问）
		priceGroup := apiGroup.Group("/prices")
		{
			priceGroup.GET("/:symbol", GetAssetPrice)
		}
	}
}
