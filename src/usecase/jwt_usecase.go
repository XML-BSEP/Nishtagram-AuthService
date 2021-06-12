package usecase

import (
	"auth-service/domain"
	"auth-service/infrastructure/tracer"
	"auth-service/repository"
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	logger "github.com/jelena-vlajkov/logger/logger"
	"github.com/twinj/uuid"
	"os"
	"time"
)

type jwtUsecase struct {
	RedisUsecase RedisUsecase
	roleRepository repository.RoleRepository
	AuthenticationUsecase AuthenticationUsecase
	logger *logger.Logger
}

func (j *jwtUsecase) RefreshToken(context context.Context, tokenString string) (*string, *string, error) {
	span := tracer.StartSpanFromContext(context, "usecase/RefreshToken")
	defer span.Finish()

	ctx1 := tracer.ContextWithSpan(context, span)
	userId, err := j.ExtractUserId(ctx1, tokenString)

	if err != nil {
		tracer.LogError(span, err)
		return nil, nil, err
	}

	role, err := j.ExtractRole(ctx1, tokenString)

	if err != nil {
		tracer.LogError(span, err)
		return nil, nil, err
	}

	td := &domain.TokenDetails{}
	_, err = j.CreateAccessToken(ctx1, *role, *userId, td)

	if err != nil {
		tracer.LogError(span, err)
		return nil, nil, err
	}


	if err := j.AuthenticationUsecase.SaveAuthToken(ctx1, 0, td); err != nil {
		tracer.LogError(span, err)
		return nil, nil, err
	}
	return &td.AccessToken, &td.TokenUuid, nil
}

func (j *jwtUsecase) ExtractUserId(context context.Context, tokenString string) (*string, error) {
	span := tracer.StartSpanFromContext(context, "usecase/ExtractTokenUuid")
	defer span.Finish()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})

	claims, ok := token.Claims.(jwt.MapClaims)

	if ok  {
		role, ok := claims["user_id"].(string)
		if !ok {
			return nil, err
		}
		return &role, nil
	}

	return nil, err
}
func (j *jwtUsecase) ExtractRole(context context.Context, tokenString string) (*string, error) {
	span := tracer.StartSpanFromContext(context, "usecase/ExtractTokenUuid")
	defer span.Finish()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})

	claims, ok := token.Claims.(jwt.MapClaims)

	if ok  {
		role, ok := claims["role"].(string)
		if !ok {
			return nil, err
		}
		return &role, nil
	}

	return nil, err
}

func (j *jwtUsecase) ExtractRefreshUuid(context context.Context, tokenString string) (*string, error) {
	span := tracer.StartSpanFromContext(context, "usecase/ExtractRefreshUuid")
	defer span.Finish()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})

	claims, ok := token.Claims.(jwt.MapClaims)

	if ok  {
		refreshUuid, ok := claims["refresh_uuid"].(string)
		if !ok {
			return nil, err
		}
		return &refreshUuid, nil
	}


	return nil, err
}

func (j *jwtUsecase) ExtractExpiration(context context.Context, tokenString string) (*time.Time, error) {
	span := tracer.StartSpanFromContext(context, "usecase/ExtractExpiration")
	defer span.Finish()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})

	claims, ok := token.Claims.(jwt.MapClaims)

	if ok  {
		expFloat, ok := claims["exp"].(float64)
		if !ok {
			return nil, err
		}
		exp := time.Unix(int64(expFloat), 0)
		return &exp, nil
	}



	return nil, err
}

type JwtUsecase interface {
	CreateAccessToken(context context.Context, role string, userId string, td *domain.TokenDetails) (*domain.TokenDetails, error)
	CreateTemporaryToken(context context.Context, role, userId string) (*domain.TemporaryTokenDetails, error)
	ValidateToken(context context.Context, tokenString string) (string,error)
	CreateRefreshToken(context context.Context, userId, role string, td *domain.TokenDetails) (*domain.TokenDetails, error)
	CreateToken(context context.Context, role, userId string, refresh bool) (*domain.TokenDetails, error)
	ExtractExpiration(context context.Context, tokenString string) (*time.Time, error)
	ExtractRole(context context.Context, tokenString string) (*string, error)
	RefreshToken(context context.Context, tokenString string) (*string, *string, error)
}
func NewJwtUsecase(usecase RedisUsecase, logger *logger.Logger, authUsecase AuthenticationUsecase) JwtUsecase {
	return &jwtUsecase{RedisUsecase: usecase, logger: logger, AuthenticationUsecase: authUsecase}
}

