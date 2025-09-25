package db

import (
	"gorm.io/gorm"
)

func GetUserByEmail(d *gorm.DB, email string) (*User, error) {
	var u User
	if err := d.Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func CreateUser(d *gorm.DB, u *User) error {
	return d.Create(u).Error
}

// get by ID
func GetUserByID(d *gorm.DB, id uint) (*User, error) {
	var u User
	if err := d.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// check uniqueness excluding a user
func ExistsOtherUserWithEmail(d *gorm.DB, email string, excludeID uint) (bool, error) {
	var cnt int64
	if err := d.Model(&User{}).
		Where("email = ? AND id <> ?", email, excludeID).
		Count(&cnt).Error; err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// persist updates
func UpdateUser(d *gorm.DB, u *User) error {
	return d.Save(u).Error
}