package TicketController

import (
	"net/http"

	"student-services-platform-backend/internal/openapi"
	"github.com/gin-gonic/gin"
)

// POST /tickets/:id/rate
func Rate(c *gin.Context) {
	uid, ok := currentUID(c)
	if !ok {
		return
	}
	tid, ok := paramTicketID(c)
	if !ok {
		return
	}

	var req openapi.TicketsIdRatePostRequest
	if !mustBindJSON(c, &req) {
		return
	}

	r, svcErr := Svc.RateTicket(uid, tid, int(req.Stars), req.Comment)
	if svcErr != nil {
		handleTicketSvcErr(c, svcErr, "评分失败")
		return
	}
	c.JSON(http.StatusCreated, r)
}