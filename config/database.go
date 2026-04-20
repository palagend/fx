package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBType 数据库类型
type DBType string

const (
	// DBTypeMySQL MySQL数据库
	DBTypeMySQL DBType = "mysql"
	// DBTypePostgres PostgreSQL数据库
	DBTypePostgres DBType = "postgres"
	// DBTypeSQLite SQLite数据库
	DBTypeSQLite DBType = "sqlite"
)

// PoolConfig 连接池配置
type PoolConfig struct {
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`     // 最大空闲连接数
	MaxOpenConns    int           `mapstructure:"max_open_conns"`     // 最大打开连接数
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`  // 连接最大生命周期
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"` // 连接最大空闲时间
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type     DBType     `mapstructure:"type"`      // 数据库类型: mysql, postgres, sqlite
	Host     string     `mapstructure:"host"`      // 主机地址
	Port     int        `mapstructure:"port"`      // 端口
	User     string     `mapstructure:"user"`      // 用户名
	Password string     `mapstructure:"password"`  // 密码
	Name     string     `mapstructure:"name"`      // 数据库名
	Charset  string     `mapstructure:"charset"`   // 字符集
	SSLMode  string     `mapstructure:"ssl_mode"`  // SSL模式 (仅PostgreSQL)
	Path     string     `mapstructure:"path"`      // 数据库文件路径 (仅SQLite)
	Pool     PoolConfig `mapstructure:"pool"`      // 连接池配置
	LogLevel string     `mapstructure:"log_level"` // 日志级别: silent, error, warn, info
}

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB(cfg *Config) *gorm.DB {
	db, err := createDBConnection(cfg)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 配置连接池
	if err := configurePool(db, cfg); err != nil {
		log.Fatalf("配置连接池失败: %v", err)
	}

	DB = db
	fmt.Printf("数据库连接成功 [%s]\n", cfg.Database.Type)
	return db
}

// createDBConnection 创建数据库连接
func createDBConnection(cfg *Config) (*gorm.DB, error) {
	dialector, err := getDialector(cfg.Database)
	if err != nil {
		return nil, err
	}

	gormConfig := &gorm.Config{
		Logger: createLogger(cfg.Database.LogLevel),
	}

	return gorm.Open(dialector, gormConfig)
}

// getDialector 根据数据库类型获取对应的dialector
func getDialector(dbConfig DatabaseConfig) (gorm.Dialector, error) {
	switch dbConfig.Type {
	case DBTypeMySQL:
		return mysql.Open(buildMySQLDSN(dbConfig)), nil
	case DBTypePostgres:
		return postgres.Open(buildPostgresDSN(dbConfig)), nil
	case DBTypeSQLite:
		return sqlite.Open(buildSQLiteDSN(dbConfig)), nil
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", dbConfig.Type)
	}
}

// buildMySQLDSN 构建MySQL连接字符串
func buildMySQLDSN(dbConfig DatabaseConfig) string {
	charset := dbConfig.Charset
	if charset == "" {
		charset = "utf8mb4"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=UTC",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Name,
		charset,
	)
}

// buildPostgresDSN 构建PostgreSQL连接字符串
func buildPostgresDSN(dbConfig DatabaseConfig) string {
	sslMode := dbConfig.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Name,
		sslMode,
	)
}

// buildSQLiteDSN 构建SQLite连接字符串
func buildSQLiteDSN(dbConfig DatabaseConfig) string {
	path := dbConfig.Path
	if path == "" {
		path = dbConfig.Name
	}
	if path == "" {
		path = "fx.db"
	}
	return path
}

// createLogger 创建GORM日志记录器
func createLogger(logLevel string) logger.Interface {
	level := parseLogLevel(logLevel)

	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  level,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  false,
		},
	)
}

// parseLogLevel 解析日志级别字符串
func parseLogLevel(level string) logger.LogLevel {
	switch level {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		// 默认根据环境判断
		if globalConfig != nil && globalConfig.Server.Mode == "debug" {
			return logger.Info
		}
		return logger.Silent
	}
}

// configurePool 配置连接池
func configurePool(db *gorm.DB, cfg *Config) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	pool := cfg.Database.Pool

	// 设置最大空闲连接数
	if pool.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(pool.MaxIdleConns)
	} else {
		sqlDB.SetMaxIdleConns(10) // 默认值
	}

	// 设置最大打开连接数
	if pool.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(pool.MaxOpenConns)
	} else {
		sqlDB.SetMaxOpenConns(100) // 默认值
	}

	// 设置连接最大生命周期
	if pool.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(pool.ConnMaxLifetime)
	} else {
		sqlDB.SetConnMaxLifetime(time.Hour) // 默认1小时
	}

	// 设置连接最大空闲时间
	if pool.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(pool.ConnMaxIdleTime)
	}

	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	if DB == nil {
		log.Fatal("数据库未初始化，请先调用 InitDB")
	}
	return DB
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// PingDB 检查数据库连接是否正常
func PingDB() error {
	if DB == nil {
		return fmt.Errorf("数据库未初始化")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Ping()
}

// GetPoolStats 获取连接池统计信息
func GetPoolStats() (map[string]interface{}, error) {
	if DB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration,
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}, nil
}
