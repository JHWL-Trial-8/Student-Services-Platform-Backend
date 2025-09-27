package cannedapi

import (
	"net/http"
	"strconv"
	"strings"

	"student-services-platform-backend/app/contextkeys"
	"student-services-platform-backend/internal/openapi"

	"github.com/gin-gonic/gin"
)

func parsePage(c *gin.Context) (int, int) {
	page := 1
	pageSize := 20
	if v := strings.TrimSpace(c.Query("page")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 {
			page = n
		}
	}
	if v := strings.TrimSpace(c.Query("page_size")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			if n < 1 {
				n = 1
			}
			if n > 100 {
				n = 100
			}
			pageSize = n
		}
	}
	return page, pageSize
}

func (h *Handler) List(c *gin.Context) {
	val, ok := c.Get(string(contextkeys.UserIDKey))
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}
	uid := val.(uint)

	page, pageSize := parsePage(c)
	out, err := h.svc.List(uid, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) Create(c *gin.Context) {
	val, ok := c.Get(string(contextkeys.UserIDKey))
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}
	uid := val.(uint)

	var req openapi.CannedReplyCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}
	cr, err := h.svc.Create(uid, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cr)
}

func (h *Handler) Update(c *gin.Context) {
	val, ok := c.Get(string(contextkeys.UserIDKey))
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}
	uid := val.(uint)

	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req openapi.CannedReplyUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	cr, svcErr := h.svc.Update(uid, uint(id64), req)
	if svcErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": svcErr.Error()})
		return
	}
	c.JSON(http.StatusOK, cr)
}

func (h *Handler) Delete(c *gin.Context) {
	val, ok := c.Get(string(contextkeys.UserIDKey))
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}
	uid := val.(uint)

	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	if err := h.svc.Delete(uid, uint(id64)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}