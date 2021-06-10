package handler

import (
	"auth-service/http/middleware"
	"auth-service/infrastructure/dto"
	"auth-service/infrastructure/tracer"
	"auth-service/usecase"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"image/jpeg"
	"os"
)

const (
	totp_error = "Error generating totp secret"
	totp_validation_error = "Your code is not correct"
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

	ctx1 := tracer.ContextWithSpan(ctx, span)

	if !t.TotpUsecase.Validate(ctx1, totpSecretDto.Passcode, totpSecretDto.UserId) {
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
	fmt.Println(ctx.Request.Header.Get("Authorization"))
	token_id := middleware.GetTokenId(ctx1, ctx.Request)

	ctx.JSON(200, token_id)

}


func (t *totpHandler) logMetadata(span opentracing.Span, ctx *gin.Context) {
	span.LogFields(
		tracer.LogString("handler: ", fmt.Sprintf("handling login at %s\n", ctx.Request.URL.Path)),
		tracer.LogString("handler: ", fmt.Sprintf("client ip= %s\n", ctx.ClientIP())),
		tracer.LogString("handler", fmt.Sprintf("method= %s\n", ctx.Request.Method)),
		tracer.LogString("handler", fmt.Sprintf("header= %s\n", ctx.Request.Header)),
	)
}
