package router

import (
	"student-services-platform-backend/app/controllers/AuthController"

	"github.com/gin-gonic/gin"
)

func Init(r *gin.Engine) {
	auth := r.Group("/auth")
	{
		auth.POST("/login", AuthController.AuthByPassword)
	}
}
