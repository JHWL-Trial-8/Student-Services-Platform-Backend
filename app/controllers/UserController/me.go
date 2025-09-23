package UserController

import (
	"net/http"
	"os/user"
	"student-services-platform-backend/app/services/user"

	"github.com/gin-gonic/gin"
)

func GetUserInform(c *gin.Context) {
	idVal, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "用户 ID 未找到",
		})
		return
	}
	idStr, ok := idVal.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "用户 ID 类型错误",
		})
		return
	}
	userinfo, err := user.GetUserInfomationByID(idStr)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取用户失败",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, userinfo)
}
