package images

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime"
	"net/http"
	"os"
	"strconv"
	"strings"

	dbpkg "student-services-platform-backend/internal/db"
	"student-services-platform-backend/internal/filestore"
	"student-services-platform-backend/internal/openapi"

	"golang.org/x/image/webp"
	"gorm.io/gorm"
)

type Service struct {
	db    *gorm.DB
	store filestore.Store
}

func NewService(db *gorm.DB, store filestore.Store) *Service {
	return &Service{db: db, store: store}
}

var allowedMIMEs = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/webp": {},
}

func isAllowed(m string) bool {
	_, ok := allowedMIMEs[strings.ToLower(m)]
	return ok
}

func extForMime(m string) string {
	switch strings.ToLower(m) {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		// 尽力从 mime 包获取
		if exts, _ := mime.ExtensionsByType(m); len(exts) > 0 {
			return exts[0]
		}
		return ""
	}
}

func (s *Service) Upload(reader io.Reader) (*openapi.ImagesPost201Response, error) {
	// 1) 读取前 512 字节用于内容嗅探
	head := make([]byte, 512)
	nHead, _ := io.ReadFull(reader, head)
	if nHead <= 0 {
		return nil, fmt.Errorf("空文件或读取失败")
	}
	head = head[:nHead]

	ctype := http.DetectContentType(head)
	if !isAllowed(ctype) {
		return nil, fmt.Errorf("不支持的图片类型: %s", ctype)
	}

	// 2) 在计算哈希的同时，将数据流式写入临时文件
	tmpFile, err := os.CreateTemp("", "ssp-img-*")
	if err != nil {
		return nil, fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	h := sha256.New()
	tee := io.TeeReader(io.MultiReader(bytes.NewReader(head), reader), h)
	written, err := io.Copy(tmpFile, tee)
	if err != nil {
		return nil, fmt.Errorf("写入临时文件失败: %w", err)
	}
	if written <= 0 {
		return nil, fmt.Errorf("空文件")
	}
	hashHex := hex.EncodeToString(h.Sum(nil))

	// 3) 如果文件已存在：复用并返回
	if exists, _, _ := s.store.Exists(hashHex); exists {
		if im, err := dbpkg.GetImageBySHA(s.db, hashHex); err == nil {
			return &openapi.ImagesPost201Response{
				ImageId: int32(im.ID),
				Sha256:  im.Sha256,
				Mime:    im.Mime,
				Width:   int32(im.Width),
				Height:  int32(im.Height),
			}, nil
		} // 否则继续重建数据库行
	}

	// 4) 解码配置以获取尺寸（验证文件确实可以解码）
	if _, err := tmpFile.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("临时文件seek失败: %w", err)
	}
	var cfg image.Config
	switch ctype {
	case "image/webp":
		// x/image/webp 只通过 webp 包暴露 DecodeConfig
		cfg, err = webp.DecodeConfig(tmpFile)
	default:
		cfg, _, err = image.DecodeConfig(tmpFile)
	}
	if err != nil {
		return nil, fmt.Errorf("图片解码失败（可能损坏或伪造）: %w", err)
	}
	width, height := cfg.Width, cfg.Height

	// 5) 持久化到文件存储
	objectKey, absPath, err := s.store.PutFromFile(hashHex, tmpFile.Name())
	if err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}
	// 我们移动了 tmpFile；防止延迟的 remove 操作删除了真实的对象
	_ = os.Remove(tmpFile.Name())

	// 6) 记录数据库行（通过 sha256 唯一索引去重）
	fi, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("读取文件信息失败: %w", err)
	}
	im := &dbpkg.Image{
		Sha256:    hashHex,
		Mime:      ctype,
		Size:      fi.Size(),
		Width:     width,
		Height:    height,
		ObjectKey: objectKey,
		RefCount:  1, // 逻辑上的初始引用计数
	}
	if err := dbpkg.CreateImage(s.db, im); err != nil {
		// 如果发生竞争（sha唯一），获取已存在的记录
		if ex, er2 := dbpkg.GetImageBySHA(s.db, hashHex); er2 == nil {
			im = ex
		} else {
			return nil, fmt.Errorf("写入数据库失败: %w", err)
		}
	}

	return &openapi.ImagesPost201Response{
		ImageId: int32(im.ID),
		Sha256:  im.Sha256,
		Mime:    im.Mime,
		Width:   int32(im.Width),
		Height:  int32(im.Height),
	}, nil
}

// ResolveForDownload 验证访问权限并返回 (文件, 大小, MIME类型, 下载文件名)
func (s *Service) ResolveForDownload(requestUID, imageID uint) (f *os.File, size int64, mimeType, name string, err error) {
	// 1) 加载图片记录
	im, err := dbpkg.GetImageByID(s.db, imageID)
	if err != nil {
		return nil, 0, "", "", fmt.Errorf("图片不存在")
	}

	// 2) 如果是管理员 -> 允许；否则通过工单关联检查
	isAdmin, err := dbpkg.IsAdminOrSuperAdmin(s.db, requestUID)
	if err != nil {
		return nil, 0, "", "", err
	}
	if !isAdmin {
		ok, err := dbpkg.IsImageAccessibleByUser(s.db, uint(im.ID), requestUID)
		if err != nil {
			return nil, 0, "", "", err
		}
		if !ok {
			// 如果图片没有任何关联，明确说明：只有管理员能看
			if linked, _ := dbpkg.DoesImageHaveAnyTicket(s.db, uint(im.ID)); !linked {
				return nil, 0, "", "", fmt.Errorf("无权限：该图片未关联任何工单")
			}
			return nil, 0, "", "", fmt.Errorf("无权限访问该图片")
		}
	}

	// 3) 从文件存储中打开
	file, _, err := s.store.Open(im.Sha256)
	if err != nil {
		return nil, 0, "", "", fmt.Errorf("文件已丢失或不可读")
	}
	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, 0, "", "", err
	}

	// 友好的内联文件名: image-<id>.<ext>
	name = "image-" + strconv.FormatUint(uint64(im.ID), 10) + extForMime(im.Mime)
	return file, stat.Size(), im.Mime, name, nil
}