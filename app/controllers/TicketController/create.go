package TicketController

import (
	"net/http"
	"strconv"
	"strings"

	ticketsvc "student-services-platform-backend/app/services/ticket"
	"student-services-platform-backend/internal/openapi"

	"github.com/gin-gonic/gin"
)

var Svc *ticketsvc.Service

func Create(c *gin.Context) {
	uidStr := c.GetString("id")
	if uidStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}
	uid64, err := strconv.ParseUint(uidStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户 ID"})
		return
	}

	var req openapi.TicketCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 轻度清洗（避免前后空格导致校验不一致）
	req.Title = strings.TrimSpace(req.Title)
	req.Content = strings.TrimSpace(req.Content)
	req.Category = strings.TrimSpace(req.Category)

	out, err := Svc.CreateTicket(uint(uid64), req)
	if err != nil {
		switch e := err.(type) {
		case *ticketsvc.ErrValidation:
			c.JSON(http.StatusBadRequest, gin.H{"error": e.Error(), "details": e.Details})
			return
		case *ticketsvc.ErrImageNotFound:
			c.JSON(http.StatusBadRequest, gin.H{"error": "部分图片不存在", "details": gin.H{"missing_image_ids": e.Missing}})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建工单失败", "details": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, out)
}