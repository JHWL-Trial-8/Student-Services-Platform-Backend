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