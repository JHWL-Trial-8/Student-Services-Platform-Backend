package router

import (
	authapi "student-services-platform-backend/app/api/auth"
	imagesapi "student-services-platform-backend/app/api/images"
	ticketapi "student-services-platform-backend/app/api/ticket"
	userapi "student-services-platform-backend/app/api/user"

	adminstatsapi "student-services-platform-backend/app/api/adminstats"
	cannedapi "student-services-platform-backend/app/api/canned"

	"student-services-platform-backend/app/middleware"
	"student-services-platform-backend/internal/config"
	dbpkg "student-services-platform-backend/internal/db"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Init(
	api *gin.RouterGroup,
	cfg *config.Config,
	database *gorm.DB,
	authH *authapi.Handler,
	userH *userapi.Handler,
	ticketH *ticketapi.Handler,
	imagesH *imagesapi.Handler,
	adminStatsH *adminstatsapi.Handler,
	cannedH *cannedapi.Handler,
) {
	authRG := api.Group("/auth")
	{
		authRG.POST("/login", authH.Login)
		authRG.POST("/register", authH.Register)
	}

	userRG := api.Group("/users")
	{
		userRG.GET("/me", middleware.JWTAuth(cfg.JWT.SecretKey), userH.GetMe)
		userRG.PUT("/me", middleware.JWTAuth(cfg.JWT.SecretKey), userH.UpdateMe)
	}

	// 图片端点（需要认证）
	imagesRG := api.Group("/images", middleware.JWTAuth(cfg.JWT.SecretKey))
	{
		imagesRG.POST("", imagesH.Upload)
		imagesRG.GET("/:id", imagesH.Download)
	}

	ticketsRG := api.Group("/tickets", middleware.JWTAuth(cfg.JWT.SecretKey))
	{
		// 学生/管理员共有
		ticketsRG.POST("", ticketH.Create)
		ticketsRG.GET("", ticketH.List)
		ticketsRG.GET("/:id", ticketH.Detail)
		ticketsRG.GET("/:id/messages", ticketH.ListMessages)
		ticketsRG.POST("/:id/messages", ticketH.PostMessage)
		ticketsRG.POST("/:id/rate", ticketH.Rate)

		// 管理员工作流
		adminOnly := middleware.RequireRole(database, dbpkg.RoleAdmin, dbpkg.RoleSuperAdmin)
		superAdminOnly := middleware.RequireRole(database, dbpkg.RoleSuperAdmin)

		ticketsRG.POST("/:id/claim", adminOnly, ticketH.Claim)
		ticketsRG.POST("/:id/unclaim", adminOnly, ticketH.Unclaim)
		ticketsRG.POST("/:id/resolve", adminOnly, ticketH.Resolve)
		ticketsRG.POST("/:id/close", adminOnly, ticketH.Close)

		// 垃圾标记 & 审核
		ticketsRG.POST("/:id/spam-flag", adminOnly, ticketH.SpamFlag)
		ticketsRG.POST("/:id/spam-review", superAdminOnly, ticketH.SpamReview)
	}

	// 管理员：统计（仅限超级管理员）
	adminRG := api.Group("/admin",
		middleware.JWTAuth(cfg.JWT.SecretKey),
		middleware.RequireRole(database, dbpkg.RoleSuperAdmin),
	)
	{
		adminRG.GET("/stats", adminStatsH.Get)
	}

	// 管理员：常用回复（管理员 + 超级管理员）
	cannedRG := api.Group("/canned-replies",
		middleware.JWTAuth(cfg.JWT.SecretKey),
		middleware.RequireRole(database, dbpkg.RoleAdmin, dbpkg.RoleSuperAdmin),
	)
	{
		cannedRG.GET("", cannedH.List)
		cannedRG.POST("", cannedH.Create)
		cannedRG.PUT("/:id", cannedH.Update)
		cannedRG.DELETE("/:id", cannedH.Delete)
	}
}
