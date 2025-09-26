package router

import (
	"student-services-platform-backend/app/api/auth"
	"student-services-platform-backend/app/api/ticket"
	"student-services-platform-backend/app/api/user"
	"student-services-platform-backend/app/middleware"
	"student-services-platform-backend/internal/config"

	"github.com/gin-gonic/gin"
)

func Init(api *gin.RouterGroup, cfg *config.Config,
	authH *authapi.Handler,
	userH *userapi.Handler,
	ticketH *ticketapi.Handler,
) {

	auth := api.Group("/auth")
	{
		auth.POST("/login", authH.Login)
		auth.POST("/register", authH.Register)
	}

	user := api.Group("/users")
	{
		user.GET("/me", middleware.JWTAuth(cfg.JWT.SecretKey), userH.GetMe)
		user.PUT("/me", middleware.JWTAuth(cfg.JWT.SecretKey), userH.UpdateMe)
	}

	tickets := api.Group("/tickets", middleware.JWTAuth(cfg.JWT.SecretKey))
	{
		// 创建
		tickets.POST("", ticketH.Create)

		// 列表 & 详情
		tickets.GET("", ticketH.List)
		tickets.GET("/:id", ticketH.Detail)

		// 消息流
		tickets.GET("/:id/messages", ticketH.ListMessages)
		tickets.POST("/:id/messages", ticketH.PostMessage)

		// 评分
		tickets.POST("/:id/rate", ticketH.Rate)
	}
}
