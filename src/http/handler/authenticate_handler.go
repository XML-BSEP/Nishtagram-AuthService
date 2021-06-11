package handler

import (
	"auth-service/helper"
	"auth-service/infrastructure/dto"
	"auth-service/infrastructure/tracer"
	"auth-service/usecase"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	"github.com/opentracing/opentracing-go"
	"strings"
	"time"
	"unicode"
)

const (
	body_decoding_err = "Body decoding error"
	invalid_credentials_err = "Wrong username or password"
	token_err = "Can not create token"
)
const (
	redisKeyPattern = "passwordResetRequest"
)
type authenticateHandler struct {
	AuthenticationUsecase usecase.AuthenticationUsecase
	JwtUsecase usecase.JwtUsecase
	ProfileInfoUsecase usecase.ProfileInfoUsecase
	Tracer opentracing.Tracer
	RedisUsecase usecase.RedisUsecase
}

type AuthenticationHandler interface {
	Login(ctx *gin.Context)
	ValidateToken(ctx *gin.Context)
	Logout(ctx *gin.Context)
	SendResetMail(ctx *gin.Context)
	ResetPassword(ctx *gin.Context)

}

func NewAuthenticationHandler(authUsecase usecase.AuthenticationUsecase, jwtUSecase usecase.JwtUsecase, profileInfoUsecase usecase.ProfileInfoUsecase, tracer opentracing.Tracer, redis usecase.RedisUsecase) AuthenticationHandler {
	return &authenticateHandler{authUsecase, jwtUSecase,  profileInfoUsecase, tracer, redis}
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

	policy := bluemonday.UGCPolicy()
	authenticationDto.Username = strings.TrimSpace(policy.Sanitize(authenticationDto.Username))
	authenticationDto.Password = strings.TrimSpace(policy.Sanitize(authenticationDto.Password))

	if authenticationDto.Password == " " || authenticationDto.Username == " " {
		ctx.JSON(400, gin.H{"message" : "Field are empty or xss attack happened"})
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

	policy := bluemonday.UGCPolicy()
	tokenDto.TokenId = strings.TrimSpace(policy.Sanitize(tokenDto.TokenId))

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

func (a *authenticateHandler) ResetPassword(ctx *gin.Context) {
	decoder := json.NewDecoder(ctx.Request.Body)

	var resetDto dto.ResetPassDTO

	err := decoder.Decode(&resetDto)

	policy := bluemonday.UGCPolicy();
	resetDto.Email = strings.TrimSpace(policy.Sanitize(resetDto.Email))
	resetDto.Password = strings.TrimSpace(policy.Sanitize(resetDto.Password))
	resetDto.ConfirmedPassword = strings.TrimSpace(policy.Sanitize(resetDto.ConfirmedPassword))
	resetDto.VerificationCode = strings.TrimSpace(policy.Sanitize(resetDto.VerificationCode))

	if err != nil {
		ctx.JSON(400, gin.H{"message" : "Decoding error"})
		return
	}

	if pasval1, pasval2, pasval3, pasval4 := verifyPassword(resetDto.Password); pasval1 == false || pasval2 == false || pasval3 == false || pasval4 == false {
		ctx.JSON(400, gin.H{"message" : "Password must have minimum 1 uppercase letter, 1 lowercase letter, 1 digit and 1 special character and needs to be minimum 8 characters long"})
		return
	}

	errorMessage := a.ProfileInfoUsecase.ResetPassword(ctx, resetDto)

	if errorMessage != "" {
		ctx.JSON(400, gin.H{"message" : errorMessage})
		return
	}

	ctx.JSON(200, gin.H{"message" : "Successfully changed password!"})
	return
}

func (r *authenticateHandler) SendResetMail(ctx *gin.Context) {
	decoder := json.NewDecoder(ctx.Request.Body)

	type Email struct {
		Email	string	`json:"email"`
	}

	var req Email
	err := decoder.Decode(&req)

	policy := bluemonday.UGCPolicy();
	req.Email = strings.TrimSpace(policy.Sanitize(req.Email))

	if err != nil {
		ctx.JSON(400, gin.H{"message" : "Error decoding email"})
		return
	}

	exists :=  r.ProfileInfoUsecase.ExistsByUsernameOrEmail(ctx, "", req.Email)
	if !exists {
		ctx.JSON(400, gin.H{"message" : "User does not exist"})
		return
	}


	user, _ := r.ProfileInfoUsecase.GetProfileInfoByEmail(ctx, req.Email)

	code := helper.RandomStringGenerator(8)

	expiration  := 1000000000 * 3600 * 2 //2h

	hash, _ := helper.Hash(code)

	go usecase.SendRestartPasswordMail(user.Email , code)


	redisKey := redisKeyPattern + user.Email


	err = r.RedisUsecase.AddKeyValueSet(ctx, redisKey, hash, time.Duration(expiration))
	if err != nil {
		ctx.JSON(500, gin.H{"message" : "Error"})
		return
	}

	ctx.JSON(200, gin.H{"message" : "Check email!"})
	return
}


func (a *authenticateHandler) Login1(ctx *gin.Context) {

	ctx.JSON(200, gin.H{"message" : "Sucessful logout"})
}

func serialize(value string) ([]byte, error){
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

func verifyPassword(s string) (eightOrMore, number, upper, special bool)  {
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
