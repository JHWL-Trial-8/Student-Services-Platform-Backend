package router

import (
	"student-services-platform-backend/app/controllers/AuthController"

	"github.com/gin-gonic/gin"
)

func Init(api *gin.RouterGroup) {
	auth := api.Group("/auth")
	{
		auth.POST("/login", AuthController.AuthByPassword)
		auth.POST("/register", AuthController.RegisterByPassword)
	}
}
