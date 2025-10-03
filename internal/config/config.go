package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

// EmailConfig 邮件服务配置
type EmailConfig struct {
	SMTPHost      string `mapstructure:"smtp_host"`
	SMTPPort      int    `mapstructure:"smtp_port"`
	SMTPUsername  string `mapstructure:"smtp_username"`
	SMTPPassword  string `mapstructure:"smtp_password"`
	FromEmail     string `mapstructure:"from_email"`
	FromName      string `mapstructure:"from_name"`
	TLSEnabled    bool   `mapstructure:"tls_enabled"`
	TemplatesPath string `mapstructure:"templates_path"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// WorkerConfig Worker配置
type WorkerConfig struct {
	Concurrency int            `mapstructure:"concurrency"`
	Queues      map[string]int `mapstructure:"queues"`
	RetryDelay  string         `mapstructure:"retry_delay"`
	MaxRetry    int            `mapstructure:"max_retry"`
}

// FrontendConfig 前端配置
type FrontendConfig struct {
	BaseURL string `mapstructure:"base_url"`
}

type ServerConfig struct {
	// "8080"
	Port string `mapstructure:"port"`
	// "debug" | "release" | "test"
	Mode string `mapstructure:"mode"`
}

type CORSConfig struct {
	// AllowedOrigins: 使用 ["*"] 允许所有来源（如果 credentials=true，将回显请求的 Origin 而不是 "*"，以符合现代浏览器的要求）
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
}

// 数据库配置
type DatabaseConfig struct {
	// driver: postgres | mysql | sqlite
	Driver string `mapstructure:"driver"`
	// DSN
	DSN string `mapstructure:"dsn"`

	// 连接池
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime string `mapstructure:"conn_max_lifetime"` // 例如 "30m", "1h"

	// GORM 日志级别: silent | error | warn | info
	LogLevel string `mapstructure:"log_level"`
}

type JWTConfig struct {
	SecretKey      string `mapstructure:"secret_key"`
	AccessTokenExp string `mapstructure:"access_token_exp"` // e.g. "1h"
	Issuer         string `mapstructure:"issuer"`
	Audience       string `mapstructure:"audience"`
}

// 文件存储配置
type FileStoreConfig struct {
	// 所有存储对象的根目录（可以是相对路径或绝对路径）。
	// 把图片放在 <root>/images/<sha[:2]>/<sha> 目录下
	Root string `mapstructure:"root"`
}

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	CORS      CORSConfig      `mapstructure:"cors"`
	Database  DatabaseConfig  `mapstructure:"database"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Email     EmailConfig     `mapstructure:"email"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Worker    WorkerConfig    `mapstructure:"worker"`
	FileStore FileStoreConfig `mapstructure:"filestore"`
	Frontend  FrontendConfig  `mapstructure:"frontend"`
}

func defaults(v *viper.Viper) {
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.mode", "debug")

	v.SetDefault("cors.allowed_origins", []string{"*"})
	v.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	v.SetDefault("cors.allowed_headers", []string{"Authorization", "Content-Type", "X-Requested-With"})
	v.SetDefault("cors.allow_credentials", true)

	v.SetDefault("database.driver", "postgres")
	v.SetDefault("database.dsn", "postgres://postgres:postgres@localhost:5432/ssp?sslmode=disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", "30m")
	v.SetDefault("database.log_level", "warn")

	v.SetDefault("jwt.secret_key", "")
	v.SetDefault("jwt.access_token_exp", "6h")
	v.SetDefault("jwt.issuer", "ssp")
	v.SetDefault("jwt.audience", "ssp-web")

	v.SetDefault("filestore.root", "data")

	// 邮件配置默认值
	v.SetDefault("email.smtp_host", "")
	v.SetDefault("email.smtp_port", 587)
	v.SetDefault("email.smtp_username", "")
	v.SetDefault("email.smtp_password", "")
	v.SetDefault("email.from_email", "")
	v.SetDefault("email.from_name", "学生服务平台")
	v.SetDefault("email.tls_enabled", true)
	v.SetDefault("email.templates_path", "./templates/email")

	// Redis配置默认值
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	// Worker配置默认值
	v.SetDefault("worker.concurrency", 10)
	v.SetDefault("worker.queues.critical", 6)
	v.SetDefault("worker.queues.default", 3)
	v.SetDefault("worker.queues.low", 1)
	v.SetDefault("worker.retry_delay", "1s")
	v.SetDefault("worker.max_retry", 3)

	// 前端配置默认值
	v.SetDefault("frontend.base_url", "http://localhost:3000")
}

// Load 从以下位置返回一个配置（按优先级顺序）：
// 1) 默认值 -> 2) 配置文件 (config/config.yaml) -> 3) 环境变量 (SSP_*)
func Load() (*Config, error) {
	v := viper.New()

	// 默认值
	defaults(v)

	// 文件：在 ./config 和项目根目录中查找
	v.SetConfigName("config")
	v.AddConfigPath("./config")
	v.AddConfigPath(".") // root fallback
	// 允许 yaml/yml/json 格式
	v.SetConfigType("yaml")

	// 环境变量：SSP_SERVER_PORT, SSP_CORS_ALLOWED_ORIGINS 等
	v.SetEnvPrefix("SSP")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 尝试读取配置文件（如果存在）（如果缺失则不报错）
	if err := v.ReadInConfig(); err != nil {
		// 只有在找到但配置无效时才记录日志
		if _, notFound := err.(viper.ConfigFileNotFoundError); !notFound {
			log.Printf("config: using defaults/env because config file error: %v", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		log.Fatalf("config: failed to load: %v", err)
	}
	return cfg
}
