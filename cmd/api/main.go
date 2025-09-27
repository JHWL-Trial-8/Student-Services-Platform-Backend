package main

import (
    "log"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"

    // API Handlers
    authapi "student-services-platform-backend/app/api/auth"
    imagesapi "student-services-platform-backend/app/api/images"
    ticketapi "student-services-platform-backend/app/api/ticket"
    userapi "student-services-platform-backend/app/api/user"
    adminstatsapi "student-services-platform-backend/app/api/adminstats"
    cannedapi "student-services-platform-backend/app/api/canned"

    // Router
    "student-services-platform-backend/app/router"
    
    // Services
    authsvc "student-services-platform-backend/app/services/auth"
    imagessvc "student-services-platform-backend/app/services/images"
    ticketsvc "student-services-platform-backend/app/services/ticket"
    usersvc "student-services-platform-backend/app/services/user"
    adminstatssvc "student-services-platform-backend/app/services/adminstats"
    cannedsvc "student-services-platform-backend/app/services/canned"

    "student-services-platform-backend/internal/config"
    dbpkg "student-services-platform-backend/internal/db"
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

    accessExp, _ := time.ParseDuration(cfg.JWT.AccessTokenExp)
    authSvc := authsvc.NewService(database, &authsvc.JWTConfig{
        SecretKey:      cfg.JWT.SecretKey,
        AccessTokenExp: accessExp,
        Issuer:         cfg.JWT.Issuer,
        Audience:       cfg.JWT.Audience,
    })

    authH := authapi.New(authSvc)
    userH := userapi.New(usersvc.NewService(database))
    ticketH := ticketapi.New(ticketsvc.NewService(database))
    imagesH := imagesapi.New(imagessvc.NewService(database, store))
    adminStatsH := adminstatsapi.New(adminstatssvc.NewService(database))
    cannedH := cannedapi.New(cannedsvc.NewService(database))
    
    r := gin.New()
    r.Use(gin.Logger(), gin.Recovery(), httpserver.CORS(cfg.CORS))

    api := r.Group("/api/v1")
    {
        api.GET("/healthz", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{"ok": true, "ts": time.Now().UTC().Format(time.RFC3339)})
        })
        router.Init(api, cfg, database, authH, userH, ticketH, imagesH, adminStatsH, cannedH)
    }

    log.Printf("listening on :%s (mode=%s)", cfg.Server.Port, gin.Mode())
    if err := r.Run(":" + cfg.Server.Port); err != nil {
        log.Fatal(err)
    }
}