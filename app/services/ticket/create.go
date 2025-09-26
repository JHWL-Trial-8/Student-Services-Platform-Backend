package ticket

import (
	"fmt"
	"sort"
	"strings"
	"time"

	dbpkg "student-services-platform-backend/internal/db"
	"student-services-platform-backend/internal/openapi"

	"gorm.io/gorm"
)

// Service 封装工单领域逻辑
type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service { return &Service{db: db} }

// ---- Errors ----

type ErrValidation struct {
	Message string
	Details map[string]interface{}
}

func (e *ErrValidation) Error() string { return e.Message }

type ErrImageNotFound struct {
	Missing []uint
}

func (e *ErrImageNotFound) Error() string { return fmt.Sprintf("图片不存在: %v", e.Missing) }

// ---- Usecases ----

// CreateTicket 创建工单并可选关联图片
func (s *Service) CreateTicket(userID uint, in openapi.TicketCreate) (*openapi.Ticket, error) {
	// 输入校验（与 OpenAPI 对齐）
	details := map[string]interface{}{}
	if in.Title == "" {
		details["title"] = "必填"
	} else if len([]rune(in.Title)) > 120 {
		details["title"] = "长度不能超过 120"
	}
	if in.Content == "" {
		details["content"] = "必填"
	} else if len([]rune(in.Content)) > 4000 {
		details["content"] = "长度不能超过 4000"
	}
	if in.Category == "" {
		details["category"] = "必填"
	}
	if len(details) > 0 {
		return nil, &ErrValidation{Message: "字段校验失败", Details: details}
	}

	// 去重并转换 image_ids -> []uint
	uniqImg := make([]uint, 0, len(in.ImageIds))
	seen := map[uint]struct{}{}
	for _, id32 := range in.ImageIds {
		if id32 <= 0 {
			continue
		}
		u := uint(id32)
		if _, ok := seen[u]; !ok {
			seen[u] = struct{}{}
			uniqImg = append(uniqImg, u)
		}
	}
	sort.Slice(uniqImg, func(i, j int) bool { return uniqImg[i] < uniqImg[j] })

	var created *dbpkg.Ticket
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 校验图片是否存在
		if len(uniqImg) > 0 {
			existsMap, err := dbpkg.GetExistingImageIDs(tx, uniqImg)
			if err != nil {
				return err
			}
			var missing []uint
			for _, id := range uniqImg {
				if !existsMap[id] {
					missing = append(missing, id)
				}
			}
			if len(missing) > 0 {
				return &ErrImageNotFound{Missing: missing}
			}
		}

		// 创建 Ticket（默认 NEW）
		now := time.Now()
		t := &dbpkg.Ticket{
			UserID:      userID,
			Title:       strings.TrimSpace(in.Title),
			Content:     strings.TrimSpace(in.Content),
			Category:    strings.TrimSpace(in.Category),
			IsUrgent:    in.IsUrgent,
			IsAnonymous: in.IsAnonymous,
			Status:      dbpkg.TicketStatusNew,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := tx.Create(t).Error; err != nil {
			return err
		}

		// 关联图片（如有）
		if err := dbpkg.LinkTicketImages(tx, t.ID, uniqImg); err != nil {
			return err
		}

		created = t
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 查询最终 image_ids（保证返回的是数据库状态）
	imgIDs, err := dbpkg.GetTicketImageIDs(s.db, created.ID)
	if err != nil {
		return nil, err
	}
	imgIDs32 := make([]int32, 0, len(imgIDs))
	for _, id := range imgIDs {
		imgIDs32 = append(imgIDs32, int32(id))
	}

	out := &openapi.Ticket{
		Id:            int32(created.ID),
		UserId:        int32(created.UserID),
		Title:         created.Title,
		Content:       created.Content,
		Category:      created.Category,
		IsUrgent:      created.IsUrgent,
		IsAnonymous:   created.IsAnonymous,
		Status:        openapi.TicketStatus(created.Status),
		AssignedAdminId: nil,
		ClaimedAt:       nil,
		CreatedAt:       created.CreatedAt,
		UpdatedAt:       created.UpdatedAt,
		ImageIds:        imgIDs32,
	}
	return out, nil
}