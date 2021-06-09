package usecase

import (
	"auth-service/domain"
	"auth-service/infrastructure/tracer"
	"context"
	"time"
)
const (
	authToken = "authToken"
	refreshToken = "refreshToken"
)
type authenticationUsecase struct {
	RedisUsecase RedisUsecase
}

type AuthenticationUsecase interface {
	SaveAuthToken(ctx context.Context, userId uint, td *domain.TokenDetails) error
	SaveRefreshToken(ctx context.Context, userId uint, td *domain.TokenDetails) error
	FetchAuthToken(ctx context.Context, tokenUuid string) ([]byte, error)
	FetchRefreshToken(ctx context.Context, refreshTokenUuid string) ([]byte, error)
	DeleteAuthToken(ctx context.Context, tokenUuid string) error
}

func NewAuthenticationUsecase(redisUsecase RedisUsecase) AuthenticationUsecase{
	return &authenticationUsecase{redisUsecase}
}


func (a *authenticationUsecase) SaveAuthToken(ctx context.Context, userId uint, td *domain.TokenDetails) error {
	span := tracer.StartSpanFromContext(ctx, "SaveAuthToken")
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
	key := authToken + tokenUuid
	value, err := a.RedisUsecase.GetValueByKey(ctx, key)

	if err != nil {
		return nil, err
	}

	return value, err
}

func (a *authenticationUsecase) SaveRefreshToken(ctx context.Context, userId uint, td *domain.TokenDetails) error {
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()

	key := refreshToken + td.RefreshUuid
	if err := a.RedisUsecase.AddKeyValueSet(ctx, key, td.AccessToken, rt.Sub(now)); err != nil {
		return err
	}

	return nil
}

func (a *authenticationUsecase) FetchRefreshToken(ctx context.Context, refreshTokenUuid string) ([]byte, error) {
	key := refreshToken + refreshTokenUuid
	value, err := a.RedisUsecase.GetValueByKey(ctx, key)

	if err != nil {
		return nil, err
	}

	return value, err
}

func (a *authenticationUsecase) DeleteAuthToken(ctx context.Context, tokenUuid string) error {
	key := authToken + tokenUuid

	return a.RedisUsecase.DeleteValueByKey(ctx, key)
}




