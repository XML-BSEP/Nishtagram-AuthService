package handler

import (
	"auth-service/domain"
	"auth-service/helper"
	"auth-service/http/middleware"
	"auth-service/infrastructure/dto"
	"auth-service/infrastructure/tracer"
	"auth-service/usecase"
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	logger "github.com/jelena-vlajkov/logger/logger"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	"github.com/opentracing/opentracing-go"
)

const (
	body_decoding_err       = "Body decoding error"
	invalid_credentials_err = "Wrong username or password"
	token_err               = "Can not create token"
	totp_invalid_user_id    = "User id is not valid"
)
const (
	redisKeyPattern = "passwordResetRequest"
)

type authenticateHandler struct {
	AuthenticationUsecase usecase.AuthenticationUsecase
	JwtUsecase            usecase.JwtUsecase
	ProfileInfoUsecase    usecase.ProfileInfoUsecase
	TotpUsecase           usecase.TotpUsecase
	Tracer                opentracing.Tracer
	RedisUsecase          usecase.RedisUsecase
	logger *logger.Logger
}


type AuthenticationHandler interface {
	Login(ctx *gin.Context)
	ValidateToken(ctx *gin.Context)
	Logout(ctx *gin.Context)
	Validate(ctx *gin.Context)
	ValidateTemporaryToken(ctx *gin.Context)
	SendResetMail(ctx *gin.Context)
	ResetPassword(ctx *gin.Context)
	RefreshToken(ctx *gin.Context)
	Login1(ctx *gin.Context)
}

func NewAuthenticationHandler(authUsecase usecase.AuthenticationUsecase, jwtUSecase usecase.JwtUsecase, profileInfoUsecase usecase.ProfileInfoUsecase, tracer opentracing.Tracer, redis usecase.RedisUsecase, totpUsecase usecase.TotpUsecase, logger *logger.Logger) AuthenticationHandler {
	return &authenticateHandler{authUsecase, jwtUSecase, profileInfoUsecase, totpUsecase, tracer, redis, logger}

}

func (a *authenticateHandler) Login(ctx *gin.Context) {
	a.logger.Logger.Println("Handling LOGIN")
	fmt.Println("tusam 1")
	span := tracer.StartSpanFromRequest("Login", a.Tracer, ctx.Request)
	defer span.Finish()
	a.logMetadata(span, ctx)
	var authenticationDto dto.AuthenticationDto

	decoder := json.NewDecoder(ctx.Request.Body)

	if err := decoder.Decode(&authenticationDto); err != nil {
		a.logger.Logger.Errorf("error while decoding json, error: %v\n", err)
		tracer.LogError(span, fmt.Errorf("message=%s; err=%s\n", body_decoding_err, err))
		ctx.JSON(400, gin.H{"message": body_decoding_err})
		ctx.Abort()
		return
	}

	fmt.Println(authenticationDto)
	policy := bluemonday.UGCPolicy()
	authenticationDto.Username = strings.TrimSpace(policy.Sanitize(authenticationDto.Username))
	authenticationDto.Password = strings.TrimSpace(policy.Sanitize(authenticationDto.Password))

	if authenticationDto.Password == " " || authenticationDto.Username == " " {
		a.logger.Logger.Error("error while login, error: fields are empty")
		a.logger.Logger.Warnf("possible xss attack from IP address %v\n", ctx.Request.Host)
		ctx.JSON(400, gin.H{"message": "Field are empty or xss attack happened"})
		return
	}

	span.LogFields(tracer.LogString("handler", fmt.Sprintf("request_username= %s", authenticationDto.Username)))

	ctx1 := tracer.ContextWithSpan(ctx, span)
	profileInfo, err := a.ProfileInfoUsecase.GetProfileInfoByUsername(ctx1, authenticationDto.Username)
	if err != nil {
		a.logger.Logger.Errorf("error while getting profile info by for email %v, error: %v\n", authenticationDto.Username, err)
		tracer.LogError(span, fmt.Errorf("message=%s", invalid_credentials_err))
		ctx.JSON(400, gin.H{"message": invalid_credentials_err})
		ctx.Abort()
		return
	}

	span.LogFields(tracer.LogString("handler", fmt.Sprintf("username= %s; user_id= %s", profileInfo.Username, profileInfo.ID)))
	if err := usecase.VerifyPassword(ctx1, authenticationDto.Password, profileInfo.Password); err != nil {
		a.logger.Logger.Warnf("invalid password for user %v from IP address %v, error: %v\n", authenticationDto.Username, ctx.Request.Host, err)
		ctx.JSON(400, gin.H{"message": invalid_credentials_err})
		ctx.Abort()
		return
	}

	_, err = a.TotpUsecase.GetSecretByProfileInfoId(ctx1, profileInfo.ID)

	fmt.Println("generating secret")
	if err == nil {
		userInfo, err := a.CreateTemporaryToken(ctx1, profileInfo)
		if err != nil {
			a.logger.Logger.Errorf("error while getting temporary user token for %v, error: %v\n", profileInfo.ID, err)
			tracer.LogError(span, err)
			ctx.JSON(400, gin.H{"message": "Can not create temporary code"})
			return
		}
		userInfo.Role = "temporary_user"
		ctx.JSON(200, userInfo)
		return
	}

	val, err := a.generateToken(ctx1, profileInfo, authenticationDto.Refresh)

	fmt.Println("generating token")
	if err != nil {
		a.logger.Logger.Errorf("error while generating token for %v, error: %v\n", profileInfo.ID, err)
		ctx.JSON(400, gin.H{"message": token_err})
		ctx.Abort()
		return
	}

	ctx.JSON(200, val)

	/*token, err := a.JwtUsecase.CreateToken(ctx1, profileInfo.Role.RoleName, profileInfo.ID)
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

	ctx.JSON(200, authenticatedUserInfo)*/
}

