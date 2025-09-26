package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"student-services-platform-backend/internal/config"
	dbpkg "student-services-platform-backend/internal/db"
	httpserver "student-services-platform-backend/internal/httpserver"

	"student-services-platform-backend/app/api/auth"
	"student-services-platform-backend/app/api/ticket"
	"student-services-platform-backend/app/api/user"
	"student-services-platform-backend/app/router"
	"student-services-platform-backend/app/services/auth"
	ticketsvc "student-services-platform-backend/app/services/ticket"
	usersvc "student-services-platform-backend/app/services/user"
)

func main() {
	cfg := config.MustLoad()

	// Gin 模式
	switch cfg.Server.Mode {
	case gin.ReleaseMode:
		gin.SetMode(gin.ReleaseMode)
	case gin.TestMode:
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	// 数据库
	database := dbpkg.MustOpen(cfg.Database)
	if err := dbpkg.AutoMigrate(database); err != nil {
		log.Fatalf("db: 自动迁移失败: %v", err)
	}
	if sqlDB, err := database.DB(); err == nil {
		defer sqlDB.Close()
	}

	// 认证服务
	accessExp, _ := time.ParseDuration(cfg.JWT.AccessTokenExp)
	authSvc := auth.NewService(database, &auth.JWTConfig{
		SecretKey:      cfg.JWT.SecretKey,
		AccessTokenExp: accessExp,
		Issuer:         cfg.JWT.Issuer,
		Audience:       cfg.JWT.Audience,
	})

	// 创建处理器
	authH := authapi.New(authSvc)
	userH := userapi.New(usersvc.NewService(database))
	ticketH := ticketapi.New(ticketsvc.NewService(database))

	// 路由
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), httpserver.CORS(cfg.CORS))

	api := r.Group("/api/v1")
	{
		api.GET("/healthz", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true, "ts": time.Now().UTC().Format(time.RFC3339)})
		})
		router.Init(api, cfg, authH, userH, ticketH)
	}

	log.Printf("listening on :%s (mode=%s)", cfg.Server.Port, gin.Mode())
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal(err)
	}
}
