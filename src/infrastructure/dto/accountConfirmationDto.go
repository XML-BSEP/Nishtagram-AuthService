package dto

type AccountConfirmationDto struct{
	Username string `json:"username"`
	Code string `json:"code"`
}
