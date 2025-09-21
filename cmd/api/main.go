package main

import (
    "log"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"

    "student-services-platform-backend/internal/config"
    httpmw "student-services-platform-backend/internal/http"
)

func main() {
    cfg := config.MustLoad()

    // 从配置中获取Gin模式
    switch cfg.Server.Mode {
    case gin.ReleaseMode:
        gin.SetMode(gin.ReleaseMode)
    case gin.TestMode:
        gin.SetMode(gin.TestMode)
    default:
        gin.SetMode(gin.DebugMode)
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