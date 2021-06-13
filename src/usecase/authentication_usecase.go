package usecase

import (
	"auth-service/domain"
	"auth-service/infrastructure/tracer"
	"context"
	logger "github.com/jelena-vlajkov/logger/logger"
	"time"
)
const (
	authToken = "authToken"
	refreshToken = "refreshToken"
	totp_token = "totpToken"
)
type authenticationUsecase struct {
	RedisUsecase RedisUsecase
	logger *logger.Logger
}


type AuthenticationUsecase interface {
	SaveAuthToken(ctx context.Context, userId uint, td *domain.TokenDetails) error
	SaveRefreshToken(ctx context.Context, userId uint, td *domain.TokenDetails) error
	FetchAuthToken(ctx context.Context, tokenUuid string) ([]byte, error)
	FetchRefreshToken(ctx context.Context, refreshTokenUuid string) ([]byte, error)
	DeleteAuthToken(ctx context.Context, tokenUuid string) error
	SaveTemporaryToken(ctx context.Context, td *domain.TemporaryTokenDetails) error
	FetchTemporaryToken(ctx context.Context, tokenUuid string) ([]byte, error)
	DeleteTemporaryToken(ctx context.Context, tokenUuid string) error
}

func NewAuthenticationUsecase(redisUsecase RedisUsecase, logger *logger.Logger) AuthenticationUsecase{
	return &authenticationUsecase{redisUsecase, logger}
}


func (a *authenticationUsecase) SaveAuthToken(ctx context.Context, userId uint, td *domain.TokenDetails) error {
	a.logger.Logger.Infof("saving auth token for user %v\n", userId)
	span := tracer.StartSpanFromContext(ctx, "usecase/SaveAuthToken")
	defer span.Finish()

	at := time.Unix(td.AtExpires, 0)
	now := time.Now()

	key := authToken + td.TokenUuid

	if err := a.RedisUsecase.AddKeyValueSet(ctx, key, td.AccessToken, at.Sub(now)); err != nil {
		tracer.LogError(span, err)
		return err
	}

	return nil
}

func (a *authenticationUsecase) FetchAuthToken(ctx context.Context, tokenUuid string) ([]byte, error) {
	a.logger.Logger.Infof("fetching auth token")
	key := authToken + tokenUuid
	value, err := a.RedisUsecase.GetValueByKey(ctx, key)

	if err != nil {
		return nil, err
	}

	return value, err
}

func (a *authenticationUsecase) SaveRefreshToken(ctx context.Context, userId uint, td *domain.TokenDetails) error {
	a.logger.Logger.Infof("saving refresh token for user %v\n", userId)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()

	key := refreshToken + td.TokenUuid
	if err := a.RedisUsecase.AddKeyValueSet(ctx, key, td.RefreshToken, rt.Sub(now)); err != nil {
		return err
	}

	return nil
}

func (a *authenticationUsecase) FetchRefreshToken(ctx context.Context, refreshTokenUuid string) ([]byte, error) {
	a.logger.Logger.Infof("fetching refresh token")
	key := refreshToken + refreshTokenUuid
	value, err := a.RedisUsecase.GetValueByKey(ctx, key)

	if err != nil {
		return nil, err
	}

	return value, err
}

func (a *authenticationUsecase) DeleteAuthToken(ctx context.Context, tokenUuid string) error {
	a.logger.Logger.Infof("deleting auth token")
	key := authToken + tokenUuid

	return a.RedisUsecase.DeleteValueByKey(ctx, key)
}

func (a *authenticationUsecase) SaveTemporaryToken(ctx context.Context, td *domain.TemporaryTokenDetails) error {
	a.logger.Logger.Infof("saving temporary token")
	span := tracer.StartSpanFromContext(ctx, "usecase/SaveTemporaryToken")
	defer span.Finish()

	tokenExp := time.Unix(td.Expires, 0)
	now := time.Now()

	key := totp_token + td.TokenUuid

	if err := a.RedisUsecase.AddKeyValueSet(ctx, key, td.AccessToken, tokenExp.Sub(now)); err != nil {
		tracer.LogError(span, err)
		return err
	}

	return nil
}

func (a *authenticationUsecase) FetchTemporaryToken(ctx context.Context, tokenUuid string) ([]byte, error) {
	a.logger.Logger.Infof("fetching temporary token")
	span := tracer.StartSpanFromContext(ctx, "usecase/FetchTemporaryToken")
	defer span.Finish()

	key := totp_token + tokenUuid

	value, err := a.RedisUsecase.GetValueByKey(ctx, key)

	if err != nil {
		tracer.LogError(span, err)
		return nil, err
	}

	return value, err

}

func (a *authenticationUsecase) DeleteTemporaryToken(ctx context.Context, tokenUuid string) error {
	a.logger.Logger.Infof("deleting temporary token")
	span := tracer.StartSpanFromContext(ctx, "usecase/DeleteTemporaryToken")
	defer span.Finish()

	key := totp_token + tokenUuid

	if err := a.RedisUsecase.DeleteValueByKey(ctx, key); err != nil {
		tracer.LogError(span, err)
		return err
	}

	return nil
}




