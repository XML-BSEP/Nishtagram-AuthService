package repository

import (
	"auth-service/src/domain"
	"gorm.io/gorm"
)

type profileInfoRepository struct {
	Conn *gorm.DB
}

type ProfileInfoRepository interface {
	GetProfileInfoByEmail(email string) (domain.ProfileInfo, error)
}

func NewProfileInfoRepository(conn *gorm.DB) ProfileInfoRepository {
	return &profileInfoRepository{Conn: conn}
}

func (p *profileInfoRepository) GetProfileInfoByEmail(email string) (domain.ProfileInfo, error) {
	user := domain.ProfileInfo{}
	err := p.Conn.Where("email = ?", email).Take(&user).Error

	return user, err
}



