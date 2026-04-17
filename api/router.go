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
			portfolioGroup.GET("", GetPortfolio)               // 获取完整资产组合
			portfolioGroup.GET("/holdings", GetHoldings)       // 获取持仓
			portfolioGroup.GET("/investments", GetInvestments) // 获取投资记录（实现盈亏）
			portfolioGroup.GET("/trades", GetTrades)           // 获取交易记录
			portfolioGroup.POST("/trades", CreateTrade)        // 创建交易
			portfolioGroup.DELETE("/trades/:id", DeleteTrade)  // 删除交易
			portfolioGroup.DELETE("/trades", ClearTrades)      // 清空交易
		}

		// 价格相关接口（公开访问）
		priceGroup := apiGroup.Group("/prices")
		{
			priceGroup.GET("", GetAllPrices)
			priceGroup.GET("/:symbol", GetAssetPrice)
		}
	}
}
