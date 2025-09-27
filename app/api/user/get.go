package userapi

import (
	"net/http"
	"strconv"
	"student-services-platform-backend/app/contextkeys"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetMe(c *gin.Context) {
	val, exists := c.Get(string(contextkeys.UserIDKey))
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户 ID 未找到"})
		return
	}
	uid, ok := val.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "上下文用户ID类型错误"})
		return
	}

	// service 层接收的是 string，所以这里转换一下
	userinfo, err := h.svc.GetByID(strconv.FormatUint(uint64(uid), 10))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取用户失败",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, userinfo)
}