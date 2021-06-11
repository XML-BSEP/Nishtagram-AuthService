package usecase

import (
	"auth-service/domain"
	"auth-service/infrastructure/tracer"
	"auth-service/repository"
	"context"
)

type profileInfoUsecase struct {
	ProfileInfoRepository repository.ProfileInfoRepository
}


type ProfileInfoUsecase interface {
	GetProfileInfoByUsername(context context.Context, username string) (domain.ProfileInfo, error)
	Create(context context.Context, profileInfo *domain.ProfileInfo) error
	ExistsByUsernameOrEmail(context context.Context, username, email string) bool
	GetProfileInfoById(context context.Context, id string) (*domain.ProfileInfo, error)
}

func NewProfileInfoUsecase(p repository.ProfileInfoRepository) ProfileInfoUsecase {
	return &profileInfoUsecase{ProfileInfoRepository: p}
}

func (p *profileInfoUsecase) GetProfileInfoByUsername(context context.Context, username string) (domain.ProfileInfo, error) {
	span := tracer.StartSpanFromContext(context, "usecase/GetProfileInfoByUsername")
	defer span.Finish()

	ctx1 := tracer.ContextWithSpan(context, span)
	return p.ProfileInfoRepository.GetProfileInfoByUsername(ctx1, username)
}

func (p *profileInfoUsecase) Create(context context.Context, profileInfo *domain.ProfileInfo) error {
	return p.ProfileInfoRepository.Create(context, profileInfo)
}


func (p *profileInfoUsecase) ExistsByUsernameOrEmail(context context.Context, username, email string) bool {
	if err := p.ProfileInfoRepository.GetProfileinfoByUsernameOrEmail(context, username, email); err != nil {
		return false
	}
	return true
}

func (p *profileInfoUsecase) GetProfileInfoById(context context.Context, id string) (*domain.ProfileInfo, error) {
	span := tracer.StartSpanFromContext(context, "usecase/GetProfileInfoById")
	defer span.Finish()

	ctx1 := tracer.ContextWithSpan(context, span)
	return p.ProfileInfoRepository.GetProfileInfoById(ctx1, id)
}