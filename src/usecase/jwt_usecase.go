package usecase

import (
	"auth-service/src/domain"
	"auth-service/src/repository"
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/twinj/uuid"
	"os"
	"time"
)

type jwtUsecase struct {
	RedisUsecase RedisUsecase
	roleRepository repository.RoleRepository
}

type JwtUsecase interface {
	CreateToken(context context.Context, role string, userId string) (*domain.TokenDetails, error)
	ValidateToken(context context.Context, tokenString string) (string,error)
}
func NewJwtUsecase(usecase RedisUsecase) JwtUsecase {
	return &jwtUsecase{RedisUsecase: usecase}
}

func (j *jwtUsecase) CreateToken(context context.Context, role string, userId string) (*domain.TokenDetails, error) {
	td := &domain.TokenDetails{}
	td.AtExpires = time.Now().Add(time.Minute * 15).Unix()
	td.TokenUuid = uuid.NewV4().String()
	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
	td.RefreshUuid = td.TokenUuid + "++" + uuid.NewV4().String()

	var err error

	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.TokenUuid
	atClaims["refresh_uuid"] = td.RefreshUuid
	atClaims["exp"] = td.AtExpires
	atClaims["role"] = role


	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))

	if err != nil {
		return nil, err
	}

	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user_id"] = userId
	rtClaims["exp"] = td.RtExpires

	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(os.Getenv("REFRESH_SECRET")))
	if err != nil {
		return nil, err
	}
	return td, nil
}

func (j *jwtUsecase) ValidateToken(context context.Context, tokenString string) (string,error) {
	token, err := verifyToken(tokenString)
	if err != nil {
		return "", err
	}
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return "", err
	}
	return tokenString, nil
}

func verifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}

	return token, nil
}
