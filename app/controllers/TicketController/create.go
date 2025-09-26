package TicketController

import (
	"net/http"

	"student-services-platform-backend/internal/openapi"

	"github.com/gin-gonic/gin"
)

func Create(c *gin.Context) {
	uid, ok := currentUID(c)
	if !ok {
		return
	}

	var req openapi.TicketCreate
	if !mustBindJSON(c, &req) {
		return
	}

	out, err := Svc.CreateTicket(uid, req)
	if err != nil {
		handleTicketSvcErr(c, err, "创建工单失败")
		return
	}
	c.JSON(http.StatusCreated, out)
}