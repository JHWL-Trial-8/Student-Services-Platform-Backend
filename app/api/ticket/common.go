package ticketapi

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"student-services-platform-backend/app/contextkeys"
	ticketsvc "student-services-platform-backend/app/services/ticket"

	"github.com/gin-gonic/gin"
)

// currentUID 从 context 安全地获取用户 ID
func (h *Handler) currentUID(c *gin.Context) (uint, bool) {
	val, exists := c.Get(string(contextkeys.UserIDKey))
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return 0, false
	}

	uid, ok := val.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "上下文用户ID类型错误"})
		return 0, false
	}
	return uid, true
}

// 解析路径参数 :id
func (h *Handler) paramTicketID(c *gin.Context) (uint, bool) {
	idStr := c.Param("id")
	tid64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || tid64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的工单 ID"})
		return 0, false
	}
	return uint(tid64), true
}

// 分页
func (h *Handler) parsePaging(c *gin.Context) (int, int) {
	page := 1
	pageSize := 20
	if v := c.Query("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 {
			page = n
		}
	}
	if v := c.Query("page_size"); v != "" {
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

// 解析布尔查询参数；错误时直接返回 400
func (h *Handler) parseBoolQuery(c *gin.Context, key string) (*bool, bool) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return nil, true
	}
	b, err := strconv.ParseBool(raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": key + " 参数无效"})
		return nil, false
	}
	return &b, true
}

// 将 service 错误统一映射为 HTTP
func (h *Handler) handleTicketSvcErr(c *gin.Context, err error, fallback string) {
	switch e := err.(type) {
	case *ticketsvc.ErrForbidden:
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限", "details": e.Reason})
	case *ticketsvc.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "资源不存在"})
	case *ticketsvc.ErrValidation:
		c.JSON(http.StatusBadRequest, gin.H{"error": e.Message, "details": e.Details})
	case *ticketsvc.ErrImageNotFound:
		c.JSON(http.StatusBadRequest, gin.H{"error": "部分图片不存在", "details": gin.H{"missing_image_ids": e.Missing}})
	case *ticketsvc.ErrAlreadyRated:
		c.JSON(http.StatusConflict, gin.H{"error": "该工单已评价"})
	case *ticketsvc.ErrConflict:
		c.JSON(http.StatusConflict, gin.H{"error": e.Message})
	case *ticketsvc.ErrInvalidState:
		c.JSON(http.StatusBadRequest, gin.H{"error": e.Message}) // 状态机错误通常是客户端请求时机不对
	default:
		// 避免暴露过多内部错误细节
		log.Printf("Internal server error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fallback})
	}
}

// 统一 JSON 绑定
func (h *Handler) mustBindJSON(c *gin.Context, dst interface{}) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return false
	}
	return true
}