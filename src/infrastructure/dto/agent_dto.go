package dto

type ConfirmAgentAccountDto struct {
	Confirm bool `json:"confirm"`
	Email string `json:"email"`
}