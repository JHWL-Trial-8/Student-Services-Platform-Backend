package config

import (
    "log"
    "strings"
    "time"

    "github.com/spf13/viper"
)

type ServerConfig struct {
    // 例如："8080"
    Port string `mapstructure:"port"`
    // 例如："debug" | "release" | "test"
    Mode string `mapstructure:"mode"`
}

type CORSConfig struct {
    // AllowedOrigins: 使用 ["*"] 允许所有来源（如果 credentials=true，将回显请求的 Origin 而不是 "*"，以符合现代浏览器的要求）
    AllowedOrigins []string `mapstructure:"allowed_origins"`
    AllowedMethods []string `mapstructure:"allowed_methods"`
    AllowedHeaders []string `mapstructure:"allowed_headers"`
    AllowCredentials bool   `mapstructure:"allow_credentials"`
}

type Config struct {
    Server ServerConfig `mapstructure:"server"`
    CORS   CORSConfig   `mapstructure:"cors"`
}

func defaults(v *viper.Viper) {
    v.SetDefault("server.port", "8080")
    v.SetDefault("server.mode", "debug")

    v.SetDefault("cors.allowed_origins", []string{"*"})
    v.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
    v.SetDefault("cors.allowed_headers", []string{"Authorization", "Content-Type", "X-Requested-With"})
    v.SetDefault("cors.allow_credentials", true)
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