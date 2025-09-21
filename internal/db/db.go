package db

import (
    "log"
    "time"

    "student-services-platform-backend/internal/config"

    "gorm.io/driver/mysql"
    "gorm.io/driver/postgres"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

// 打开数据库连接并配置连接池
func Open(cfg config.DatabaseConfig) (*gorm.DB, error) {
    // 将配置中的日志级别映射到 GORM
    var lvl logger.LogLevel
    switch cfg.LogLevel {
    case "silent":
        lvl = logger.Silent
    case "error":
        lvl = logger.Error
    case "info":
        lvl = logger.Info
    default:
        lvl = logger.Warn // 默认 warn
    }

    gormCfg := &gorm.Config{
        Logger: logger.New(log.New(log.Writer(), "gorm: ", log.LstdFlags), logger.Config{
            SlowThreshold: 200 * time.Millisecond, // 慢查询阈值
            LogLevel:      lvl,
            Colorful:      true,
        }),
        // 迁移时不生成外键约束（仅用索引，避免因历史数据顺序导致迁移失败）
        DisableForeignKeyConstraintWhenMigrating: true,
    }

    // 选择驱动（默认 postgres）
    var dial gorm.Dialector
    switch cfg.Driver {
    case "mysql":
        dial = mysql.Open(cfg.DSN)
    case "sqlite":
        dial = sqlite.Open(cfg.DSN)
    default:
        dial = postgres.Open(cfg.DSN)
    }

    db, err := gorm.Open(dial, gormCfg)
    if err != nil {
        return nil, err
    }

    // 配置连接池（从 *gorm.DB 拿到底层 *sql.DB）
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }
    if cfg.MaxOpenConns > 0 {
        sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
    }
    if cfg.MaxIdleConns > 0 {
        sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
    }
    if cfg.ConnMaxLifetime != "" {
        if d, err := time.ParseDuration(cfg.ConnMaxLifetime); err == nil {
            sqlDB.SetConnMaxLifetime(d)
        }
    }

    return db, nil
}

// 打开数据库连接（失败直接退出）
func MustOpen(cfg config.DatabaseConfig) *gorm.DB {
    db, err := Open(cfg)
    if err != nil {
        log.Fatalf("db: 打开数据库失败: %v", err)
    }
    return db
}

// 自动迁移表结构（仅表结构）
func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &User{},
        &Ticket{},
        &TicketMessage{},
        &Image{},
        &Rating{},
        &AuditLog{},
        &SpamFlag{},
        &CannedReply{},
        &TicketImage{},
    )
}