func (a *authenticateHandler) ValidateToken(ctx *gin.Context) {
	a.logger.Logger.Println("Handling VALIDATING TOKEN")
	var tokenDto dto.TokenDto

	decoder := json.NewDecoder(ctx.Request.Body)

	if err := decoder.Decode(&tokenDto); err != nil {
		a.logger.Logger.Errorf("error while decoding json, error: %v\n", err)
		ctx.JSON(400, gin.H{"message": "Token decoding error"})
		ctx.Abort()
		return
	}

	policy := bluemonday.UGCPolicy()
	tokenDto.TokenId = strings.TrimSpace(policy.Sanitize(tokenDto.TokenId))

	at, err := a.AuthenticationUsecase.FetchAuthToken(ctx, tokenDto.TokenId)

	if err != nil {
		a.logger.Logger.Errorf("error while fetching token, error: %v\n", err)
		ctx.JSON(401, gin.H{"message": "Invalid token"})
		ctx.Abort()
		return
	}

	token, err := a.JwtUsecase.ValidateToken(ctx, string(at))
	if err != nil || token == "" {
		a.logger.Logger.Errorf("error while validating token, error: %v\n", err)
		ctx.JSON(401, gin.H{"message": "Invalid token"})
		ctx.Abort()
		return
	}

	ctx.JSON(200, token)

}

func (a *authenticateHandler) Logout(ctx *gin.Context) {
	a.logger.Logger.Println("Handling LOGOUT")
	var tokenDto dto.TokenDto

	decoder := json.NewDecoder(ctx.Request.Body)

	if err := decoder.Decode(&tokenDto); err != nil {
		a.logger.Logger.Errorf("error while decoding json, error: %v\n", err)
		ctx.JSON(400, gin.H{"message": "Token decoding error"})
		ctx.Abort()
		return
	}

	if err := a.AuthenticationUsecase.DeleteAuthToken(ctx, tokenDto.TokenId); err != nil {
		a.logger.Logger.Errorf("error while deleting auth token, error: %v\n", err)
		ctx.JSON(400, gin.H{"message": "Token deleting error"})
		ctx.Abort()
		return
	}

	ctx.JSON(200, gin.H{"message": "Sucessful logout"})
}

