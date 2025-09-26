package router

import (
	"student-services-platform-backend/app/controllers/AuthController"
	"student-services-platform-backend/app/controllers/TicketController"
	"student-services-platform-backend/app/controllers/UserController"
	"student-services-platform-backend/app/midwares"
	"student-services-platform-backend/internal/config"

	"github.com/gin-gonic/gin"
)

func Init(api *gin.RouterGroup, cfg *config.Config) {

	auth := api.Group("/auth")
	{
		auth.POST("/login", AuthController.AuthByPassword)
		auth.POST("/register", AuthController.RegisterByPassword)
	}

	user := api.Group("/users")
	{
		user.GET("/me", midwares.JWTAuthMidware(cfg.JWT.SecretKey), UserController.GetUserInform)
		user.PUT("/me", midwares.JWTAuthMidware(cfg.JWT.SecretKey), UserController.UpdateMe)
	}

	tickets := api.Group("/tickets", midwares.JWTAuthMidware(cfg.JWT.SecretKey))
	{
		// POST /tickets
		tickets.POST("", TicketController.Create)

		// 列表 & 详情
		tickets.GET("", TicketController.List)
		tickets.GET("/:id", TicketController.Detail)

		// 消息流
		tickets.GET("/:id/messages", TicketController.ListMessages)
		tickets.POST("/:id/messages", TicketController.PostMessage)

		// 评分
		tickets.POST("/:id/rate", TicketController.Rate)
	}
}
