package db

import (
	"fmt"
	"os"
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db   *gorm.DB
	once sync.Once
)

// GetDB 返回数据库连接（单例模式，适配 Serverless）
func GetDB() *gorm.DB {
	once.Do(func() {
		dsn := os.Getenv("DATABASE_URL")
		if dsn == "" {
			// 如果没有设置 DATABASE_URL，尝试从各个环境变量构建
			host := os.Getenv("DB_HOST")
			port := os.Getenv("DB_PORT")
			user := os.Getenv("DB_USER")
			password := os.Getenv("DB_PASSWORD")
			dbname := os.Getenv("DB_NAME")
			charset := os.Getenv("DB_CHARSET")

			if host == "" || user == "" || password == "" || dbname == "" {
				panic("数据库配置不完整，需要设置 DATABASE_URL 或 DB_HOST, DB_USER, DB_PASSWORD, DB_NAME")
			}

			if charset == "" {
				charset = "utf8mb4"
			}
			if port == "" {
				port = "3306"
			}

			// MySQL DSN 格式: user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
			dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
				user, password, host, port, dbname, charset)
		}

		var err error
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			panic(fmt.Sprintf("failed to connect database: %v", err))
		}
	})

	return db
}

// AutoMigrate 自动迁移数据库结构
func AutoMigrate(models ...interface{}) error {
	return GetDB().AutoMigrate(models...)
}
