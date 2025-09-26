package TicketController

import (
	"net/http"
	"strconv"

	ticketsvc "student-services-platform-backend/app/services/ticket"

	"github.com/gin-gonic/gin"
)

func Detail(c *gin.Context) {
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

	detail, svcErr := Svc.GetTicketDetail(uint(uid64), uint(tid64))
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

	c.JSON(http.StatusOK, detail)
}