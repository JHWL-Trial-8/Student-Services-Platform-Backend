package TicketController

import (
	"net/http"

	"student-services-platform-backend/internal/openapi"
	"github.com/gin-gonic/gin"
)

// GET /tickets/:id/messages
func ListMessages(c *gin.Context) {
	uid, ok := currentUID(c)
	if !ok {
		return
	}
	tid, ok := paramTicketID(c)
	if !ok {
		return
	}
	page, pageSize := parsePaging(c)

	out, svcErr := Svc.ListMessages(uid, tid, page, pageSize)
	if svcErr != nil {
		handleTicketSvcErr(c, svcErr, "查询失败")
		return
	}
	c.JSON(http.StatusOK, out)
}

// POST /tickets/:id/messages
func PostMessage(c *gin.Context) {
	uid, ok := currentUID(c)
	if !ok {
		return
	}
	tid, ok := paramTicketID(c)
	if !ok {
		return
	}

	var req openapi.TicketsIdMessagesPostRequest
	if !mustBindJSON(c, &req) {
		return
	}

	msg, svcErr := Svc.PostMessage(uid, tid, req.Body, req.IsInternalNote)
	if svcErr != nil {
		handleTicketSvcErr(c, svcErr, "创建消息失败")
		return
	}
	c.JSON(http.StatusCreated, msg)
}