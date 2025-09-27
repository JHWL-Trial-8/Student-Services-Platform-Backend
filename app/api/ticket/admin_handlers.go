package ticketapi

import (
	"net/http"

	"student-services-platform-backend/internal/openapi"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Claim(c *gin.Context) {
	uid, ok := h.currentUID(c)
	if !ok {
		return
	}
	tid, ok := h.paramTicketID(c)
	if !ok {
		return
	}
	if err := h.svc.ClaimTicket(c.Request.Context(), uid, tid); err != nil {
		h.handleTicketSvcErr(c, err, "接单失败")
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) Unclaim(c *gin.Context) {
	uid, ok := h.currentUID(c)
	if !ok {
		return
	}
	tid, ok := h.paramTicketID(c)
	if !ok {
		return
	}
	if err := h.svc.UnclaimTicket(c.Request.Context(), uid, tid); err != nil {
		h.handleTicketSvcErr(c, err, "撤销接单失败")
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) Resolve(c *gin.Context) {
	uid, ok := h.currentUID(c)
	if !ok {
		return
	}
	tid, ok := h.paramTicketID(c)
	if !ok {
		return
	}
	if err := h.svc.ResolveTicket(c.Request.Context(), uid, tid); err != nil {
		h.handleTicketSvcErr(c, err, "标记已处理失败")
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) Close(c *gin.Context) {
	uid, ok := h.currentUID(c)
	if !ok {
		return
	}
	tid, ok := h.paramTicketID(c)
	if !ok {
		return
	}
	if err := h.svc.CloseTicket(c.Request.Context(), uid, tid); err != nil {
		h.handleTicketSvcErr(c, err, "关闭工单失败")
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) SpamFlag(c *gin.Context) {
	uid, ok := h.currentUID(c)
	if !ok {
		return
	}
	tid, ok := h.paramTicketID(c)
	if !ok {
		return
	}
	var req openapi.TicketsIdSpamFlagPostRequest
	if !h.mustBindJSON(c, &req) {
		return
	}
	sf, err := h.svc.SpamFlag(c.Request.Context(), uid, tid, req.Reason)
	if err != nil {
		h.handleTicketSvcErr(c, err, "标记垃圾失败")
		return
	}
	c.JSON(http.StatusCreated, sf)
}

func (h *Handler) SpamReview(c *gin.Context) {
	uid, ok := h.currentUID(c)
	if !ok {
		return
	}
	tid, ok := h.paramTicketID(c)
	if !ok {
		return
	}
	var req openapi.TicketsIdSpamReviewPostRequest
	if !h.mustBindJSON(c, &req) {
		return
	}
	if err := h.svc.SpamReview(c.Request.Context(), uid, tid, req.Action); err != nil {
		h.handleTicketSvcErr(c, err, "垃圾审核失败")
		return
	}
	c.Status(http.StatusNoContent)
}