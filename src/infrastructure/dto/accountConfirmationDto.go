package dto

type AccountConfirmationDto struct{
	Code string `json:"code"`
	Email string `json:"email"`
	Username string `json:"username"`
}
