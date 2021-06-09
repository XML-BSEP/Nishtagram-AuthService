package repository

import (
	"auth-service/domain"
	"auth-service/infrastructure/tracer"
	"context"
	"gorm.io/gorm"
)

type profileInfoRepository struct {
	Conn *gorm.DB
}


type ProfileInfoRepository interface {
	GetProfileInfoByEmail(context context.Context, email string) (domain.ProfileInfo, error)
	GetProfileInfoByUsername(context context.Context, username string) (domain.ProfileInfo, error)
	Create(context context.Context, profileInfo *domain.ProfileInfo) error
	GetProfileinfoByUsernameOrEmail(context context.Context, username, email string) error
}

func NewProfileInfoRepository(conn *gorm.DB) ProfileInfoRepository {
	return &profileInfoRepository{Conn: conn}
}

func (p *profileInfoRepository) GetProfileInfoByEmail(context context.Context, email string) (domain.ProfileInfo, error) {
	profileInfo := domain.ProfileInfo{}
	err := p.Conn.Joins("Role").Take(&profileInfo,"email = ?", email).Error

	return profileInfo, err
}

func (p *profileInfoRepository) GetProfileInfoByUsername(context context.Context, username string) (domain.ProfileInfo, error) {
	span := tracer.StartSpanFromContext(context, "GetProfileInfoByUsername")
	defer span.Finish()

	profileInfo := domain.ProfileInfo{}
	err := p.Conn.Joins("Role").Take(&profileInfo,"username = ?", username).Error

	if err != nil {
		tracer.LogError(span, err)
	}

	return profileInfo, err
}

func (p *profileInfoRepository) Create(context context.Context, profileInfo *domain.ProfileInfo) error {
	return p.Conn.Create(profileInfo).Error
}

func (p *profileInfoRepository) GetByUsernameOrEmail(context context.Context, username, email string) (domain.ProfileInfo, error) {
	var value domain.ProfileInfo
	err := p.Conn.Where("username = ? or email = ?", username, email).Take(&value).Error
	return value, err
}

func (p *profileInfoRepository) GetProfileinfoByUsernameOrEmail(context context.Context, username, email string)  error{
	var value *domain.ProfileInfo
	return p.Conn.Where("username = ? or email = ?", username, email).Take(&value).Error

}





