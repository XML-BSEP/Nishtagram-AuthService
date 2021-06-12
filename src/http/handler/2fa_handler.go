package handler

import (
	"auth-service/http/middleware"
	"auth-service/infrastructure/dto"
	"auth-service/infrastructure/tracer"
	"auth-service/usecase"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	"github.com/opentracing/opentracing-go"
	"image/jpeg"
	"os"
	"strings"
)

const (
	totp_error = "Error generating totp secret"
	totp_validation_error = "Your code is not correct"
	totp_is_enabled_wrong_id_error = "User does not exists"
	totp_is_enabled = "Two factor authentication is already enabled"
	totp_disable_error = "Two factor authentication disabling error"
	totp_disable = "Two factor authentication successfully disabled"
	token_invalid = "Temporary token is not valid"
)

type totpHandler struct {
	TotpUsecase usecase.TotpUsecase
	ProfileInfoUsecase usecase.ProfileInfoUsecase
	Tracer opentracing.Tracer
}


type TotpHandler interface {
	 GenerateSecret(ctx *gin.Context)
	 Verify(ctx *gin.Context)
	 IsEnabled(ctx *gin.Context)
	 Disable(ctx *gin.Context)
}

func NewTotpHandler(totpUsecase usecase.TotpUsecase, tracer opentracing.Tracer, profileInfoUsecase usecase.ProfileInfoUsecase) TotpHandler {
	return &totpHandler{totpUsecase,  profileInfoUsecase ,tracer}
}

func (t *totpHandler) GenerateSecret(ctx *gin.Context) {
	span := tracer.StartSpanFromRequest("handler", t.Tracer, ctx.Request)
	defer span.Finish()
	t.logMetadata(span, ctx)

	decoder := json.NewDecoder(ctx.Request.Body)

	var totpDto dto.TotpDto
	if err := decoder.Decode(&totpDto); err != nil {
		tracer.LogError(span, fmt.Errorf("message=%s; err=%s\n",body_decoding_err, err))
		ctx.JSON(400, gin.H{"message" : body_decoding_err})
		return
	}

	policy := bluemonday.UGCPolicy()
	totpDto.Username = strings.TrimSpace(policy.Sanitize(totpDto.Username))

	ctx1 := tracer.ContextWithSpan(ctx, span)

	key, err := t.TotpUsecase.GenereateTotpSecret(ctx1, totpDto.Username)

	if err != nil {
		tracer.LogError(span, fmt.Errorf("message=%s; err=%s\n",totp_error, err))
		ctx.JSON(400, gin.H{"message" : totp_error})
		return
	}

	user, err := t.ProfileInfoUsecase.GetProfileInfoByUsername(ctx1, totpDto.Username)

	if err != nil {
		tracer.LogError(span, fmt.Errorf("message=%s; err=%s\n",totp_error, err))
		ctx.JSON(400, gin.H{"message" : totp_error})
		return
	}

	if err := t.TotpUsecase.SaveSecretTemporarily(ctx1, user.ID, key.Secret()); err != nil {
		tracer.LogError(span, fmt.Errorf("message=%s; err=%s\n",totp_error, err))
		ctx.JSON(400, gin.H{"message" : totp_error})
		return
	}

	f, err := os.Create("img.jpg")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	img, _ := t.TotpUsecase.GetSecretImage(ctx1, key, 300, 300)
	jpeg.Encode(f, *img, nil)

	ctx.JSON(200, key.Secret())

}

func (t *totpHandler) Verify(ctx *gin.Context) {
	span := tracer.StartSpanFromRequest("handler", t.Tracer, ctx.Request)
	defer span.Finish()
	t.logMetadata(span, ctx)

	decoder := json.NewDecoder(ctx.Request.Body)

	var totpSecretDto dto.TotpSecretDto
	if err := decoder.Decode(&totpSecretDto); err != nil {
		tracer.LogError(span, fmt.Errorf("message=%s; err=%s\n",body_decoding_err, err))
		ctx.JSON(400, gin.H{"message" : body_decoding_err})
		return
	}

	policy := bluemonday.UGCPolicy()
	totpSecretDto.Passcode = strings.TrimSpace(policy.Sanitize(totpSecretDto.Passcode))
	totpSecretDto.UserId = strings.TrimSpace(policy.Sanitize(totpSecretDto.UserId))

	ctx1 := tracer.ContextWithSpan(ctx, span)

	if !t.TotpUsecase.Verify(ctx1, totpSecretDto.Passcode, totpSecretDto.UserId) {
		tracer.LogError(span, fmt.Errorf("message=%s",totp_validation_error))
		ctx.JSON(400, gin.H{"message" : totp_validation_error})
		return
	}

	if err := t.TotpUsecase.SaveSecret(ctx1, totpSecretDto.UserId); err != nil {
		tracer.LogError(span, fmt.Errorf("message=%s",totp_validation_error))
		ctx.JSON(400, gin.H{"message" : totp_validation_error})
		return
	}

	ctx.JSON(200, gin.H{"message" : "Two factor authentication enabled"})
}

