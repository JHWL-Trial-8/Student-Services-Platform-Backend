package authapi

import (
	"net/http"
	"student-services-platform-backend/internal/openapi"
	"github.com/gin-gonic/gin"
)

func (h *Handler) Register(c *gin.Context) {
	var req openapi.UserCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	u, err := h.svc.Register(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, u)
}