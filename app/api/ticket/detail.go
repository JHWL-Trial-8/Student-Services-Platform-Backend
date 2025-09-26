package ticketapi

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

func (h *Handler) Detail(c *gin.Context) {
	uid, ok := h.currentUID(c)
	if !ok {
		return
	}
	tid, ok := h.paramTicketID(c)
	if !ok {
		return
	}

	detail, svcErr := h.svc.GetTicketDetail(uid, tid)
	if svcErr != nil {
		h.handleTicketSvcErr(c, svcErr, "查询失败")
		return
	}
	c.JSON(http.StatusOK, detail)
}