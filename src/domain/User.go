package domain

import (
	"time"
)

type User struct {
	ID string
	Username string `json:"username" ,gorm:"unique" `
	Password string `json:"password"`
	Email string `json:"email" ,gorm:"unique"`
	Address string `json:"address"`
	Phone string `json:"phone"`
	Birthday time.Time `json:"birthday"`
	Gender string `json:"gender"`
	Web string `json:"web"`
	Bio string `json:"bio"`
	Image string `json:"image"`
	ConfirmationCode string
}