func (a *authenticateHandler) ResetPassword(ctx *gin.Context) {
	a.logger.Logger.Println("Handling RESET PASSWORD")
	decoder := json.NewDecoder(ctx.Request.Body)

	var resetDto dto.ResetPassDTO

	err := decoder.Decode(&resetDto)

	policy := bluemonday.UGCPolicy()
	resetDto.Email = strings.TrimSpace(policy.Sanitize(resetDto.Email))
	resetDto.Password = strings.TrimSpace(policy.Sanitize(resetDto.Password))
	resetDto.ConfirmedPassword = strings.TrimSpace(policy.Sanitize(resetDto.ConfirmedPassword))
	resetDto.VerificationCode = strings.TrimSpace(policy.Sanitize(resetDto.VerificationCode))

	if err != nil {
		a.logger.Logger.Errorf("error while decoding json, error: %v\n", err)
		ctx.JSON(400, gin.H{"message": "Decoding error"})
		return
	}

	if pasval1, pasval2, pasval3, pasval4 := verifyAuthPassword(resetDto.Password); pasval1 == false || pasval2 == false || pasval3 == false || pasval4 == false {
		a.logger.Logger.Error("error while verifying password, password is not matching pattern")
		ctx.JSON(400, gin.H{"message": "Password must have minimum 1 uppercase letter, 1 lowercase letter, 1 digit and 1 special character and needs to be minimum 8 characters long"})
		return
	}

	errorMessage := a.ProfileInfoUsecase.ResetPassword(ctx, resetDto)

	if errorMessage != "" {
		a.logger.Logger.Error(errorMessage)
		ctx.JSON(400, gin.H{"message": errorMessage})
		return
	}

	ctx.JSON(200, gin.H{"message": "Successfully changed password!"})
	return
}

func (r *authenticateHandler) SendResetMail(ctx *gin.Context) {
	r.logger.Logger.Println("Handling SENDING RESET MAIL")
	decoder := json.NewDecoder(ctx.Request.Body)

	type Email struct {
		Email string `json:"email"`
	}

	var req Email
	err := decoder.Decode(&req)

	policy := bluemonday.UGCPolicy()
	req.Email = strings.TrimSpace(policy.Sanitize(req.Email))

	if err != nil {
		r.logger.Logger.Errorf("error while decoding json, error: %v\n", err)
		ctx.JSON(400, gin.H{"message": "Error decoding email"})
		return
	}

	exists := r.ProfileInfoUsecase.ExistsByUsernameOrEmail(ctx, "", req.Email)
	if !exists {
		r.logger.Logger.Errorf("error while sending reset mail, error: no user with email %v\n", req.Email)
		ctx.JSON(400, gin.H{"message": "User does not exist"})
		return
	}

	user, _ := r.ProfileInfoUsecase.GetProfileInfoByEmail(ctx, req.Email)

	code := helper.RandomStringGenerator(8)

	expiration := 1000000000 * 3600 * 2 //2h

	hash, _ := helper.Hash(code)

	go usecase.SendRestartPasswordMail(user.Email, code)

	redisKey := redisKeyPattern + user.Email

	err = r.RedisUsecase.AddKeyValueSet(ctx, redisKey, hash, time.Duration(expiration))
	if err != nil {
		r.logger.Logger.Errorf("error while adding token to redis, error: %v\n", err)
		ctx.JSON(500, gin.H{"message": "Error"})
		return
	}

	ctx.JSON(200, gin.H{"message": "Check email!"})
	return
}

func (a *authenticateHandler) Login1(ctx *gin.Context) {
	var authenticationDto dto.AuthenticationDto

	decoder := json.NewDecoder(ctx.Request.Body)

	if err := decoder.Decode(&authenticationDto); err != nil {
		ctx.JSON(400, gin.H{"message": body_decoding_err})
		ctx.Abort()
		return
	}

	ctx.JSON(200, gin.H{"username": authenticationDto.Username, "password" : authenticationDto.Password})
}

func (a *authenticateHandler) CreateTemporaryToken(ctx context.Context, profileInfo domain.ProfileInfo) (*dto.AuthenticatedUserInfoDto, error) {
	a.logger.Logger.Println("Handling CREATING TEMPORARY TOKEN")
	span := tracer.StartSpanFromContext(ctx, "handler/createTemporaryToken")
	defer span.Finish()

	ctx1 := tracer.ContextWithSpan(ctx, span)

	temporaryToken, err := a.JwtUsecase.CreateTemporaryToken(ctx1, profileInfo.Role.RoleName, profileInfo.ID)

	if err != nil {
		a.logger.Logger.Errorf("error while creating temporary token, error: %v\n", err)
		tracer.LogError(span, err)
		return nil, err
	}

	authenticatedUserInfo := dto.AuthenticatedUserInfoDto{
		Token: temporaryToken.TokenUuid,
		Role:  profileInfo.Role.RoleName,
		Id:    profileInfo.ID,
	}

	if err := a.AuthenticationUsecase.SaveTemporaryToken(ctx1, temporaryToken); err != nil {
		a.logger.Logger.Errorf("error while saving temporary token, error %v\n", err)
		tracer.LogError(span, err)
		return nil, err
	}

	return &authenticatedUserInfo, nil
}

