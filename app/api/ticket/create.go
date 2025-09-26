package ticketapi

import (
	"net/http"
	"student-services-platform-backend/internal/openapi"
	"github.com/gin-gonic/gin"
)

func (h *Handler) Create(c *gin.Context) {
	uid, ok := h.currentUID(c)
	if !ok {
		return
	}

	var req openapi.TicketCreate
	if !h.mustBindJSON(c, &req) {
		return
	}

	out, err := h.svc.CreateTicket(uid, req)
	if err != nil {
		h.handleTicketSvcErr(c, err, "创建工单失败")
		return
	}
	c.JSON(http.StatusCreated, out)
}