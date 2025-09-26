package db

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 获取一批图片是否存在的 map[id]=>true
func GetExistingImageIDs(d *gorm.DB, ids []uint) (map[uint]bool, error) {
	out := make(map[uint]bool, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	var rows []uint
	if err := d.Model(&Image{}).Where("id IN ?", ids).Pluck("id", &rows).Error; err != nil {
		return nil, err
	}
	for _, id := range rows {
		out[id] = true
	}
	return out, nil
}

// 关联 Ticket <-> Images（忽略重复）
func LinkTicketImages(d *gorm.DB, ticketID uint, imageIDs []uint) error {
	if len(imageIDs) == 0 {
		return nil
	}
	rels := make([]TicketImage, 0, len(imageIDs))
	for _, iid := range imageIDs {
		rels = append(rels, TicketImage{TicketID: ticketID, ImageID: iid})
	}
	return d.Clauses(clause.OnConflict{DoNothing: true}).Create(&rels).Error
}

// 列出工单的 image_id 列表
func GetTicketImageIDs(d *gorm.DB, ticketID uint) ([]uint, error) {
	var ids []uint
	err := d.Model(&TicketImage{}).Where("ticket_id = ?", ticketID).Pluck("image_id", &ids).Error
	return ids, err
}