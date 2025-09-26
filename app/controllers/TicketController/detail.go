package TicketController

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Detail(c *gin.Context) {
	uid, ok := currentUID(c)
	if !ok {
		return
	}
	tid, ok := paramTicketID(c)
	if !ok {
		return
	}

	detail, svcErr := Svc.GetTicketDetail(uid, tid)
	if svcErr != nil {
		handleTicketSvcErr(c, svcErr, "查询失败")
		return
	}
	c.JSON(http.StatusOK, detail)
}