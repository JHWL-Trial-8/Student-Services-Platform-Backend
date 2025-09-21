package routes

import "github.com/gin-gonic/gin"

func Init(r *gin.Engine) {
	auth := r.Group("/auth")
	{
		auth.POST("/login")

	}
}
