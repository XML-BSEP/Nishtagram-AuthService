package domain

import (
	"gorm.io/gorm"
	"time"
)

type ProfileInfo struct {
	ID string `json:"id" ,gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Username string `json:"username" ,gorm:"unique"`
	Email string `json:"email" ,gorm:"unique"`
	Password string `json:"password"`
	Role Role
	RoleId uint
}
