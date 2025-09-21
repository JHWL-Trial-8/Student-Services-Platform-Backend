package main

import (
    "log"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"

    "student-services-platform-backend/internal/config"
    httpmw "student-services-platform-backend/internal/http"
    dbpkg "student-services-platform-backend/internal/db"
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
    // 关闭底层连接池
    sqlDB, err := database.DB()
    if err == nil {
        defer sqlDB.Close()
    }

    r := gin.New()
    r.Use(gin.Logger(), gin.Recovery(), httpmw.CORS(cfg.CORS))

    api := r.Group("/api/v1")
    {
        api.GET("/healthz", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{
                "ok": true,
                "ts": time.Now().UTC().Format(time.RFC3339),
            })
        })
    }

    log.Printf("listening on :%s (mode=%s)", cfg.Server.Port, gin.Mode())
    if err := r.Run(":" + cfg.Server.Port); err != nil {
        log.Fatal(err)
    }
}