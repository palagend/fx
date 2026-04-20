package main

import (
	"context"
	"embed"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitee.com/palagend/fx/api"
	"gitee.com/palagend/fx/config"
	"gitee.com/palagend/fx/models"
	"github.com/gin-gonic/gin"
)

//go:embed web/dist
var embeddedWebDist embed.FS

func main() {
	// 初始化配置
	cfg := config.InitConfig()

	// 初始化数据库
	db := config.InitDB(cfg)

	if err := models.AutoMigrate(db); err != nil {
		panic("数据库迁移失败: " + err.Error())
	}

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)
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

	// 在后台启动HTTP服务器
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic("HTTP服务器启动失败: " + err.Error())
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 优雅关闭HTTP服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		panic("HTTP服务器关闭失败: " + err.Error())
	}

	// 关闭数据库连接
	if err := config.CloseDB(); err != nil {
		panic("数据库关闭失败: " + err.Error())
	}
}
