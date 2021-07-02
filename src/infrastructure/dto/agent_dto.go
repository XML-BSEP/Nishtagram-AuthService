package dto

type ConfirmAgentAccountDto struct {
	Confirm bool `json:"confirm"`
	Email string `json:"email"`
}

type AgentInformationDto struct {
	Name string `json:"name" validate:"required,name"`
	Surname string `json:"surname" validate:"required,surname"`
	Username string `json:"username" ,gorm:"unique" validate:"required,username" `
	Email string `json:"email" ,gorm:"unique" validate:"required,email"`
	Web string `json:"web"`

}