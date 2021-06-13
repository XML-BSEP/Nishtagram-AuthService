package dto

type RefreshTokenDto struct{
	TokenUuid string `json:"token_uuid"`
	Token string `json:"token"`
}
