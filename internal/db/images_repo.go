package db

import (
	"errors"

	"gorm.io/gorm"
)

func GetImageBySHA(d *gorm.DB, sha string) (*Image, error) {
	var im Image
	if err := d.Where("sha256 = ?", sha).First(&im).Error; err != nil {
		return nil, err
	}
	return &im, nil
}

func GetImageByID(d *gorm.DB, id uint) (*Image, error) {
	var im Image
	if err := d.First(&im, id).Error; err != nil {
		return nil, err
	}
	return &im, nil
}

func CreateImage(d *gorm.DB, im *Image) error {
	return d.Create(im).Error
}

// IsImageAccessibleByUser 判断用户 uid 是否能通过工单关联访问该图片。
// 如果存在与图片关联的工单，并且满足以下任一条件，则授予访问权限：
// - ticket.user_id = uid  或
// - ticket.assigned_admin_id = uid
// (管理员可以查看所有内容；该检查由调用方完成。)
func IsImageAccessibleByUser(d *gorm.DB, imageID, uid uint) (bool, error) {
	var cnt int64
	err := d.
		Table("tickets AS t").
		Joins("JOIN ticket_images ti ON ti.ticket_id = t.id").
		Where("ti.image_id = ? AND (t.user_id = ? OR t.assigned_admin_id = ?)", imageID, uid, uid).
		Count(&cnt).Error
	if err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// DoesImageHaveAnyTicket 检查图片是否与任何工单关联（用于提供更具信息量的决策）。
func DoesImageHaveAnyTicket(d *gorm.DB, imageID uint) (bool, error) {
	var cnt int64
	err := d.Model(&TicketImage{}).Where("image_id = ?", imageID).Count(&cnt).Error
	if err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// IsAdminOrSuperAdmin 快速检查用户角色
func IsAdminOrSuperAdmin(d *gorm.DB, uid uint) (bool, error) {
	var u User
	if err := d.Select("id", "role").First(&u, uid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return u.Role == RoleAdmin || u.Role == RoleSuperAdmin, nil
}