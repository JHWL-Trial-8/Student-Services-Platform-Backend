package TicketController

import (
	"net/http"
	"strconv"

	ticketsvc "student-services-platform-backend/app/services/ticket"
	"student-services-platform-backend/internal/openapi"

	"github.com/gin-gonic/gin"
)

func Rate(c *gin.Context) {
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

	var req openapi.TicketsIdRatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	r, svcErr := Svc.RateTicket(uint(uid64), uint(tid64), int(req.Stars), req.Comment)
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
		case *ticketsvc.ErrAlreadyRated:
			c.JSON(http.StatusConflict, gin.H{"error": "该工单已评价"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "评分失败", "details": svcErr.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, r)
}