func (a *authenticateHandler) Validate(ctx *gin.Context) {
	a.logger.Logger.Println("Handling VALIDATING TOTP")
	span := tracer.StartSpanFromRequest("handler", a.Tracer, ctx.Request)
	defer span.Finish()
	a.logMetadata(span, ctx)

	decoder := json.NewDecoder(ctx.Request.Body)

	var totpSecretDto dto.TotpValidationDto
	if err := decoder.Decode(&totpSecretDto); err != nil {
		a.logger.Logger.Errorf("error while decoding json, error: %v\n", err)
		tracer.LogError(span, fmt.Errorf("message=%s; err=%s\n", body_decoding_err, err))
		ctx.JSON(400, gin.H{"message": body_decoding_err})
		return
	}

	ctx1 := tracer.ContextWithSpan(ctx, span)
	userId, err := middleware.ExtractUserId(ctx1, ctx.Request)

	if err != nil {
		a.logger.Logger.Errorf("error while extracting user from token, error: %v\n", err)
		tracer.LogError(span, fmt.Errorf("message=%s; err=%s\n", token_invalid, err))
		ctx.JSON(400, gin.H{"message": token_invalid})
		return
	}

	if !a.TotpUsecase.Validate(ctx1, userId, totpSecretDto.Passcode) {
		a.logger.Logger.Error("error while validating totp passcode, error: passcode not valid")
		tracer.LogError(span, fmt.Errorf("message= %s", totp_validation_error))
		ctx.JSON(400, gin.H{"message": totp_validation_error})
		return
	}

	profileInfo, err := a.ProfileInfoUsecase.GetProfileInfoById(ctx1, userId)
	val, err := a.generateToken(ctx1, *profileInfo, totpSecretDto.Refresh)

	if err != nil {
		a.logger.Logger.Errorf("error while generating token, error: %v\n", err)
		ctx.JSON(400, gin.H{"message": token_err})
		ctx.Abort()
		return
	}

	tokenUuid, err := middleware.ExtractTokenUuid(ctx1, ctx.Request)
	if err := a.AuthenticationUsecase.DeleteTemporaryToken(ctx1, tokenUuid); err != nil {
		a.logger.Logger.Errorf("error while extracting token uuid, error: %v\n", err)
		tracer.LogError(span, fmt.Errorf("message= %s", totp_validation_error))
		ctx.JSON(400, gin.H{"message": totp_validation_error})
		return
	}

	ctx.JSON(200, val)

}

func (a *authenticateHandler) generateToken(ctx context.Context, profileInfo domain.ProfileInfo, refresh bool) (*dto.AuthenticatedUserInfoDto, error) {
	a.logger.Logger.Println("Handling GENERATING TOKEN")
	span := tracer.StartSpanFromContext(ctx, "handler/generateToken")
	defer span.Finish()

	ctx1 := tracer.ContextWithSpan(ctx, span)

	token, err := a.JwtUsecase.CreateToken(ctx1, profileInfo.Role.RoleName, profileInfo.ID, true)
	if err != nil {
		a.logger.Logger.Errorf("error while creating token, error: %v\n", err)
		return nil, err
	}
	authenticatedUserInfo := dto.AuthenticatedUserInfoDto{
		Token: token.TokenUuid,
		Role:  profileInfo.Role.RoleName,
		Id:    profileInfo.ID,
	}


	return &authenticatedUserInfo, nil
}

