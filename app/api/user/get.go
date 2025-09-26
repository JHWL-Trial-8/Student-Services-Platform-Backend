package userapi

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetMe(c *gin.Context) {
	idStr := c.GetString("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户 ID 未找到"})
		return
	}

	userinfo, err := h.svc.GetByID(idStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取用户失败",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, userinfo)
}