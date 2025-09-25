package UserController

import (
	"net/http"
	usersvc "student-services-platform-backend/app/services/user"

	"github.com/gin-gonic/gin"
)

// 注入
var Svc *usersvc.Service

func GetUserInform(c *gin.Context) {
	idStr := c.GetString("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户 ID 未找到"})
		return
	}

	userinfo, err := Svc.GetUserInfomationByID(idStr)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取用户失败",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, userinfo)
}
