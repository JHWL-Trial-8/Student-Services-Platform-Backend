package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	authapi "student-services-platform-backend/app/api/auth"
	ticketapi "student-services-platform-backend/app/api/ticket"
	userapi "student-services-platform-backend/app/api/user"
	"student-services-platform-backend/app/router"
	authsvc "student-services-platform-backend/app/services/auth"
	ticketsvc "student-services-platform-backend/app/services/ticket"
	usersvc "student-services-platform-backend/app/services/user"
	"student-services-platform-backend/internal/config"
	dbpkg "student-services-platform-backend/internal/db"
	"student-services-platform-backend/internal/email"
	httpserver "student-services-platform-backend/internal/httpserver"
)

func main() {
	cfg := config.MustLoad()
	gin.SetMode(cfg.Server.Mode)

	database := dbpkg.MustOpen(cfg.Database)
	if err := dbpkg.AutoMigrate(database); err != nil {
		log.Fatalf("db: 自动迁移失败: %v", err)
	}
	if sqlDB, err := database.DB(); err == nil {
		defer sqlDB.Close()
	}

	// 创建邮件服务（可选）
	var emailNotifier ticketsvc.EmailNotifier
	if cfg.Email.SMTPHost != "" {
		// 配置了邮件服务，创建邮件通知器
		emailConfig := &email.Config{
			SMTPHost:     cfg.Email.SMTPHost,
			SMTPPort:     cfg.Email.SMTPPort,
			SMTPUsername: cfg.Email.SMTPUsername,
			SMTPPassword: cfg.Email.SMTPPassword,
			FromEmail:    cfg.Email.FromEmail,
			FromName:     cfg.Email.FromName,
			TLSEnabled:   cfg.Email.TLSEnabled,
		}

		emailService := email.NewService(emailConfig)
		if err := emailService.ValidateConfig(); err != nil {
			log.Printf("邮件配置无效，禁用邮件通知: %v", err)
		} else {
			emailNotifier = email.NewNotifier(emailService)
			log.Println("邮件通知已启用")
		}
	} else {
		log.Println("未配置邮件服务，禁用邮件通知")
	}

	accessExp, _ := time.ParseDuration(cfg.JWT.AccessTokenExp)
	authSvc := authsvc.NewService(database, &authsvc.JWTConfig{
		SecretKey:      cfg.JWT.SecretKey,
		AccessTokenExp: accessExp,
		Issuer:         cfg.JWT.Issuer,
		Audience:       cfg.JWT.Audience,
	})

	authH := authapi.New(authSvc)
	userH := userapi.New(usersvc.NewService(database))

	// 根据是否有邮件通知器来创建工单服务
	var ticketSvc *ticketsvc.Service
	if emailNotifier != nil {
		ticketSvc = ticketsvc.NewServiceWithNotifier(database, emailNotifier)
	} else {
		ticketSvc = ticketsvc.NewService(database)
	}
	ticketH := ticketapi.New(ticketSvc)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), httpserver.CORS(cfg.CORS))

	api := r.Group("/api/v1")
	{
		api.GET("/healthz", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true, "ts": time.Now().UTC().Format(time.RFC3339)})
		})
		router.Init(api, cfg, database, authH, userH, ticketH)
	}

	log.Printf("listening on :%s (mode=%s)", cfg.Server.Port, gin.Mode())
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal(err)
	}
}
