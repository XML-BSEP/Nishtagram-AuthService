package usecase

import (
	"auth-service/domain"
	"auth-service/helper"
	"auth-service/infrastructure/dto"
	"auth-service/infrastructure/tracer"
	"auth-service/repository"
	"context"
	logger "github.com/jelena-vlajkov/logger/logger"
)

type profileInfoUsecase struct {
	ProfileInfoRepository repository.ProfileInfoRepository
	RedisUsecase          RedisUsecase
	logger *logger.Logger
}


const (
	userNotFound             = "user not found"
	emailNotSent             = "email not sent"
	invalidCode              = "invalidCode"
	invalidPass              = "password can't be the same as last one"
	hashError                = "error while hashing pass"
	updateError              = "error while updating user"
	redisError               = "error while deleting redis key"
	passwordsError           = "enter same passwords"
	redisPassResetKeyPattern = "passwordResetRequest"
)

type ProfileInfoUsecase interface {
	GetProfileInfoByUsername(context context.Context, username string) (domain.ProfileInfo, error)
	GetProfileInfoByEmail(context context.Context, email string) (domain.ProfileInfo, error)
	Create(context context.Context, profileInfo *domain.ProfileInfo) error
	ExistsByUsernameOrEmail(context context.Context, username, email string) bool
	GetProfileInfoById(context context.Context, id string) (*domain.ProfileInfo, error)
	ResetPassword(ctx context.Context, dto dto.ResetPassDTO) string
	DeleteProfileInfo(ctx context.Context, username string) error
}

func NewProfileInfoUsecase(p repository.ProfileInfoRepository, r RedisUsecase, logger *logger.Logger) ProfileInfoUsecase {
	return &profileInfoUsecase{ProfileInfoRepository: p, RedisUsecase: r, logger: logger}
}

func (p *profileInfoUsecase) DeleteProfileInfo(ctx context.Context, username string) error {
	return p.ProfileInfoRepository.DeleteProfileInfo(ctx, username)
}
func (p *profileInfoUsecase) GetProfileInfoByUsername(context context.Context, username string) (domain.ProfileInfo, error) {
	p.logger.Logger.Infof("getting profile info by username for %v\n", username)
	span := tracer.StartSpanFromContext(context, "usecase/GetProfileInfoByUsername")
	defer span.Finish()

	ctx1 := tracer.ContextWithSpan(context, span)
	return p.ProfileInfoRepository.GetProfileInfoByUsername(ctx1, username)
}

func (p *profileInfoUsecase) Create(context context.Context, profileInfo *domain.ProfileInfo) error {
	p.logger.Logger.Infof("creating profile info for email %v\n", profileInfo.Email)
	return p.ProfileInfoRepository.Create(context, profileInfo)
}

func (p *profileInfoUsecase) GetProfileInfoByEmail(context context.Context, email string) (domain.ProfileInfo, error) {
	p.logger.Logger.Infof("getting profile info by email %v\n", email)
	span := tracer.StartSpanFromContext(context, "usecase/GetProfileInfoByUsername")
	defer span.Finish()

	ctx1 := tracer.ContextWithSpan(context, span)
	return p.ProfileInfoRepository.GetProfileInfoByEmail(ctx1, email)
}

func (p *profileInfoUsecase) ResetPassword(ctx context.Context, dto dto.ResetPassDTO) string {
	p.logger.Logger.Infof("reseting password for user %v\n", dto.Email)

	if passwordCompare := dto.Password == dto.ConfirmedPassword; !passwordCompare {
		return passwordsError
	}

	exists := p.ExistsByUsernameOrEmail(ctx, "", dto.Email)
	if !exists {
		p.logger.Logger.Errorf("error while reseting password, error: user %v not found\n", dto.Email)
		return userNotFound
	}
	account, err := p.ProfileInfoRepository.GetProfileInfoByEmail(ctx, dto.Email)
	if err != nil {
		return userNotFound
	}
	key := redisPassResetKeyPattern + dto.Email
	codeValue, err := p.RedisUsecase.GetValueByKey(ctx, key)
	if err != nil {
		p.logger.Logger.Errorf("error while reseting password, error: email not sent to %v\n", dto.Email)
		return emailNotSent
	}

	err = VerifyPassword(ctx, dto.VerificationCode, string(codeValue))
	if err != nil {
		p.logger.Logger.Errorf("error while reseting password, error: %v\n", invalidCode)
		return invalidCode
	}

	err = VerifyPassword(ctx, dto.Password, account.Password)
	if err == nil {
		p.logger.Logger.Errorf("error while reseting password, error: %v\n", invalidPass)
		return invalidPass
	}

	newPass, err := helper.Hash(dto.Password)

	if err != nil {
		p.logger.Logger.Errorf("error while reseting password, error: %v\n", hashError)
		return hashError
	}

	account.Password = string(newPass)

	err = p.ProfileInfoRepository.Update(ctx, &account)

	if err != nil {
		p.logger.Logger.Errorf("error while reseting password, error: %v\n", updateError)
		return updateError
	}

	err = p.RedisUsecase.DeleteValueByKey(ctx, key)

	if err != nil {
		return redisError
	}

	return ""

}

func (p *profileInfoUsecase) ExistsByUsernameOrEmail(context context.Context, username, email string) bool {
	p.logger.Logger.Infof("checking if user exists")
	if err := p.ProfileInfoRepository.GetProfileInfoByUsernameOrEmail(context, username, email); err != nil {
		return false
	}
	return true
}

func (p *profileInfoUsecase) GetProfileInfoById(context context.Context, id string) (*domain.ProfileInfo, error) {
	p.logger.Logger.Infof("getting profile info by id for %v\n", id)
	span := tracer.StartSpanFromContext(context, "usecase/GetProfileInfoById")
	defer span.Finish()

	ctx1 := tracer.ContextWithSpan(context, span)
	return p.ProfileInfoRepository.GetProfileInfoById(ctx1, id)
}
