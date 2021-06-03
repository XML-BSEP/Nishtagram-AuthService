package usecase

import (
	"auth-service/src/domain"
	"auth-service/src/repository"
)

type profileInfoUsecase struct {
	ProfileInfoRepository repository.ProfileInfoRepository
}

type ProfileInfoUsecase interface {
	GetProfileInfoByUsername(username string) (domain.ProfileInfo, error)
}

func NewProfileInfoUsecase(p repository.ProfileInfoRepository) ProfileInfoUsecase {
	return &profileInfoUsecase{ProfileInfoRepository: p}
}

func (p *profileInfoUsecase) GetProfileInfoByUsername(username string) (domain.ProfileInfo, error) {
	return p.ProfileInfoRepository.GetProfileInfoByUsername(username)
}