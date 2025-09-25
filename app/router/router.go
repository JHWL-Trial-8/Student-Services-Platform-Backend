package router

import (
	"student-services-platform-backend/app/controllers/AuthController"
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
		// update current user profile
		user.PUT("/me", midwares.JWTAuthMidware(cfg.JWT.SecretKey), UserController.UpdateMe)
	}
}
