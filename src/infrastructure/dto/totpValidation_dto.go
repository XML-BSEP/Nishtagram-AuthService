package dto

type TotpValidationDto struct {
	Passcode string `json:"passcode"`
	Refresh bool `json:"refresh"`
}
