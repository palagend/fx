package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() *gorm.DB {
	dsn := "admin:ctsi@Passw0rd@tcp(172.23.112.1:3306)/insight_onchain?charset=utf8mb4&parseTime=True&loc=UTC"

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("获取数据库实例失败: %v", err)
	}

	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db
	fmt.Println("数据库连接成功")
	return db
}

func GetDB() *gorm.DB {
	if DB == nil {
		return InitDB()
	}
	return DB
}
