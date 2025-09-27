package image

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	dbpkg "student-services-platform-backend/internal/db"
	"student-services-platform-backend/internal/openapi"

	"gorm.io/gorm"
)

// Service 封装图片上传相关逻辑
type Service struct {
	db          *gorm.DB
	uploadDir   string // 图片存储目录
	maxFileSize int64  // 最大文件大小（字节）
}

func NewService(db *gorm.DB, uploadDir string) *Service {
	return &Service{
		db:          db,
		uploadDir:   uploadDir,
		maxFileSize: 10 * 1024 * 1024, // 10MB
	}
}

// ---- 错误类型 ----

type ErrInvalidFileType struct {
	MimeType string
}

func (e *ErrInvalidFileType) Error() string {
	return fmt.Sprintf("不支持的文件类型: %s", e.MimeType)
}

type ErrFileTooLarge struct {
	Size    int64
	MaxSize int64
}

func (e *ErrFileTooLarge) Error() string {
	return fmt.Sprintf("文件过大: %d bytes, 最大允许: %d bytes", e.Size, e.MaxSize)
}

type ErrInvalidImage struct {
	Reason string
}

func (e *ErrInvalidImage) Error() string {
	return fmt.Sprintf("无效的图片文件: %s", e.Reason)
}

// UploadImage 上传图片并返回图片信息
func (s *Service) UploadImage(file multipart.File, header *multipart.FileHeader) (*openapi.ImagesPost201Response, error) {
	// 1. 基础验证
	if err := s.validateFile(header); err != nil {
		return nil, err
	}

	// 2. 检测文件类型
	mimeType, err := s.detectMimeType(file)
	if err != nil {
		return nil, err
	}

	// 3. 验证是否为支持的图片格式
	if !s.isSupportedImageType(mimeType) {
		return nil, &ErrInvalidFileType{MimeType: mimeType}
	}

	// 4. 重置文件指针到开头
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("重置文件指针失败: %w", err)
	}

	// 5. 计算 SHA256 哈希
	hasher := sha256.New()
	size, err := io.Copy(hasher, file)
	if err != nil {
		return nil, fmt.Errorf("计算文件哈希失败: %w", err)
	}
	sha256Hash := fmt.Sprintf("%x", hasher.Sum(nil))

	// 6. 检查文件是否已存在（去重）
	var existingImage dbpkg.Image
	if err := s.db.Where("sha256 = ?", sha256Hash).First(&existingImage).Error; err == nil {
		// 文件已存在，返回现有记录
		return &openapi.ImagesPost201Response{
			ImageId: int32(existingImage.ID),
			Sha256:  existingImage.Sha256,
			Mime:    existingImage.Mime,
			Width:   int32(existingImage.Width),
			Height:  int32(existingImage.Height),
		}, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询现有图片失败: %w", err)
	}

	// 7. 重置文件指针准备读取图片尺寸
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("重置文件指针失败: %w", err)
	}

	// 8. 获取图片尺寸
	width, height, err := s.getImageDimensions(file)
	if err != nil {
		return nil, &ErrInvalidImage{Reason: err.Error()}
	}

	// 9. 重置文件指针准备保存文件
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("重置文件指针失败: %w", err)
	}

	// 10. 生成文件存储路径
	objectKey := s.generateObjectKey(sha256Hash, mimeType)
	filePath := filepath.Join(s.uploadDir, objectKey)

	// 11. 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	// 12. 保存文件到磁盘
	if err := s.saveFileToDisk(file, filePath); err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}

	// 13. 保存记录到数据库
	now := time.Now()
	imageRecord := &dbpkg.Image{
		Sha256:    sha256Hash,
		Mime:      mimeType,
		Size:      size,
		Width:     width,
		Height:    height,
		ObjectKey: objectKey,
		RefCount:  0, // 初始引用计数为0
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.db.Create(imageRecord).Error; err != nil {
		// 如果数据库保存失败，删除已保存的文件
		os.Remove(filePath)
		return nil, fmt.Errorf("保存图片记录失败: %w", err)
	}

	return &openapi.ImagesPost201Response{
		ImageId: int32(imageRecord.ID),
		Sha256:  imageRecord.Sha256,
		Mime:    imageRecord.Mime,
		Width:   int32(imageRecord.Width),
		Height:  int32(imageRecord.Height),
	}, nil
}

// ---- 私有辅助方法 ----

// validateFile 基础文件验证
func (s *Service) validateFile(header *multipart.FileHeader) error {
	if header.Size > s.maxFileSize {
		return &ErrFileTooLarge{
			Size:    header.Size,
			MaxSize: s.maxFileSize,
		}
	}
	return nil
}

// detectMimeType 检测文件 MIME 类型
func (s *Service) detectMimeType(file multipart.File) (string, error) {
	// 读取文件头部512字节来检测类型
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("读取文件头失败: %w", err)
	}

	mimeType := http.DetectContentType(buffer)
	return mimeType, nil
}

// isSupportedImageType 检查是否为支持的图片类型
func (s *Service) isSupportedImageType(mimeType string) bool {
	supportedTypes := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
	}

	for _, supportedType := range supportedTypes {
		if mimeType == supportedType {
			return true
		}
	}
	return false
}

// getImageDimensions 获取图片尺寸
func (s *Service) getImageDimensions(file multipart.File) (int, int, error) {
	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, fmt.Errorf("解析图片配置失败: %w", err)
	}
	return config.Width, config.Height, nil
}

// generateObjectKey 生成对象存储键名
func (s *Service) generateObjectKey(sha256Hash, mimeType string) string {
	// 根据 MIME 类型确定文件扩展名
	var ext string
	switch mimeType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/gif":
		ext = ".gif"
	default:
		ext = ".bin"
	}

	// 使用日期分层存储，避免单个目录文件过多
	now := time.Now()
	datePath := now.Format("2006/01/02")
	
	// 使用 SHA256 前缀作为文件名，便于去重和查找
	return fmt.Sprintf("%s/%s%s", datePath, sha256Hash[:16], ext)
}

// saveFileToDisk 保存文件到磁盘
func (s *Service) saveFileToDisk(file multipart.File, filePath string) error {
	dst, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	return err
}