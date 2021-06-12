package handler

import (
	"auth-service/http/middleware"
	"auth-service/infrastructure/dto"
	"auth-service/infrastructure/tracer"
	"auth-service/usecase"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	logger "github.com/jelena-vlajkov/logger/logger"
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
	logger *logger.Logger
}


type TotpHandler interface {
	 GenerateSecret(ctx *gin.Context)
	 Verify(ctx *gin.Context)
	 IsEnabled(ctx *gin.Context)
	 Disable(ctx *gin.Context)
}

func NewTotpHandler(totpUsecase usecase.TotpUsecase, tracer opentracing.Tracer, profileInfoUsecase usecase.ProfileInfoUsecase, logger *logger.Logger) TotpHandler {
	return &totpHandler{totpUsecase,  profileInfoUsecase ,tracer, logger}
}

func (t *totpHandler) GenerateSecret(ctx *gin.Context) {
	t.logger.Logger.Infof("Handling GENERATING SECRET")
	span := tracer.StartSpanFromRequest("handler", t.Tracer, ctx.Request)
	defer span.Finish()
	t.logMetadata(span, ctx)
/*

	decoder := json.NewDecoder(ctx.Request.Body)

	var totpDto dto.TotpDto
	if err := decoder.Decode(&totpDto); err != nil {
		t.logger.Logger.Errorf("error while decoding json, error: %v\n", err)
		tracer.LogError(span, fmt.Errorf("message=%s; err=%s\n",body_decoding_err, err))
		ctx.JSON(400, gin.H{"message" : body_decoding_err})
		return
	}

	policy := bluemonday.UGCPolicy()
	totpDto.Username = strings.TrimSpace(policy.Sanitize(totpDto.Username))*/

	ctx1 := tracer.ContextWithSpan(ctx, span)

	userId, err := middleware.ExtractUserId(ctx1, ctx.Request)

	if err != nil {
		t.logger.Logger.Errorf("error while extracting user id, error: %v\n", err)
	}

	user, err := t.ProfileInfoUsecase.GetProfileInfoById(ctx1, userId)
	key, err := t.TotpUsecase.GenereateTotpSecret(ctx1, user.Username)

	if err != nil {
		t.logger.Logger.Errorf("error while generating totp secret, error: %v\n", err)
		tracer.LogError(span, fmt.Errorf("message=%s; err=%s\n",totp_error, err))
		ctx.JSON(400, gin.H{"message" : totp_error})
		return
	}


	if err != nil {
		t.logger.Logger.Errorf("error while getting user info for %v, error: %v\n", user.Username, err)
		tracer.LogError(span, fmt.Errorf("message=%s; err=%s\n",totp_error, err))
		ctx.JSON(400, gin.H{"message" : totp_error})
		return
	}

	if err := t.TotpUsecase.SaveSecretTemporarily(ctx1, user.ID, key.Secret()); err != nil {
		t.logger.Logger.Errorf("error while saving secret for user %v, error: %v\n", user.ID, err)
		tracer.LogError(span, fmt.Errorf("message=%s; err=%s\n",totp_error, err))
		ctx.JSON(400, gin.H{"message" : totp_error})
		return
	}

	workingDirectory, _ := os.Getwd()
	if !strings.HasSuffix(workingDirectory, "src") {
		firstPart := strings.Split(workingDirectory, "src")
		value := firstPart[0] + "/src"
		workingDirectory = value
		os.Chdir(workingDirectory)
	}

	f, err := os.Create(userId + ".jpg")
	if err != nil {
		t.logger.Logger.Errorf("error while creating qr image, error: %v\n", err)
	}
	defer f.Close()

	img, _ := t.TotpUsecase.GetSecretImage(ctx1, key, 300, 300)
	jpeg.Encode(f, *img, nil)
	base64Img, err := t.TotpUsecase.Base64Image(ctx1, userId + ".jpg")

	if err != nil {
		t.logger.Logger.Errorf("error while generating secret for user %v, error: %v\n", user.ID, err)
		tracer.LogError(span, fmt.Errorf("message=%s; err=%s\n",totp_error, err))
		ctx.JSON(400, gin.H{"message" : totp_error})
		return
	}

	ctx.JSON(200, dto.ScanTotpDto{QRCode: base64Img, Secret: key.Secret()})

}

