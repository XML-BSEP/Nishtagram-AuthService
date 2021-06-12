package repository

import (
	"auth-service/domain"
	"auth-service/infrastructure/tracer"
	"context"
	logger "github.com/jelena-vlajkov/logger/logger"
	"gorm.io/gorm"
)

type totpRepository struct {
	Conn *gorm.DB
	logger *logger.Logger
}


type TotpRepository interface {
	GetSecretByProfileInfoId(context context.Context, profileInfoId string) (*string, error)
	Create(context context.Context, totpSecret domain.TotpSecret) error
	DeleteByProfileInfoId(context context.Context, profileInfoId string) error
}

func NewTotpRepository(conn *gorm.DB, logger *logger.Logger) TotpRepository {
	return &totpRepository{conn, logger}
}

func (t *totpRepository) GetSecretByProfileInfoId(context context.Context, profileInfoId string) (*string, error) {
	span := tracer.StartSpanFromContext(context, "repository/GetSecretByProfileInfoId")
	defer span.Finish()

	var totpSecret domain.TotpSecret
	if err := t.Conn.Joins("ProfileInfo").Where("profile_info_id = ?", profileInfoId).Take(&totpSecret).Error; err != nil {
		t.logger.Logger.Errorf("error while getting secret by profile info id %v, error: %v\n", profileInfoId, err)
		tracer.LogError(span, err)
		return nil, err
	}

	return &totpSecret.Secret, nil
}


func (t *totpRepository) Create(context context.Context, totpSecret domain.TotpSecret) error {
	span := tracer.StartSpanFromContext(context, "repository/Create")
	defer span.Finish()

	if err := t.Conn.Create(&totpSecret).Error; err != nil {
		t.logger.Logger.Errorf("error while creating secret for profile info id %v, error: %v\n", totpSecret.ProfileInfoId, err)
		tracer.LogError(span, err)
		return err
	}

	return nil
}

func (t *totpRepository) DeleteByProfileInfoId(context context.Context, profileInfoId string) error {
	span := tracer.StartSpanFromContext(context, "repository/DeleteByProfileInfoId")
	defer span.Finish()

	totpSecret := domain.TotpSecret{}
	if err := t.Conn.Where("profile_info_id = ?", profileInfoId).Delete(&totpSecret).Error; err != nil {
		t.logger.Logger.Errorf("error while deleting secret for profile info id %v, error: %v\n", totpSecret.ProfileInfoId, err)
		tracer.LogError(span, err)
		return err
	}

	return nil
}
