package usecase

import (
	"auth-service/src/domain"
	"auth-service/src/repository"
	"context"
)

type profileInfoUsecase struct {
	ProfileInfoRepository repository.ProfileInfoRepository
}

type ProfileInfoUsecase interface {
	GetProfileInfoByUsername(context context.Context, username string) (domain.ProfileInfo, error)
	Create(context context.Context, profileInfo *domain.ProfileInfo) error
	ExistsByUsernameOrEmail(context context.Context, username, email string) bool
}

func NewProfileInfoUsecase(p repository.ProfileInfoRepository) ProfileInfoUsecase {
	return &profileInfoUsecase{ProfileInfoRepository: p}
}

func (p *profileInfoUsecase) GetProfileInfoByUsername(context context.Context, username string) (domain.ProfileInfo, error) {
	return p.ProfileInfoRepository.GetProfileInfoByUsername(context, username)
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