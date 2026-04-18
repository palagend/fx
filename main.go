package main

import (
	"embed"
	"io/fs"
	"net/http"

	"gitee.com/palagend/fx/api"
	"gitee.com/palagend/fx/config"
	"gitee.com/palagend/fx/models"
	"github.com/gin-gonic/gin"
)

//go:embed web/dist
var embeddedWebDist embed.FS

func main() {
	db := config.InitDB()

	if err := models.AutoMigrate(db); err != nil {
		panic("数据库迁移失败: " + err.Error())
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	webDist, err := fs.Sub(embeddedWebDist, "web/dist")
	if err != nil {
		panic(err)
	}

	api.RegisterRoutes(r)

	r.NoRoute(func(c *gin.Context) {
		handler := http.FileServer(http.FS(webDist))
		handler.ServeHTTP(c.Writer, c.Request)
	})

	r.Run(":8080")
}