func (t *totpHandler) Verify(ctx *gin.Context) {
	t.logger.Logger.Println("Handling VERIFYING TOTP")
	span := tracer.StartSpanFromRequest("handler", t.Tracer, ctx.Request)
	defer span.Finish()
	t.logMetadata(span, ctx)

	decoder := json.NewDecoder(ctx.Request.Body)

	var totpSecretDto dto.TotpSecretDto
	if err := decoder.Decode(&totpSecretDto); err != nil {
		t.logger.Logger.Errorf("error while decoding json, error: %v\n", err)
		tracer.LogError(span, fmt.Errorf("message=%s; err=%s\n",body_decoding_err, err))
		ctx.JSON(400, gin.H{"message" : body_decoding_err})
		return
	}

	policy := bluemonday.UGCPolicy()
	totpSecretDto.Passcode = strings.TrimSpace(policy.Sanitize(totpSecretDto.Passcode))
	totpSecretDto.UserId = strings.TrimSpace(policy.Sanitize(totpSecretDto.UserId))

	ctx1 := tracer.ContextWithSpan(ctx, span)

	if !t.TotpUsecase.Verify(ctx1, totpSecretDto.Passcode, totpSecretDto.UserId) {
		t.logger.Logger.Errorf("error while verifying totp for user %v", totpSecretDto.UserId)
		tracer.LogError(span, fmt.Errorf("message=%s",totp_validation_error))
		ctx.JSON(400, gin.H{"message" : totp_validation_error})
		return
	}

	if err := t.TotpUsecase.SaveSecret(ctx1, totpSecretDto.UserId); err != nil {
		t.logger.Logger.Errorf("error while saving totp secret for user %v, error: %v\n", totpSecretDto.UserId, err)
		tracer.LogError(span, fmt.Errorf("message=%s",totp_validation_error))
		ctx.JSON(400, gin.H{"message" : totp_validation_error})
		return
	}

	ctx.JSON(200, gin.H{"message" : "Two factor authentication enabled"})
}

func (t *totpHandler) IsEnabled(ctx *gin.Context) {
	t.logger.Logger.Println("Handling IS ENABLED TOTP")
	span := tracer.StartSpanFromRequest("handler", t.Tracer, ctx.Request)
	defer span.Finish()

	ctx1 := tracer.ContextWithSpan(ctx, span)
	//userId, err := middleware.ExtractUserId(ctx1, ctx.Request)
	decoder := json.NewDecoder(ctx.Request.Body)

	var totpDto dto.TotpDto
	if err := decoder.Decode(&totpDto); err != nil {
		t.logger.Logger.Errorf("error while decoding json, error: %v\n", err)
		tracer.LogError(span, fmt.Errorf("message=%s; err=%s\n",body_decoding_err, err))
		ctx.JSON(400, gin.H{"message" : body_decoding_err})
		return
	}
	policy := bluemonday.UGCPolicy()
	totpDto.Username = strings.TrimSpace(policy.Sanitize(totpDto.Username))

	userId := totpDto.Username
	secret, err := t.TotpUsecase.GetSecretByProfileInfoId(ctx1, userId)

	if userId == "" {
		t.logger.Logger.Error("error while getting getting value if totp is enabled, no user id\n")
		tracer.LogError(span, fmt.Errorf("messge= %s; err= %s", totp_is_enabled_wrong_id_error, err))
		ctx.JSON(400, gin.H{"message" : totp_is_enabled_wrong_id_error})
		return
	}

	if secret != nil {
		t.logger.Logger.Errorf("error while getting secret from profile info, no secret")
		tracer.LogError(span, fmt.Errorf("messge= %s; err= %s", totp_is_enabled, err))
		ctx.JSON(400, gin.H{"message" : totp_is_enabled})
		return
	}

	ctx.JSON(200, gin.H{"message" : "Totp is not enabled"})

}

func (t *totpHandler) Disable(ctx *gin.Context) {
	t.logger.Logger.Println("Handling DISABLING TOTP")
	span := tracer.StartSpanFromRequest("handler", t.Tracer, ctx.Request)
	defer span.Finish()
	t.logMetadata(span, ctx)

	decoder := json.NewDecoder(ctx.Request.Body)

	var totpSecretDto dto.TotpSecretDto
	if err := decoder.Decode(&totpSecretDto); err != nil {
		t.logger.Logger.Errorf("error while decoding json, error: %v\n", err)
		tracer.LogError(span, fmt.Errorf("message=%s; err=%s\n",body_decoding_err, err))
		ctx.JSON(400, gin.H{"message" : body_decoding_err})
		return
	}

	policy := bluemonday.UGCPolicy()
	totpSecretDto.Passcode = strings.TrimSpace(policy.Sanitize(totpSecretDto.Passcode))
	totpSecretDto.UserId = strings.TrimSpace(policy.Sanitize(totpSecretDto.UserId))

	ctx1 := tracer.ContextWithSpan(ctx, span)

	if !t.TotpUsecase.Validate(ctx1, totpSecretDto.UserId, totpSecretDto.Passcode) {
		t.logger.Logger.Errorf("error while validating secret for user %v, invalid passcode\n", totpSecretDto.UserId)
		tracer.LogError(span, fmt.Errorf("message=%s",totp_validation_error))
		ctx.JSON(400, gin.H{"message" : totp_validation_error})
		return
	}

	if err := t.TotpUsecase.DeleteSecretByProfileId(ctx1, totpSecretDto.UserId); err != nil {
		t.logger.Logger.Errorf("error while deleting secret for user %v, error: %v\n", totpSecretDto.UserId, err)
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



