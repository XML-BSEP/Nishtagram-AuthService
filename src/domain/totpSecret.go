package domain

import "gorm.io/gorm"

type TotpSecret struct {
	gorm.Model
	ProfileInfo ProfileInfo
	ProfileInfoId string
	Secret string
}
