package main

import (
	"embed"
	"io/fs"
	"net/http"

	"gitee.com/palagend/fx/api"
	"github.com/gin-gonic/gin"
)

//go:embed web/dist
var embeddedWebDist embed.FS

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(gin.Recovery())

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