func (a *authenticateHandler) ValidateTemporaryToken(ctx *gin.Context) {
	a.logger.Logger.Println("Handling VALIDATING TEMPORARY TOKEN")
	span := tracer.StartSpanFromContext(ctx, "handler/ValidateTemporaryToken")
	defer span.Finish()

	var tokenDto dto.TokenDto

	decoder := json.NewDecoder(ctx.Request.Body)

	if err := decoder.Decode(&tokenDto); err != nil {
		a.logger.Logger.Errorf("error while decoding json, error: %v\n", err)
		ctx.JSON(400, gin.H{"message": "Token decoding error"})
		ctx.Abort()
		return
	}

	at, err := a.AuthenticationUsecase.FetchTemporaryToken(ctx, tokenDto.TokenId)

	if err != nil {
		a.logger.Logger.Errorf("error while fetch temporary token, error: %v\n", err)
		ctx.JSON(401, gin.H{"message": "Invalid token"})
		ctx.Abort()
		return
	}

	token, err := a.JwtUsecase.ValidateToken(ctx, string(at))
	if err != nil || token == "" {
		a.logger.Logger.Errorf("error while validating token, error: %v\n", err)
		ctx.JSON(401, gin.H{"message": "Invalid token"})
		ctx.Abort()
		return
	}

	ctx.JSON(200, token)
}

func serialize(value string) ([]byte, error) {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	if err := e.Encode(value); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (a *authenticateHandler) logMetadata(span opentracing.Span, ctx *gin.Context) {
	span.LogFields(
		tracer.LogString("handler: ", fmt.Sprintf("handling login at %s\n", ctx.Request.URL.Path)),
		tracer.LogString("handler: ", fmt.Sprintf("client ip= %s\n", ctx.ClientIP())),
		tracer.LogString("handler", fmt.Sprintf("method= %s\n", ctx.Request.Method)),
		tracer.LogString("handler", fmt.Sprintf("header= %s\n", ctx.Request.Header)),
	)
}

func verifyAuthPassword(s string) (eightOrMore, number, upper, special bool) {
	letters := 0
	for _, c := range s {
		switch {
		case unicode.IsNumber(c):
			number = true
			letters++
		case unicode.IsUpper(c):
			upper = true
			letters++
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			special = true
			letters++
		case unicode.IsLetter(c) || c == ' ':
			letters++
		default:
			return false, false, false, false
		}
	}
	eightOrMore = letters >= 8
	return
}

func (a *authenticateHandler) RefreshToken(ctx *gin.Context) {
	a.logger.Logger.Println("Handling LOGIN")
	span := tracer.StartSpanFromRequest("Login", a.Tracer, ctx.Request)
	defer span.Finish()
	a.logMetadata(span, ctx)

	var tokenDto dto.TokenDto

	decoder := json.NewDecoder(ctx.Request.Body)

	if err := decoder.Decode(&tokenDto); err != nil {
		a.logger.Logger.Errorf("error while decoding json, error: %v\n", err)
		ctx.JSON(400, gin.H{"message": "Token decoding error"})
		ctx.Abort()
		return
	}

	policy := bluemonday.UGCPolicy()
	tokenDto.TokenId = strings.TrimSpace(policy.Sanitize(tokenDto.TokenId))

	ctx1 := tracer.ContextWithSpan(ctx, span)
	rt, err := a.AuthenticationUsecase.FetchRefreshToken(ctx1, tokenDto.TokenId)

	if err != nil {
		a.logger.Logger.Errorf("error while fetching refresh token, error: %v\n", err)
		ctx.JSON(401, gin.H{"message": "Invalid token"})
		return
	}

	token, uuid, err := a.JwtUsecase.RefreshToken(ctx1, string(rt))

	if err != nil {
		a.logger.Logger.Errorf("error while refreshing token, error: %v\n", err)
		ctx.JSON(401, gin.H{"message": "Invalid token"})
		return
	}

	err = a.JwtUsecase.DeleteRefreshToken(ctx1, tokenDto.TokenId)

	if err != nil {
		a.logger.Logger.Errorf("error while refreshing token, error: %v\n", err)
		ctx.JSON(401, gin.H{"message": "Invalid token"})
		return
	}


	refreshTokenDto := dto.RefreshTokenDto{TokenUuid: *uuid, Token: *token}
	a.logger.Logger.Infof("token refreshed for token %v\n", string(rt))
	ctx.JSON(200, refreshTokenDto)
}