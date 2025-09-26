package ticketapi

import (
	"net/http"
	"student-services-platform-backend/internal/openapi"
	"github.com/gin-gonic/gin"
)

// GET /tickets/:id/messages
func (h *Handler) ListMessages(c *gin.Context) {
	uid, ok := h.currentUID(c)
	if !ok {
		return
	}
	tid, ok := h.paramTicketID(c)
	if !ok {
		return
	}
	page, pageSize := h.parsePaging(c)

	out, svcErr := h.svc.ListMessages(uid, tid, page, pageSize)
	if svcErr != nil {
		h.handleTicketSvcErr(c, svcErr, "查询失败")
		return
	}
	c.JSON(http.StatusOK, out)
}

// POST /tickets/:id/messages
func (h *Handler) PostMessage(c *gin.Context) {
	uid, ok := h.currentUID(c)
	if !ok {
		return
	}
	tid, ok := h.paramTicketID(c)
	if !ok {
		return
	}

	var req openapi.TicketsIdMessagesPostRequest
	if !h.mustBindJSON(c, &req) {
		return
	}

	msg, svcErr := h.svc.PostMessage(uid, tid, req.Body, req.IsInternalNote)
	if svcErr != nil {
		h.handleTicketSvcErr(c, svcErr, "创建消息失败")
		return
	}
	c.JSON(http.StatusCreated, msg)
}