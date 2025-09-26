package ticketapi

import (
	"net/http"
	"strings"

	ticketsvc "student-services-platform-backend/app/services/ticket"
	"github.com/gin-gonic/gin"
)

func (h *Handler) List(c *gin.Context) {
	uid, ok := h.currentUID(c)
	if !ok {
		return
	}

	page, pageSize := h.parsePaging(c)

	status := strings.TrimSpace(c.Query("status"))
	category := strings.TrimSpace(c.Query("category"))

	isUrgent, ok := h.parseBoolQuery(c, "is_urgent")
	if !ok {
		return
	}
	assignedToMe, ok := h.parseBoolQuery(c, "assigned_to_me")
	if !ok {
		return
	}

	out, svcErr := h.svc.ListTickets(
		uid,
		ticketsvc.ListFilters{
			Status:       status,
			Category:     category,
			IsUrgent:     isUrgent,
			AssignedToMe: assignedToMe,
		},
		page, pageSize,
	)
	if svcErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败", "details": svcErr.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}