func (t *totpHandler) IsEnabled(ctx *gin.Context) {
	span := tracer.StartSpanFromRequest("handler", t.Tracer, ctx.Request)
	defer span.Finish()

	ctx1 := tracer.ContextWithSpan(ctx, span)
	userId, err := middleware.ExtractUserId(ctx1, ctx.Request)

	if err != nil {
		tracer.LogError(span, fmt.Errorf("messge= %s; err= %s", totp_is_enabled_wrong_id_error, err))
		ctx.JSON(400, gin.H{"message" : totp_is_enabled_wrong_id_error})
		return
	}

	if userId == "" {
		tracer.LogError(span, fmt.Errorf("messge= %s; err= %s", totp_is_enabled_wrong_id_error, err))
		ctx.JSON(400, gin.H{"message" : totp_is_enabled_wrong_id_error})
		return
	}

	secret, _ := t.TotpUsecase.GetSecretByProfileInfoId(ctx1, userId)


	if secret != nil {
		tracer.LogError(span, fmt.Errorf("messge= %s; err= %s", totp_is_enabled, err))
		ctx.JSON(400, gin.H{"message" : totp_is_enabled})
		return
	}

	ctx.JSON(200, gin.H{"message" : "Totp is not enabled"})

}

func (t *totpHandler) Disable(ctx *gin.Context) {
	span := tracer.StartSpanFromRequest("handler", t.Tracer, ctx.Request)
	defer span.Finish()
	t.logMetadata(span, ctx)

	decoder := json.NewDecoder(ctx.Request.Body)

	var totpSecretDto dto.TotpSecretDto
	if err := decoder.Decode(&totpSecretDto); err != nil {
		tracer.LogError(span, fmt.Errorf("message=%s; err=%s\n",body_decoding_err, err))
		ctx.JSON(400, gin.H{"message" : body_decoding_err})
		return
	}

	policy := bluemonday.UGCPolicy()
	totpSecretDto.Passcode = strings.TrimSpace(policy.Sanitize(totpSecretDto.Passcode))
	totpSecretDto.UserId = strings.TrimSpace(policy.Sanitize(totpSecretDto.UserId))

	ctx1 := tracer.ContextWithSpan(ctx, span)

	if !t.TotpUsecase.Validate(ctx1, totpSecretDto.UserId, totpSecretDto.Passcode) {
		tracer.LogError(span, fmt.Errorf("message=%s",totp_validation_error))
		ctx.JSON(400, gin.H{"message" : totp_validation_error})
		return
	}

	if err := t.TotpUsecase.DeleteSecretByProfileId(ctx1, totpSecretDto.UserId); err != nil {
		tracer.LogError(span, fmt.Errorf("message=%s; err=%s\n", totp_disable_error, err))
		ctx.JSON(400, gin.H{"message" : totp_disable_error})
		return
	}


	ctx.JSON(200, gin.H{"message" : totp_disable})

}



func (t *totpHandler) logMetadata(span opentracing.Span, ctx *gin.Context) {
	span.LogFields(
		tracer.LogString("handler: ", fmt.Sprintf("handling login at %s\n", ctx.Request.URL.Path)),
		tracer.LogString("handler: ", fmt.Sprintf("client ip= %s\n", ctx.ClientIP())),
		tracer.LogString("handler", fmt.Sprintf("method= %s\n", ctx.Request.Method)),
		tracer.LogString("handler", fmt.Sprintf("header= %s\n", ctx.Request.Header)),
	)
}



