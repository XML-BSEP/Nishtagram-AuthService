package handler

import (
	"auth-service/domain"
	"auth-service/infrastructure/dto"
	validator2 "auth-service/infrastructure/validator"
	"auth-service/usecase"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	"net/http"
	"strings"
)

type registrationHandler struct {
	RegistrationUsecase usecase.RegistrationUsecase

}

type RegistrationHandler interface {
	Register(ctx *gin.Context)
	ConfirmAccount(ctx *gin.Context)
	ResendCode(ctx *gin.Context)
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


	policy := bluemonday.UGCPolicy()
	user.ID = strings.TrimSpace(policy.Sanitize(user.ID))
	user.Name = strings.TrimSpace(policy.Sanitize(user.Name))
	user.Surname = strings.TrimSpace(policy.Sanitize(user.Surname))
	user.Username = strings.TrimSpace(policy.Sanitize(user.Username))
	user.Password = strings.TrimSpace(policy.Sanitize(user.Password))
	user.Email = strings.TrimSpace(policy.Sanitize(user.Email))
	user.Address = strings.TrimSpace(policy.Sanitize(user.Address))
	user.Phone = strings.TrimSpace(policy.Sanitize(user.Phone))
	user.Birthday = strings.TrimSpace(policy.Sanitize(user.Birthday))
	user.Gender = strings.TrimSpace(policy.Sanitize(user.Gender))
	user.Web = strings.TrimSpace(policy.Sanitize(user.Web))
	user.Bio = strings.TrimSpace(policy.Sanitize(user.Bio))
	user.Image = strings.TrimSpace(policy.Sanitize(user.Image))
	user.ConfirmationCode = strings.TrimSpace(policy.Sanitize(user.ConfirmationCode))

	if user.Name == "" || user.Surname == "" || user.Email == "" || user.Address == "" || user.Phone == "" || user.Birthday  == "" ||
		user.Gender == "" || user.Web == "" || user.Bio  == "" || user.Username == "" || user.Password == ""{
		ctx.JSON(400, gin.H{"message" : "Field are empty or xss attack happened"})
		return
	}

	customValidator := validator2.NewCustomValidator()
	translator, _ := customValidator.RegisterEnTranslation()
	errValidation := customValidator.Validator.Struct(user)
	errs := customValidator.TranslateError(errValidation, translator)
	errorsString := customValidator.GetErrorsString(errs)

	if errValidation != nil {
		ctx.JSON(400, gin.H{"message" : errorsString[0]})
		return
	}


	if user.Birthday == "" {
		ctx.JSON(400, gin.H{"message" : "Enter birthday!"})
		return
	}

	if strings.Contains(user.Username, " ") {
		ctx.JSON(400, gin.H{"message" : "Username is not in valid format!"})
		return
	}


	if r.RegistrationUsecase.IsAlreadyRegistered(ctx, user.Username, user.Email) {
		ctx.JSON(402, gin.H{"message" : "User already exists"})
		return
	}
	if err := r.RegistrationUsecase.Register(ctx, user); err != nil {
		ctx.JSON(402, gin.H{"message" : err.Error()})
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


	policy := bluemonday.UGCPolicy()
	dto.Email = strings.TrimSpace(policy.Sanitize(dto.Email))

	dto.Code = strings.TrimSpace(policy.Sanitize(dto.Code))

	if dto.Code == "" || dto.Email == ""{
		ctx.JSON(400, gin.H{"message" : "Field are empty or xss attack happened"})
		return
	}

	if err := r.RegistrationUsecase.ConfirmAccount(ctx, dto.Code, dto.Email); err != nil {
		ctx.JSON(400, gin.H{"message" : err.Error()})
		return
	}




	ctx.JSON(http.StatusOK, gin.H{"message" : "Registration successful"})
	return
}

func (r *registrationHandler) ResendCode(ctx *gin.Context) {
	decoder := json.NewDecoder(ctx.Request.Body)

	type Email struct {
		Email	string	`json:"email"`
	}

	var req Email
	err := decoder.Decode(&req)


	policy := bluemonday.UGCPolicy()
	email := strings.TrimSpace(policy.Sanitize(req.Email))

	if err != nil {
		ctx.JSON(400, gin.H{"message" : "Field are empty or xss attack happened"})
		return
	}

	err, email, code := r.RegistrationUsecase.ResendCode(ctx, email)

	if err != nil {
		ctx.JSON(400, gin.H{"message" : "Invalid email"})
		return
	}


	usecase.SendMail(email, email, code)


	ctx.JSON(200, gin.H{"message" : "Resend request successful, please check your email"})
	return

}




