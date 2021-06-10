package usecase

import (
	"auth-service/domain"
	"auth-service/infrastructure/tracer"
	"auth-service/repository"
	"context"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"image"
	"time"
)

const (
	totp_issuer = "Nishtagram"
	secret_prefix = "secret/"
)


type totpUsecase struct {
	TotpRepository repository.TotpRepository
	RedisUsecase RedisUsecase
}


type TotpUsecase interface {
	GenereateTotpSecret(context context.Context, user string) (*otp.Key, error)
	GetSecretString(context context.Context, key *otp.Key) string
	GetSecretImage(context context.Context, key *otp.Key, width, height int) (*image.Image, error)
	Validate(context context.Context, passcode, userId string) bool
	SaveSecretTemporarily(context context.Context, userId, secret string) error
	SaveSecret(context context.Context, userId string) error
}

func NewTotpUsecase(repository repository.TotpRepository, redisUsecase RedisUsecase, profileInfoUsecase ProfileInfoUsecase) TotpUsecase {
	return &totpUsecase{TotpRepository: repository, RedisUsecase: redisUsecase}
}

func (t *totpUsecase) GenereateTotpSecret(context context.Context, user string) (*otp.Key, error){
	span := tracer.StartSpanFromContext(context, "usecase/GenereateTotpSecret")
	defer span.Finish()



	options := totp.GenerateOpts{Issuer: totp_issuer, AccountName: user}

	key, err := totp.Generate(options)

	if err != nil {
		tracer.LogError(span, err)
		return nil, err
	}

	return key, nil
}

func (t *totpUsecase) GetSecretString(context context.Context, key *otp.Key) string {
	span := tracer.StartSpanFromContext(context, "usecase/GetSecretString")
	defer span.Finish()

	return key.Secret()
}

func (t *totpUsecase) GetSecretImage(context context.Context, key *otp.Key, width, height int) (*image.Image, error) {
	span := tracer.StartSpanFromContext(context, "usecase/GetSecretImage")
	defer span.Finish()

	img, err := key.Image(width, height)

	if err != nil {
		tracer.LogError(span, err)
		return nil, err
	}

	return &img, nil
}

func (t *totpUsecase) Validate(context context.Context, passcode, userId string) bool {
	span := tracer.StartSpanFromContext(context, "usecase/Validate")
	defer span.Finish()

	key := secret_prefix + userId
	ctx1 := tracer.ContextWithSpan(context, span)
	secret, err := t.RedisUsecase.GetValueByKey(ctx1, key)

	if err != nil {
		tracer.LogError(span, err)
		return false
	}

	return totp.Validate(passcode, string(secret))

}


func (t *totpUsecase) SaveSecretTemporarily(context context.Context, userId, secret string) error {
	span := tracer.StartSpanFromContext(context, "usecase/SaveSecretTemporarily")
	defer span.Finish()

	key := secret_prefix + userId
	ctx1 := tracer.ContextWithSpan(context, span)
	if err := t.RedisUsecase.AddKeyValueSet(ctx1, key, secret, time.Duration(2000000000000000000)); err != nil {
		tracer.LogError(span, err)
		return err
	}

	return nil
}

func (t *totpUsecase) SaveSecret(context context.Context, userId string) error {
	span := tracer.StartSpanFromContext(context, "usecase/SaveSecret")
	defer span.Finish()

	key := secret_prefix + userId

	ctx1 := tracer.ContextWithSpan(context, span)

	secretBytes, err := t.RedisUsecase.GetValueByKey(ctx1, key)

	if err != nil {
		tracer.LogError(span, err)
		return err
	}

	if err := t.RedisUsecase.DeleteValueByKey(ctx1, key); err != nil {
		tracer.LogError(span, err)
		return err
	}

	newTotpSecret := domain.TotpSecret{ProfileInfoId: userId, Secret: string(secretBytes)}
	if err := t.TotpRepository.Create(ctx1, newTotpSecret); err != nil {
		tracer.LogError(span, err)
		return err
	}


	return nil
}