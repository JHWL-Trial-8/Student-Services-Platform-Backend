package userapi

import (
	"net/http"
	"strconv"

	usersvc "student-services-platform-backend/app/services/user"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *usersvc.Service
}

func New(s *usersvc.Service) *Handler {
	return &Handler{svc: s}
}

func (h *Handler) GetMe(c *gin.Context) {
	idStr := c.GetString("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户 ID 未找到"})
		return
	}

	userinfo, err := h.svc.GetByID(idStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取用户失败",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, userinfo)
}

// 请求体：使用指针以识别"字段是否出现"
type updateMePayload struct {
	Email      *string `json:"email"`        // required by spec
	Name       *string `json:"name"`         // optional
	Phone      *string `json:"phone"`        // optional; "" => clear
	Dept       *string `json:"dept"`         // optional; "" => clear
	AllowEmail *bool   `json:"allow_email"`  // optional
}

func (h *Handler) UpdateMe(c *gin.Context) {
	idStr := c.GetString("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户 ID 未找到"})
		return
	}
	uid64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户 ID"})
		return
	}

	var req updateMePayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}
	// 要求 email 必填
	if req.Email == nil || *req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email 为必填字段"})
		return
	}

	out, err := h.svc.UpdateByID(uint(uid64), usersvc.UpdateFields{
		Email:      req.Email,
		Name:       req.Name,
		Phone:      req.Phone,
		Dept:       req.Dept,
		AllowEmail: req.AllowEmail,
	})
	if err != nil {
		switch e := err.(type) {
		case *usersvc.ErrEmailTaken:
			c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱已被占用", "details": gin.H{"email": e.Email}})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败", "details": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, out)
}