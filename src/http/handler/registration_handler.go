package handler

import (
	"auth-service/src/domain"
	"auth-service/src/infrastructure/dto"
	"auth-service/src/usecase"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

type registrationHandler struct {
	RegistrationUsecase usecase.RegistrationUsecase
}

type RegistrationHandler interface {
	Register(ctx *gin.Context)
	ConfirmAccount(ctx *gin.Context)
}

func NewRegistrationHandler(registrationUsecase usecase.RegistrationUsecase) RegistrationHandler {
	return &registrationHandler{RegistrationUsecase: registrationUsecase}
}

func (r *registrationHandler) Register(ctx *gin.Context) {

	decoder := json.NewDecoder(ctx.Request.Body)
	var user domain.User
	if err := decoder.Decode(&user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message" : "Error decoding json"})
		return
	}

	if r.RegistrationUsecase.IsAlreadyRegistered(ctx, user.Username, user.Email) {
		ctx.JSON(http.StatusBadRequest, gin.H{"message" : "User already exists"})
		return
	}
	if err := r.RegistrationUsecase.Register(ctx, user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message" : err.Error()})
	}

	ctx.JSON(http.StatusOK, gin.H{"message" : "Please check your email to confirm registration"})
}

func (r *registrationHandler) ConfirmAccount(ctx *gin.Context) {
	decoder := json.NewDecoder(ctx.Request.Body)
	var dto dto.AccountConfirmationDto
	if err := decoder.Decode(&dto); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message" : "Error decoding json"})
		return
	}

	if err := r.RegistrationUsecase.ConfirmAccount(ctx, dto.Code, dto.Username); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message" : err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message" : "Registration successful"})

}





