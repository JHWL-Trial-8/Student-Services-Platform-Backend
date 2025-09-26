package ticketapi

import (
	"net/http"
	"student-services-platform-backend/internal/openapi"
	"github.com/gin-gonic/gin"
)

// POST /tickets/:id/rate
func (h *Handler) Rate(c *gin.Context) {
	uid, ok := h.currentUID(c)
	if !ok {
		return
	}
	tid, ok := h.paramTicketID(c)
	if !ok {
		return
	}

	var req openapi.TicketsIdRatePostRequest
	if !h.mustBindJSON(c, &req) {
		return
	}

	r, svcErr := h.svc.RateTicket(uid, tid, int(req.Stars), req.Comment)
	if svcErr != nil {
		h.handleTicketSvcErr(c, svcErr, "评分失败")
		return
	}
	c.JSON(http.StatusCreated, r)
}