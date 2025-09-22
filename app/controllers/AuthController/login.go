package AuthController

import (
	"net/http"

	"student-services-platform-backend/app/services/auth"
	"student-services-platform-backend/internal/openapi"

	"github.com/gin-gonic/gin"
)

var Svc *auth.Service

func AuthByPassword(c *gin.Context) {
	var postRequest openapi.AuthLoginPostRequest
	if err := c.ShouldBindJSON(&postRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	jwtResponse, err := Svc.Login(postRequest.Email, postRequest.Password)
	if err != nil {
		switch e := err.(type) {
		case *auth.ErrUserNotFound:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在", "details": gin.H{"email": e.Email}})
		case *auth.ErrInvalidPassword:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误", "details": gin.H{"email": e.Email}})
		case *auth.ErrGenerateToken:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败", "details": e.Message})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "登录失败，请稍后再试"})
		}
		return
	}

	c.JSON(http.StatusOK, jwtResponse)
}