func (j *jwtUsecase) CreateAccessToken(context context.Context, role string, userId string, td *domain.TokenDetails) (*domain.TokenDetails, error) {
	j.logger.Logger.Infof("creating access token for user %v\n", userId)
	span := tracer.StartSpanFromContext(context, "CreateAccessToken")
	defer span.Finish()

	td.AtExpires = time.Now().Add(time.Second * 20).Unix()
	td.TokenUuid = uuid.NewV4().String()


	var err error

	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.TokenUuid
	atClaims["refresh_uuid"] = td.RefreshUuid
	atClaims["exp"] = td.AtExpires
	atClaims["role"] = role
	atClaims["user_id"] = userId


	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))

	if err != nil {
		j.logger.Logger.Errorf("error while creating token, error: %v\n", err)
		tracer.LogError(span, err)
		return nil, err
	}


	return td, nil
}

func (j *jwtUsecase) ValidateToken(context context.Context, tokenString string) (string,error) {
	j.logger.Logger.Infof("validating token")
	token, err := verifyToken(tokenString)
	if err != nil {
		j.logger.Logger.Errorf("error while validating token, error: %v\n", err)
		return "", err
	}
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		j.logger.Logger.Errorf("error while claiming token, error: %v\n", err)
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

func (j *jwtUsecase) CreateTemporaryToken(context context.Context, role, userId string) (*domain.TemporaryTokenDetails, error) {
	j.logger.Logger.Infof("creating temporary token for user %v\n", userId)
	span := tracer.StartSpanFromContext(context, "CreateToken")
	defer span.Finish()

	td := &domain.TemporaryTokenDetails{}
	td.Expires = time.Now().Add(time.Minute * 15).Unix()
	td.TokenUuid = uuid.NewV4().String()

	tokenClaims := jwt.MapClaims{}

	tokenClaims["exp"] = td.Expires
	tokenClaims["user_id"] = userId
	tokenClaims["role"] = "temporary_user"
	tokenClaims["token_uuid"] = td.TokenUuid

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)

	var err error

	if err != nil {
		j.logger.Logger.Errorf("error while claiming token for user %v, error: %v\n", userId, err)
		tracer.LogError(span, err)
		return nil, err
	}

	td.AccessToken, err = token.SignedString([]byte(os.Getenv("ACCESS_SECRET")))

	return td, err
}

func (j *jwtUsecase) CreateRefreshToken(context context.Context, userId, role string, td *domain.TokenDetails) (*domain.TokenDetails, error) {
	j.logger.Logger.Infof("creating refresh for user %v\n", userId)
	span := tracer.StartSpanFromContext(context, "CreateRefreshToken")
	defer span.Finish()

	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
	td.RefreshUuid = td.TokenUuid + "++" + uuid.NewV4().String()

	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user_id"] = userId
	rtClaims["exp"] = td.RtExpires
	rtClaims["role"] = role

	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	refreshTokenString, err := rt.SignedString([]byte(os.Getenv("REFRESH_SECRET")))
	td.RefreshToken = refreshTokenString
	if err != nil {
		j.logger.Logger.Errorf("error while creating refresh token, error: %v\n", err)
		tracer.LogError(span, err)
		return  nil, err
	}
	return td, nil
}

func (j *jwtUsecase) CreateToken(context context.Context, role, userId string, refresh bool) (*domain.TokenDetails, error) {
	j.logger.Logger.Infof("creating token for user %v\n", userId)
	span := tracer.StartSpanFromContext(context, "CreateToken")
	defer span.Finish()

	ctx1 := tracer.ContextWithSpan(context, span)
	td := &domain.TokenDetails{}
	if refresh {
		_, err := j.CreateRefreshToken(ctx1, userId, role, td)
		if err != nil {
			tracer.LogError(span, err)
			return nil, err
		}

		if err := j.AuthenticationUsecase.SaveRefreshToken(ctx1, 12, td); err != nil {
			tracer.LogError(span, err)
			return nil, err
		}
	}

	_, err := j.CreateAccessToken(ctx1, role, userId, td)

	if err != nil {
		tracer.LogError(span, err)
		return nil, err
	}

	if err := j.AuthenticationUsecase.SaveAuthToken(ctx1, 12, td); err != nil {
		tracer.LogError(span, err)
		return nil, err
	}


	return td, nil


}


