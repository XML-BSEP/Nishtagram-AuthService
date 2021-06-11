package domain

//TODO stavi u malo
type UserRegistrationRequest struct {
	Email	string	`json:"email" validate:"required,email"`
	Name	string	`json:"name" validate:"required,name"`
	Surname	string	`json:"surname" validate:"required,surname"`
	Username     string	`json:"username" validate:"required,min=4,max=20"`
	Password   string	`json:"password" validate:"required"`
	ConfirmationCode	string `json:"code"`

}
