package imagesapi

import (
	"net/http"
	"strconv"

	"student-services-platform-backend/app/contextkeys"
	imagesvc "student-services-platform-backend/app/services/images"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *imagesvc.Service
}

func New(s *imagesvc.Service) *Handler {
	return &Handler{svc: s}
}

// POST /images
func (h *Handler) Upload(c *gin.Context) {
	// 认证
	val, ok := c.Get(string(contextkeys.UserIDKey))
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}
	_ = val.(uint) // 当前未使用；如果需要，可用于增加所有权等

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 file 字段（multipart/form-data）"})
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "打开文件失败", "details": err.Error()})
		return
	}
	defer f.Close()

	resp, err := h.svc.Upload(f)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// GET /images/:id
func (h *Handler) Download(c *gin.Context) {
	val, ok := c.Get(string(contextkeys.UserIDKey))
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}
	uid := val.(uint)

	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的图片ID"})
		return
	}

	file, size, mimeType, name, err := h.svc.ResolveForDownload(uid, uint(id64))
	if err != nil {
		// 避免泄露过多细节
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	defer file.Close()

	extraHeaders := map[string]string{
		"Content-Disposition": `inline; filename="` + name + `"`,
		// 安全加固：防止内容嗅探覆盖
		"X-Content-Type-Options": "nosniff",
	}
	c.DataFromReader(http.StatusOK, size, mimeType, file, extraHeaders)
}