package domain

import "gorm.io/gorm"

type ProfileInfo struct {
	gorm.Model
	Username string `json:"username" ,gorm:"unique"`
	Email string `json:"email" ,gorm:"unique"`
	Password string `json:"password"`
	Role Role
	RoleId uint
}
