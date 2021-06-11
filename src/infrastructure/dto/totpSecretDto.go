package dto

type TotpSecretDto struct {
	Passcode string `json:"passcode"`
	UserId string `json:"user_id"`
}
