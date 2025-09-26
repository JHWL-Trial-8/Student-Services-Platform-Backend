package ticketapi

import (
	"net/http"
	"strconv"
	"strings"

	ticketsvc "student-services-platform-backend/app/services/ticket"
	"student-services-platform-backend/internal/openapi"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *ticketsvc.Service
}

func New(s *ticketsvc.Service) *Handler {
	return &Handler{svc: s}
}

// 取得当前登录用户 ID
func (h *Handler) currentUID(c *gin.Context) (uint, bool) {
	uidStr := c.GetString("id")
	if uidStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return 0, false
	}
	uid64, err := strconv.ParseUint(uidStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户 ID"})
		return 0, false
	}
	return uint(uid64), true
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
func (h *Handler) handleTicketSvcErr(c *gin.Context, err error, fallback string) bool {
	switch e := err.(type) {
	case *ticketsvc.ErrForbidden:
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限"})
	case *ticketsvc.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "资源不存在"})
	case *ticketsvc.ErrValidation:
		c.JSON(http.StatusBadRequest, gin.H{"error": e.Error(), "details": e.Details})
	case *ticketsvc.ErrImageNotFound:
		c.JSON(http.StatusBadRequest, gin.H{"error": "部分图片不存在", "details": gin.H{"missing_image_ids": e.Missing}})
	case *ticketsvc.ErrAlreadyRated:
		c.JSON(http.StatusConflict, gin.H{"error": "该工单已评价"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": fallback, "details": err.Error()})
	}
	return true
}

// 统一 JSON 绑定
func (h *Handler) mustBindJSON(c *gin.Context, dst interface{}) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return false
	}
	return true
}

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

func (h *Handler) List(c *gin.Context) {
	uid, ok := h.currentUID(c)
	if !ok {
		return
	}

	page, pageSize := h.parsePaging(c)

	status := strings.TrimSpace(c.Query("status"))
	category := strings.TrimSpace(c.Query("category"))

	isUrgent, ok := h.parseBoolQuery(c, "is_urgent")
	if !ok {
		return
	}
	assignedToMe, ok := h.parseBoolQuery(c, "assigned_to_me")
	if !ok {
		return
	}

	out, svcErr := h.svc.ListTickets(
		uid,
		ticketsvc.ListFilters{
			Status:       status,
			Category:     category,
			IsUrgent:     isUrgent,
			AssignedToMe: assignedToMe,
		},
		page, pageSize,
	)
	if svcErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败", "details": svcErr.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

// GET /tickets/:id/messages
func (h *Handler) ListMessages(c *gin.Context) {
	uid, ok := h.currentUID(c)
	if !ok {
		return
	}
	tid, ok := h.paramTicketID(c)
	if !ok {
		return
	}
	page, pageSize := h.parsePaging(c)

	out, svcErr := h.svc.ListMessages(uid, tid, page, pageSize)
	if svcErr != nil {
		h.handleTicketSvcErr(c, svcErr, "查询失败")
		return
	}
	c.JSON(http.StatusOK, out)
}

// POST /tickets/:id/messages
func (h *Handler) PostMessage(c *gin.Context) {
	uid, ok := h.currentUID(c)
	if !ok {
		return
	}
	tid, ok := h.paramTicketID(c)
	if !ok {
		return
	}

	var req openapi.TicketsIdMessagesPostRequest
	if !h.mustBindJSON(c, &req) {
		return
	}

	msg, svcErr := h.svc.PostMessage(uid, tid, req.Body, req.IsInternalNote)
	if svcErr != nil {
		h.handleTicketSvcErr(c, svcErr, "创建消息失败")
		return
	}
	c.JSON(http.StatusCreated, msg)
}

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