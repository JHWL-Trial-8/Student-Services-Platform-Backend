package imageapi

import (
	"net/http"

	imagesvc "student-services-platform-backend/app/services/image"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *imagesvc.Service
}

func New(s *imagesvc.Service) *Handler {
	return &Handler{svc: s}
}

// Upload 处理图片上传请求
func (h *Handler) Upload(c *gin.Context) {
	// 1. 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "获取上传文件失败",
			"details": err.Error(),
		})
		return
	}
	defer file.Close()

	// 2. 调用服务层处理上传
	result, err := h.svc.UploadImage(file, header)
	if err != nil {
		h.handleImageSvcErr(c, err)
		return
	}

	// 3. 返回成功结果
	c.JSON(http.StatusCreated, result)
}

// handleImageSvcErr 统一处理图片服务错误
func (h *Handler) handleImageSvcErr(c *gin.Context, err error) {
	switch e := err.(type) {
	case *imagesvc.ErrInvalidFileType:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "不支持的文件类型",
			"details": gin.H{"mime_type": e.MimeType},
		})
	case *imagesvc.ErrFileTooLarge:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "文件过大",
			"details": gin.H{
				"size": e.Size,
				"max_size": e.MaxSize,
			},
		})
	case *imagesvc.ErrInvalidImage:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的图片文件",
			"details": gin.H{"reason": e.Reason},
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "上传失败",
			"details": err.Error(),
		})
	}
}