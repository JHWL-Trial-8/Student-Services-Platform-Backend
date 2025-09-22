package main

import (
    "log"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"

    "student-services-platform-backend/internal/config"
    httpmw "student-services-platform-backend/internal/http"
    dbpkg "student-services-platform-backend/internal/db"

    "student-services-platform-backend/app/services/auth"
    "student-services-platform-backend/app/router"
    "student-services-platform-backend/app/controllers/AuthController"
)

func main() {
    cfg := config.MustLoad()

    // 根据配置设置 Gin 模式
    switch cfg.Server.Mode {
    case gin.ReleaseMode:
        gin.SetMode(gin.ReleaseMode)
    case gin.TestMode:
        gin.SetMode(gin.TestMode)
    default:
        gin.SetMode(gin.DebugMode)
    }

    // 初始化数据库
    database := dbpkg.MustOpen(cfg.Database)
    if err := dbpkg.AutoMigrate(database); err != nil {
        log.Fatalf("db: 自动迁移失败: %v", err)
    }
    if sqlDB, err := database.DB(); err == nil {
        defer sqlDB.Close()
    }

    // 构建认证服务
    accessExp, _ := time.ParseDuration(cfg.JWT.AccessTokenExp)
    authSvc := auth.NewService(database, &auth.JWTConfig{
        SecretKey:      cfg.JWT.SecretKey,
        AccessTokenExp: accessExp,
        Issuer:         cfg.JWT.Issuer,
        Audience:       cfg.JWT.Audience,
    })
    AuthController.Svc = authSvc // 注入

    // 路由
    r := gin.New()
    r.Use(gin.Logger(), gin.Recovery(), httpmw.CORS(cfg.CORS))

    api := r.Group("/api/v1")
    {
        api.GET("/healthz", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{"ok": true, "ts": time.Now().UTC().Format(time.RFC3339)})
        })
        router.Init(api) // 在这里挂载 /auth/login 和 /auth/register
    }

    log.Printf("listening on :%s (mode=%s)", cfg.Server.Port, gin.Mode())
    if err := r.Run(":" + cfg.Server.Port); err != nil {
        log.Fatal(err)
    }
}