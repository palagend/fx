package db

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"api/models"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
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
			host := os.Getenv("DB_HOST")
			port := os.Getenv("DB_PORT")
			user := os.Getenv("DB_USER")
			password := os.Getenv("DB_PASSWORD")
			dbname := os.Getenv("DB_NAME")

			if host == "" || user == "" || password == "" || dbname == "" {
				panic("数据库配置不完整，需要设置 DATABASE_URL 或 DB_HOST, DB_USER, DB_PASSWORD, DB_NAME")
			}

			// 检查是否使用 PostgreSQL
			dbType := os.Getenv("DB_TYPE")
			if dbType == "" {
				dbType = "mysql"
			}

			switch strings.ToLower(dbType) {
			case "postgres", "postgresql", "pg":
				if port == "" {
					port = "5432"
				}
				dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
					host, port, user, password, dbname)
			default:
				if port == "" {
					port = "3306"
				}
				charset := os.Getenv("DB_CHARSET")
				if charset == "" {
					charset = "utf8mb4"
				}
				dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
					user, password, host, port, dbname, charset)
			}
		}

		var dialector gorm.Dialector
		if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
			dialector = postgres.Open(dsn)
		} else {
			dialector = mysql.Open(dsn)
		}

		var err error
		db, err = gorm.Open(dialector, &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			panic(fmt.Sprintf("failed to connect database: %v", err))
		}

		// 自动迁移表结构（幂等，仅新增字段/表，不删除）
		if err := db.AutoMigrate(
			&models.User{},
			&models.RefreshToken{},
			&models.Trade{},
			&models.Holding{},
			&models.ExchangeRate{},
		); err != nil {
			panic(fmt.Sprintf("failed to auto migrate: %v", err))
		}
	})

	return db
}

// AutoMigrate 保留供外部手动调用
func AutoMigrate(models ...interface{}) error {
	return GetDB().AutoMigrate(models...)
}
