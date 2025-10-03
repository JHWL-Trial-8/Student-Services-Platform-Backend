package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	// API Handlers
	adminstatsapi "student-services-platform-backend/app/api/adminstats"
	adminuserapi "student-services-platform-backend/app/api/adminuser"
	authapi "student-services-platform-backend/app/api/auth"
	cannedapi "student-services-platform-backend/app/api/canned"
	imagesapi "student-services-platform-backend/app/api/images"
	ticketapi "student-services-platform-backend/app/api/ticket"
	userapi "student-services-platform-backend/app/api/user"

	// Router
	"student-services-platform-backend/app/router"

	// Services
	adminstatssvc "student-services-platform-backend/app/services/adminstats"
	adminusersvc "student-services-platform-backend/app/services/adminuser"
	authsvc "student-services-platform-backend/app/services/auth"
	cannedsvc "student-services-platform-backend/app/services/canned"
	imagessvc "student-services-platform-backend/app/services/images"
	ticketsvc "student-services-platform-backend/app/services/ticket"
	usersvc "student-services-platform-backend/app/services/user"

	"student-services-platform-backend/internal/config"
	dbpkg "student-services-platform-backend/internal/db"
	"student-services-platform-backend/internal/email"
	"student-services-platform-backend/internal/filestore"
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

	// 确保文件存储目录存在
	store := filestore.NewLocal(cfg.FileStore.Root)
	if err := store.Ensure(); err != nil {
		log.Fatalf("filestore: 初始化失败: %v", err)
	}

	// 创建邮件服务（可选）
	var emailNotifier ticketsvc.EmailNotifier
	if cfg.Email.SMTPHost != "" {
		// 配置了邮件服务，创建邮件通知器
		emailConfig := &email.Config{
			SMTPHost:      cfg.Email.SMTPHost,
			SMTPPort:      cfg.Email.SMTPPort,
			SMTPUsername:  cfg.Email.SMTPUsername,
			SMTPPassword:  cfg.Email.SMTPPassword,
			FromEmail:     cfg.Email.FromEmail,
			FromName:      cfg.Email.FromName,
			TLSEnabled:    cfg.Email.TLSEnabled,
			TemplatesPath: cfg.Email.TemplatesPath,
		}

		emailService, err := email.NewService(emailConfig)
		if err != nil {
			log.Printf("创建邮件服务失败，禁用邮件通知: %v", err)
		} else if err := emailService.ValidateConfig(); err != nil {
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
	imagesH := imagesapi.New(imagessvc.NewService(database, store))
	adminStatsH := adminstatsapi.New(adminstatssvc.NewService(database))
	cannedH := cannedapi.New(cannedsvc.NewService(database))
	adminUserH := adminuserapi.New(adminusersvc.NewService(database))

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), httpserver.CORS(cfg.CORS))

	api := r.Group("/api/v1")
	{
		api.GET("/healthz", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true, "ts": time.Now().UTC().Format(time.RFC3339)})
		})
		router.Init(api, cfg, database, authH, userH, ticketH, imagesH, adminStatsH, cannedH, adminUserH)
	}

	log.Printf("listening on :%s (mode=%s)", cfg.Server.Port, gin.Mode())
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal(err)
	}
}
