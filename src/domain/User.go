package domain

type User struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Surname string `json:"surname"`
	Username string `json:"username" ,gorm:"unique" `
	Password string `json:"password"`
	Email string `json:"email" ,gorm:"unique"`
	Address string `json:"address"`
	Phone string `json:"phone"`
	Birthday string `json:"birthday"`
	Gender string `json:"gender"`
	Web string `json:"web"`
	Bio string `json:"bio"`
	Image string `json:"image"`
	ConfirmationCode string
}

