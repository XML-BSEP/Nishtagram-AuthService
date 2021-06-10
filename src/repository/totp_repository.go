package repository

import (
	"auth-service/domain"
	"auth-service/infrastructure/tracer"
	"context"
	"gorm.io/gorm"
)

type totpRepository struct {
	Conn *gorm.DB
}

type TotpRepository interface {
	GetSecretByProfileInfoId(context context.Context, profileInfoId string) (*string, error)
	Create(context context.Context, totpSecret domain.TotpSecret) error
}

func NewTotpRepository(conn *gorm.DB) TotpRepository {
	return &totpRepository{conn}
}

func (t *totpRepository) GetSecretByProfileInfoId(context context.Context, profileInfoId string) (*string, error) {
	span := tracer.StartSpanFromContext(context, "repository/GetSecretByProfileInfoId")
	defer span.Finish()

	var secret string
	if err := t.Conn.Joins("profile_infos").Where("profile_info_id = ?", profileInfoId).Take(&secret).Error; err != nil {
		tracer.LogError(span, err)
		return nil, err
	}

	return &secret, nil
}


func (t *totpRepository) Create(context context.Context, totpSecret domain.TotpSecret) error {
	span := tracer.StartSpanFromContext(context, "repository/Create")
	defer span.Finish()

	if err := t.Conn.Create(&totpSecret).Error; err != nil {
		tracer.LogError(span, err)
		return err
	}

	return nil
}


