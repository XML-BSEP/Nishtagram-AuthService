package handler

import (
	"auth-service/domain"
	"auth-service/http/middleware"
	"auth-service/infrastructure/dto"
	"auth-service/infrastructure/tracer"
	"auth-service/usecase"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"net/http"
)

const (
	body_decoding_err = "Body decoding error"
	invalid_credentials_err = "Wrong username or password"
	token_err = "Can not create token"
	totp_invalid_user_id = "User id is not valid"
)

type authenticateHandler struct {
	AuthenticationUsecase usecase.AuthenticationUsecase
	JwtUsecase usecase.JwtUsecase
	ProfileInfoUsecase usecase.ProfileInfoUsecase
	TotpUsecase usecase.TotpUsecase
	Tracer opentracing.Tracer
}

type AuthenticationHandler interface {
	Login(ctx *gin.Context)
	ValidateToken(ctx *gin.Context)
	Logout(ctx *gin.Context)
}

func NewAuthenticationHandler(authUsecase usecase.AuthenticationUsecase, jwtUSecase usecase.JwtUsecase, profileInfoUsecase usecase.ProfileInfoUsecase, totpUsecase usecase.TotpUsecase, tracer opentracing.Tracer) AuthenticationHandler {
	return &authenticateHandler{authUsecase, jwtUSecase,  profileInfoUsecase, totpUsecase, tracer}
}

func (a *authenticateHandler) Login(ctx *gin.Context) {

	span := tracer.StartSpanFromRequest("Login", a.Tracer, ctx.Request)
	defer span.Finish()
	a.logMetadata(span, ctx)
	var authenticationDto dto.AuthenticationDto

	decoder := json.NewDecoder(ctx.Request.Body)

	if err := decoder.Decode(&authenticationDto); err != nil {
		tracer.LogError(span, fmt.Errorf("message=%s; err=%s\n",body_decoding_err, err))
		ctx.JSON(400, gin.H{"message" : body_decoding_err})
		ctx.Abort()
		return
	}

	span.LogFields(tracer.LogString("handler", fmt.Sprintf("request_username= %s", authenticationDto.Username)))

	ctx1 := tracer.ContextWithSpan(ctx, span)
	profileInfo, err := a.ProfileInfoUsecase.GetProfileInfoByUsername(ctx1, authenticationDto.Username)
	if err != nil {
		tracer.LogError(span, fmt.Errorf("message=%s", invalid_credentials_err))
		ctx.JSON(400, gin.H{"message" : invalid_credentials_err})
		ctx.Abort()
		return
	}

	span.LogFields(tracer.LogString("handler", fmt.Sprintf("username= %s; user_id= %s", profileInfo.Username, profileInfo.ID)))
	if err := usecase.VerifyPassword(ctx1, authenticationDto.Password, profileInfo.Password); err != nil {
		ctx.JSON(400, gin.H{"message" : invalid_credentials_err})
		ctx.Abort()
		return
	}



	_, err = a.TotpUsecase.GetSecretByProfileInfoId(ctx1, profileInfo.ID)

	if err == nil {
		userInfo, err := a.CreateTemporaryToken(ctx1, ctx.Request, profileInfo.ID, profileInfo)
		if err != nil {
			tracer.LogError(span, err)
			ctx.JSON(400, gin.H{"message" : "Can not create temporary code"})
			return
		}

		ctx.JSON(200, userInfo)
		return
	}


	token, err := a.JwtUsecase.CreateToken(ctx1, profileInfo.Role.RoleName, profileInfo.ID)
	if err != nil {
		ctx.JSON(400, gin.H{"message" : token_err})
		ctx.Abort()
		return
	}
	authenticatedUserInfo := dto.AuthenticatedUserInfoDto{
		Token: token.TokenUuid,
		Role: profileInfo.Role.RoleName,
		Id: profileInfo.ID,
	}
	a.AuthenticationUsecase.SaveAuthToken(ctx1, 12, token)

	ctx.JSON(200, authenticatedUserInfo)
}

func (a *authenticateHandler) ValidateToken(ctx *gin.Context) {
	var tokenDto dto.TokenDto

	decoder := json.NewDecoder(ctx.Request.Body)

	if err := decoder.Decode(&tokenDto); err != nil {
		ctx.JSON(400, gin.H{"message" : "Token decoding error"})
		ctx.Abort()
		return
	}

	at, err := a.AuthenticationUsecase.FetchAuthToken(ctx, tokenDto.TokenId)

	if err != nil {
		ctx.JSON(401, gin.H{"message" : "Invalid token"})
		ctx.Abort()
		return
	}

	token, err := a.JwtUsecase.ValidateToken(ctx, string(at))
	if err != nil || token == ""{
		ctx.JSON(401, gin.H{"message" : "Invalid token"})
		ctx.Abort()
		return
	}

	ctx.JSON(200, token)

}

func (a *authenticateHandler) Logout(ctx *gin.Context) {
	var tokenDto dto.TokenDto

	decoder := json.NewDecoder(ctx.Request.Body)

	if err := decoder.Decode(&tokenDto); err != nil {
		ctx.JSON(400, gin.H{"message" : "Token decoding error"})
		ctx.Abort()
		return
	}

	if err := a.AuthenticationUsecase.DeleteAuthToken(ctx, tokenDto.TokenId); err != nil {
		ctx.JSON(400, gin.H{"message" : "Token deleting error"})
		ctx.Abort()
		return
	}

	ctx.JSON(200, gin.H{"message" : "Sucessful logout"})
}

func (a *authenticateHandler) Login1(ctx *gin.Context) {

	ctx.JSON(200, gin.H{"message" : "Sucessful logout"})
}

func (a *authenticateHandler) CreateTemporaryToken(ctx context.Context, request *http.Request, userId string, profileInfo domain.ProfileInfo) (*dto.AuthenticatedUserInfoDto, error){
	span := tracer.StartSpanFromContext(ctx, "handler/createTemporaryToken")
	defer span.Finish()

	ctx1 := tracer.ContextWithSpan(ctx, span)

	role, err := middleware.ExtractUserRole(ctx1, request)

	if err != nil {
		tracer.LogError(span, err)
		return nil, err
	}

	temporaryToken, err := a.JwtUsecase.CreateTemporaryToken(ctx1, role, userId)

	if err != nil {
		tracer.LogError(span, err)
		return nil, err
	}

	authenticatedUserInfo := dto.AuthenticatedUserInfoDto{
		Token: temporaryToken.TokenUuid,
		Role: profileInfo.Role.RoleName,
		Id: profileInfo.ID,
	}

	a.AuthenticationUsecase.SaveTemporaryToken(ctx1, temporaryToken)


	return &authenticatedUserInfo, nil
}

func (a *authenticateHandler) logMetadata(span opentracing.Span, ctx *gin.Context) {
	span.LogFields(
		tracer.LogString("handler: ", fmt.Sprintf("handling login at %s\n", ctx.Request.URL.Path)),
		tracer.LogString("handler: ", fmt.Sprintf("client ip= %s\n", ctx.ClientIP())),
		tracer.LogString("handler", fmt.Sprintf("method= %s\n", ctx.Request.Method)),
		tracer.LogString("handler", fmt.Sprintf("header= %s\n", ctx.Request.Header)),
	)
}

