package usecase

import (
	"auth-service/domain"
	"auth-service/helper"
	"auth-service/infrastructure/dto"
	"auth-service/infrastructure/tracer"
	"auth-service/repository"
	"context"
)

type profileInfoUsecase struct {
	ProfileInfoRepository repository.ProfileInfoRepository
	RedisUsecase RedisUsecase
}
const (
	userNotFound = "user not found"
	emailNotSent = "email not sent"
	invalidCode = "invalidCode"
	invalidPass = "password can't be the same as last one"
	hashError = "error while hashing pass"
	updateError = "error while updating user"
	redisError = "error while deleting redis key"
	passwordsError = "enter same passwords"
	redisPassResetKeyPattern = "passwordResetRequest"

)
type ProfileInfoUsecase interface {
	GetProfileInfoByUsername(context context.Context, username string) (domain.ProfileInfo, error)
	GetProfileInfoByEmail(context context.Context, email string) (domain.ProfileInfo, error)

	Create(context context.Context, profileInfo *domain.ProfileInfo) error
	ExistsByUsernameOrEmail(context context.Context, username, email string) bool
	ResetPassword(ctx context.Context, dto dto.ResetPassDTO) string

}

func NewProfileInfoUsecase(p repository.ProfileInfoRepository, r RedisUsecase) ProfileInfoUsecase {
	return &profileInfoUsecase{ProfileInfoRepository: p, RedisUsecase: r}
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

func (p *profileInfoUsecase) GetProfileInfoByEmail(context context.Context, email string) (domain.ProfileInfo, error) {
	span := tracer.StartSpanFromContext(context, "usecase/GetProfileInfoByUsername")
	defer span.Finish()

	ctx1 := tracer.ContextWithSpan(context, span)
	return p.ProfileInfoRepository.GetProfileInfoByEmail(ctx1, email)
}

func (p *profileInfoUsecase) ResetPassword(ctx context.Context, dto dto.ResetPassDTO) string {

	if passwordCompare := dto.Password == dto.ConfirmedPassword; !passwordCompare {
		return passwordsError
	}

	exists := p.ExistsByUsernameOrEmail(ctx,"", dto.Email)
	if !exists {
		return userNotFound
	}
	account, err := p.ProfileInfoRepository.GetProfileInfoByEmail(ctx,dto.Email)
	if err != nil {
		return userNotFound
	}
	key :=redisPassResetKeyPattern + dto.Email
	codeValue, err := p.RedisUsecase.GetValueByKey(ctx,key)
	if err != nil {
		return emailNotSent
	}

	err = VerifyPassword(ctx,dto.VerificationCode, string(codeValue))
		if err != nil {
		return invalidCode
	}

	err = VerifyPassword(ctx, dto.Password, account.Password)
	if err == nil {
		return invalidPass
	}

	newPass, err := helper.Hash(dto.Password)

		if err != nil {
		return hashError
	}

	account.Password = string(newPass)

	err = p.ProfileInfoRepository.Update(ctx, &account)

	if err != nil {
		return updateError
	}

	err = p.RedisUsecase.DeleteValueByKey(ctx,key)

	if err != nil {
		return redisError
	}

		return ""

}

func (p *profileInfoUsecase) ExistsByUsernameOrEmail(context context.Context, username, email string) bool {
	if err := p.ProfileInfoRepository.GetProfileinfoByUsernameOrEmail(context, username, email); err != nil {
		return false
	}
	return true
}