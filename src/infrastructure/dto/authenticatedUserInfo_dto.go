package dto

type AuthenticatedUserInfoDto struct {
	Id string `json:"id"`
	Role string `json:"role"`
	Token string `json:"token"`
}
