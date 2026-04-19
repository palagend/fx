package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	// 数据库配置
	Database DatabaseConfig `mapstructure:"database"`
	
	// API Keys
	API APIConfig `mapstructure:"api"`
	
	// 服务器配置
	Server ServerConfig `mapstructure:"server"`
	
	// JWT配置
	JWT JWTConfig `mapstructure:"jwt"`
	
	// 日志配置
	Log LogConfig `mapstructure:"log"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	Charset  string `mapstructure:"charset"`
}

// APIConfig API相关配置
type APIConfig struct {
	CoinCapKey string `mapstructure:"coincap_key"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret    string `mapstructure:"secret"`
	ExpiresIn int    `mapstructure:"expires_in"` // 小时
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// globalConfig 全局配置实例
var globalConfig *Config

// InitConfig 初始化配置
func InitConfig() *Config {
	v := viper.New()
	
	// 设置配置文件名和路径
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("/etc/fx/")
	
	// 设置环境变量前缀
	v.SetEnvPrefix("FX")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	
	// 设置默认值
	setDefaults(v)
	
	// 读取配置文件（如果存在）
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("配置文件未找到，使用环境变量和默认值")
		} else {
			log.Printf("读取配置文件失败: %v", err)
		}
	} else {
		log.Printf("使用配置文件: %s", v.ConfigFileUsed())
	}
	
	// 解析到结构体
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("解析配置失败: %v", err)
	}
	
	// 验证必填配置
	if err := validateConfig(&cfg); err != nil {
		log.Fatalf("配置验证失败: %v", err)
	}
	
	globalConfig = &cfg
	return &cfg
}

// setDefaults 设置默认值
func setDefaults(v *viper.Viper) {
	// 数据库默认值
	v.SetDefault("database.host", "172.23.112.1")
	v.SetDefault("database.port", 3306)
	v.SetDefault("database.user", "admin")
	v.SetDefault("database.password", "ctsi@Passw0rd")
	v.SetDefault("database.name", "insight_onchain")
	v.SetDefault("database.charset", "utf8mb4")
	
	// API默认值
	v.SetDefault("api.coincap_key", "")
	
	// 服务器默认值
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.mode", "release")
	
	// JWT默认值
	v.SetDefault("jwt.secret", "your-secret-key-change-in-production")
	v.SetDefault("jwt.expires_in", 24)
	
	// 日志默认值
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "text")
}

// validateConfig 验证配置
func validateConfig(cfg *Config) error {
	// 验证数据库配置
	if cfg.Database.Host == "" {
		return fmt.Errorf("数据库主机不能为空")
	}
	if cfg.Database.Port <= 0 {
		return fmt.Errorf("数据库端口无效")
	}
	if cfg.Database.User == "" {
		return fmt.Errorf("数据库用户名不能为空")
	}
	if cfg.Database.Name == "" {
		return fmt.Errorf("数据库名不能为空")
	}
	
	// 验证JWT配置
	if cfg.JWT.Secret == "" || cfg.JWT.Secret == "your-secret-key-change-in-production" {
		log.Println("警告: 使用默认JWT密钥，请在生产环境中修改")
	}
	
	// 验证API Key
	if cfg.API.CoinCapKey == "" {
		log.Println("警告: CoinCap API Key 未配置，价格查询功能可能受限")
	}
	
	return nil
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	if globalConfig == nil {
		return InitConfig()
	}
	return globalConfig
}

// GetDSN 获取数据库连接字符串
func (c *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=UTC",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
		c.Database.Charset,
	)
}

// IsDevelopment 是否为开发环境
func (c *Config) IsDevelopment() bool {
	return c.Server.Mode == "debug"
}

// IsProduction 是否为生产环境
func (c *Config) IsProduction() bool {
	return c.Server.Mode == "release"
}
