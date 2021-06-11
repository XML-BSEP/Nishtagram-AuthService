package domain

type User struct {
	ID string `json:"id"`
	Name string `json:"name" validate:"required,name"`
	Surname string `json:"surname" validate:"required,surname"`
	Username string `json:"username" ,gorm:"unique"validate:"required,username" `
	Password string `json:"password"`
	Email string `json:"email" ,gorm:"unique" validate:"required,email"`
	Address string `json:"address"`
	Phone string `json:"phone" validate:"required,phone"`
	Birthday string `json:"birthday"`
	Gender string `json:"gender"`
	Web string `json:"web"`
	Bio string `json:"bio"`
	Image string `json:"image"`
	ConfirmationCode string
}

