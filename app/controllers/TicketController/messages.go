package TicketController

import (
	"net/http"
	"strconv"

	ticketsvc "student-services-platform-backend/app/services/ticket"
	"student-services-platform-backend/internal/openapi"

	"github.com/gin-gonic/gin"
)

func ListMessages(c *gin.Context) {
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

	idStr := c.Param("id")
	tid64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || tid64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的工单 ID"})
		return
	}

	page := 1
	pageSize := 20
	if v := c.Query("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 {
			page = n
		}
	}
	if v := c.Query("page_size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			if n < 1 {
				n = 1
			}
			if n > 100 {
				n = 100
			}
			pageSize = n
		}
	}

	out, svcErr := Svc.ListMessages(uint(uid64), uint(tid64), page, pageSize)
	if svcErr != nil {
		switch svcErr.(type) {
		case *ticketsvc.ErrForbidden:
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
			return
		case *ticketsvc.ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "资源不存在"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败", "details": svcErr.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, out)
}

func PostMessage(c *gin.Context) {
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

	idStr := c.Param("id")
	tid64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || tid64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的工单 ID"})
		return
	}

	var req openapi.TicketsIdMessagesPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	msg, svcErr := Svc.PostMessage(uint(uid64), uint(tid64), req.Body, req.IsInternalNote)
	if svcErr != nil {
		switch e := svcErr.(type) {
		case *ticketsvc.ErrForbidden:
			c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
			return
		case *ticketsvc.ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "资源不存在"})
			return
		case *ticketsvc.ErrValidation:
			c.JSON(http.StatusBadRequest, gin.H{"error": e.Error(), "details": e.Details})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建消息失败", "details": svcErr.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, msg)
}