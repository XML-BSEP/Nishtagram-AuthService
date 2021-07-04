package repository

import (
	"auth-service/domain"
	"auth-service/infrastructure/tracer"
	"context"
	logger "github.com/jelena-vlajkov/logger/logger"

	"gorm.io/gorm"
)

type profileInfoRepository struct {
	Conn *gorm.DB
	logger *logger.Logger
}


type ProfileInfoRepository interface {
	GetProfileInfoByEmail(context context.Context, email string) (domain.ProfileInfo, error)
	GetProfileInfoByUsername(context context.Context, username string) (domain.ProfileInfo, error)
	Create(context context.Context, profileInfo *domain.ProfileInfo) (*domain.ProfileInfo, error)
	GetProfileInfoByUsernameOrEmail(context context.Context, username, email string) error
	GetProfileInfoById(context context.Context, id string) (*domain.ProfileInfo, error)
	Update(context context.Context, profileInfo *domain.ProfileInfo) error
	DeleteProfileInfo(context context.Context, username string) error
}

func NewProfileInfoRepository(conn *gorm.DB, logger *logger.Logger) ProfileInfoRepository {
	return &profileInfoRepository{Conn: conn, logger: logger}
}

func (p *profileInfoRepository) DeleteProfileInfo(context context.Context, username string) error {
	/*profile, err := p.GetProfileInfoByUsername(context, username)
	if err != nil {
		return err
	}*/

	if err := p.Conn.Where("username = ?", username).Delete(&domain.ProfileInfo{}).Error; err != nil {
		return err
	}
	return nil
}

func (p *profileInfoRepository) GetProfileInfoByEmail(context context.Context, email string) (domain.ProfileInfo, error) {
	profileInfo := domain.ProfileInfo{}
	err := p.Conn.Joins("Role").Take(&profileInfo, "email = ?", email).Error
	if err != nil {
		p.logger.Logger.Errorf("error while getting profile info by email %v, error: %v\n", email, err)
	}
	return profileInfo, err
}

func (p *profileInfoRepository) GetProfileInfoByUsername(context context.Context, username string) (domain.ProfileInfo, error) {
	span := tracer.StartSpanFromContext(context, "GetProfileInfoByUsername")
	defer span.Finish()

	profileInfo := domain.ProfileInfo{}
	err := p.Conn.Joins("Role").Take(&profileInfo, "username = ?", username).Error

	if err != nil {
		p.logger.Logger.Errorf("error while getting profile info by username %v, error: %v\n", username, err)
		tracer.LogError(span, err)
	}

	return profileInfo, err
}

func (p *profileInfoRepository) Create(context context.Context, profileInfo *domain.ProfileInfo) (*domain.ProfileInfo, error) {
	err := p.Conn.Create(profileInfo).Error
	if err != nil {
		p.logger.Logger.Errorf("error while creating profile info for email %v, error: %v\n", profileInfo.Email, err)
		return nil, err
	}
	return profileInfo, nil
}
func (p *profileInfoRepository) Update(context context.Context, profileInfo *domain.ProfileInfo) error {
	err := p.Conn.Save(profileInfo).Error
	if err != nil {
		p.logger.Logger.Errorf("error while updating profile info for id %v, error: %v\n", profileInfo.ID, err)
	}
	return err
}

func (p *profileInfoRepository) GetByUsernameOrEmail(context context.Context, username, email string) (domain.ProfileInfo, error) {
	var value domain.ProfileInfo
	err := p.Conn.Where("username = ? or email = ?", username, email).Take(&value).Error
	if err != nil {
		p.logger.Logger.Errorf("error while getting user by email or username, error: %v\n", err)
	}
	return value, err
}

func (p *profileInfoRepository) GetProfileInfoByUsernameOrEmail(context context.Context, username, email string) error {
	var value *domain.ProfileInfo
	err := p.Conn.Where("username = ? or email = ?", username, email).Take(&value).Error

	if err != nil {
		p.logger.Logger.Errorf("error while gettin profile info by username or email, error: %v\n", err)
	}
	return err

}

func (p *profileInfoRepository) GetProfileInfoById(context context.Context, id string) (*domain.ProfileInfo, error) {
	var value *domain.ProfileInfo
	err := p.Conn.Preload("Role").Take(&value, "id = ?", id).Error

	if err != nil {
		p.logger.Logger.Errorf("error while getting profile info by id %v, error: %v\n", id, err)
	}
	return value, err
}
