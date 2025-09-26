package TicketController

import (
	"net/http"
	"strconv"
	"strings"

	ticketsvc "student-services-platform-backend/app/services/ticket"

	"github.com/gin-gonic/gin"
)

func parseBoolQuery(c *gin.Context, key string) (*bool, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return nil, nil
	}
	b, err := strconv.ParseBool(raw)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func List(c *gin.Context) {
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

	// filters
	status := strings.TrimSpace(c.Query("status"))
	category := strings.TrimSpace(c.Query("category"))
	isUrgent, err := parseBoolQuery(c, "is_urgent")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "is_urgent 参数无效"})
		return
	}
	assignedToMe, err := parseBoolQuery(c, "assigned_to_me")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "assigned_to_me 参数无效"})
		return
	}

	out, svcErr := Svc.ListTickets(uint(uid64), ticketsvc.ListFilters{
		Status:       status,
		Category:     category,
		IsUrgent:     isUrgent,
		AssignedToMe: assignedToMe,
	}, page, pageSize)

	if svcErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败", "details": svcErr.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}