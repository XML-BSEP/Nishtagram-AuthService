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
	GetProfileInfoByUsername(username string) (domain.ProfileInfo, error)
}

func NewProfileInfoRepository(conn *gorm.DB) ProfileInfoRepository {
	return &profileInfoRepository{Conn: conn}
}

func (p *profileInfoRepository) GetProfileInfoByEmail(email string) (domain.ProfileInfo, error) {
	profileInfo := domain.ProfileInfo{}
	err := p.Conn.Where("email = ?", email).Take(&profileInfo).Error

	return profileInfo, err
}

func (p *profileInfoRepository) GetProfileInfoByUsername(username string) (domain.ProfileInfo, error) {
	profileInfo := domain.ProfileInfo{}
	err := p.Conn.Where("username = ?", username).Take(&profileInfo).Error

	return profileInfo, err
}


