package handler

import (
	"auth-service/domain"
	"auth-service/infrastructure/dto"
	validator2 "auth-service/infrastructure/validator"
	"auth-service/usecase"
	"encoding/json"
	"github.com/gin-gonic/gin"
	logger "github.com/jelena-vlajkov/logger/logger"
	"github.com/microcosm-cc/bluemonday"
	"net/http"
	"strings"
	"unicode"
)

type registrationHandler struct {
	RegistrationUsecase usecase.RegistrationUsecase
	logger *logger.Logger

}


type RegistrationHandler interface {
	Register(ctx *gin.Context)
	ConfirmAccount(ctx *gin.Context)
	ResendCode(ctx *gin.Context)
	RegisterAgent(ctx *gin.Context)
	ValidateAgentAccount(ctx *gin.Context)
	GetAgentRequests(ctx *gin.Context)
	ConfirmAgentAccount(ctx *gin.Context)
}

func NewRegistrationHandler(registrationUsecase usecase.RegistrationUsecase, logger *logger.Logger) RegistrationHandler {
	return &registrationHandler{RegistrationUsecase: registrationUsecase, logger: logger}
}

func (r *registrationHandler) Register(ctx *gin.Context) {
	r.logger.Logger.Println("Handling REGISTER USER")
	decoder := json.NewDecoder(ctx.Request.Body)
	var user domain.User
	if err := decoder.Decode(&user); err != nil {
		r.logger.Logger.Errorf("error while decoding json, error: %v\n", err)
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
		r.logger.Logger.Errorf("error while verifying and validating registration fields\n")
		r.logger.Logger.Warnf("possible xss attack from IP address: %v\n", ctx.Request.Host)
		ctx.JSON(400, gin.H{"message" : "Fields are empty or xss attack happened"})
		return
	}

	customValidator := validator2.NewCustomValidator()
	translator, _ := customValidator.RegisterEnTranslation()
	errValidation := customValidator.Validator.Struct(user)
	errs := customValidator.TranslateError(errValidation, translator)
	errorsString := customValidator.GetErrorsString(errs)

	if errValidation != nil {
		r.logger.Logger.Errorf("error while validating fields, error: %v\n", errorsString[0])
		ctx.JSON(400, gin.H{"message" : errorsString[0]})
		return
	}

	if pasval1, pasval2, pasval3, pasval4 := verifyPassword(user.Password); pasval1 == false || pasval2 == false || pasval3 == false || pasval4 == false {
		r.logger.Logger.Errorf("error while veryfing password, error: password not matching pattern")
		ctx.JSON(400, gin.H{"message" : "Password must have minimum 1 uppercase letter, 1 lowercase letter, 1 digit and 1 special character and needs to be minimum 8 characters long"})
		return
	}


	if user.Birthday == "" {
		r.logger.Logger.Errorf("error while registrating user, error: no birthday")
		ctx.JSON(400, gin.H{"message" : "Enter birthday!"})
		return
	}

	if strings.Contains(user.Username, " ") {
		r.logger.Logger.Errorf("error while registrating user, error: username not in valid format")
		ctx.JSON(400, gin.H{"message" : "Username is not in valid format!"})
		return
	}


	if r.RegistrationUsecase.IsAlreadyRegistered(ctx, user.Username, user.Email) {
		r.logger.Logger.Errorf("error while registrating user, error: user already exists")
		ctx.JSON(402, gin.H{"message" : "User already exists"})
		return
	}
	if err := r.RegistrationUsecase.Register(ctx, user); err != nil {
		r.logger.Logger.Errorf("error while registrating user, error: %v\n", err)
		ctx.JSON(402, gin.H{"message" : err.Error()})
	}

	ctx.JSON(http.StatusOK, gin.H{"message" : "Please check your email to confirm registration"})
}

func (r *registrationHandler) ConfirmAccount(ctx *gin.Context) {
	r.logger.Logger.Println("Handling CONFIRMING ACCOUNT")
	decoder := json.NewDecoder(ctx.Request.Body)
	var dto dto.AccountConfirmationDto
	if err := decoder.Decode(&dto); err != nil {
		r.logger.Logger.Errorf("error while decoding json, error: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"message" : "Error decoding json"})
		return
	}


	policy := bluemonday.UGCPolicy()
	dto.Email = strings.TrimSpace(policy.Sanitize(dto.Email))

	dto.Code = strings.TrimSpace(policy.Sanitize(dto.Code))

	if dto.Code == "" || dto.Email == ""{
		r.logger.Logger.Errorf("error while verifying and validating registration fields\n")
		r.logger.Logger.Warnf("possible xss attack from IP address: %v\n", ctx.Request.Host)
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
	r.logger.Logger.Println("Handling RESEND CODE")
	decoder := json.NewDecoder(ctx.Request.Body)

	type Data struct {
		Email	string	`json:"email"`
		Username string `json:"username"`
	}

	var req Data
	err := decoder.Decode(&req)


	policy := bluemonday.UGCPolicy()
	email := strings.TrimSpace(policy.Sanitize(req.Email))
	if err != nil {
		r.logger.Logger.Errorf("error while verifying and validating registration fields, error: %v\n", err)
		r.logger.Logger.Warnf("possible xss attack from IP address: %v\n", ctx.Request.Host)
		ctx.JSON(400, gin.H{"message" : "Field are empty or xss attack happened"})
		return
	}

	err, email, code := r.RegistrationUsecase.ResendCode(ctx, email)

	if err != nil {
		ctx.JSON(400, gin.H{"message" : "Invalid email"})
		return
	}


	go usecase.SendMail(email, email, code)


	ctx.JSON(200, gin.H{"message" : "Resend request successful, please check your email"})
	return

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

func (r *registrationHandler) RegisterAgent(ctx *gin.Context) {
	r.logger.Logger.Println("Handling REGISTER USER")
	decoder := json.NewDecoder(ctx.Request.Body)
	var user domain.User
	if err := decoder.Decode(&user); err != nil {
		r.logger.Logger.Errorf("error while decoding json, error: %v\n", err)
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
		r.logger.Logger.Errorf("error while verifying and validating registration fields\n")
		r.logger.Logger.Warnf("possible xss attack from IP address: %v\n", ctx.Request.Host)
		ctx.JSON(400, gin.H{"message" : "Fields are empty or xss attack happened"})
		return
	}

	customValidator := validator2.NewCustomValidator()
	translator, _ := customValidator.RegisterEnTranslation()
	errValidation := customValidator.Validator.Struct(user)
	errs := customValidator.TranslateError(errValidation, translator)
	errorsString := customValidator.GetErrorsString(errs)

	if errValidation != nil {
		r.logger.Logger.Errorf("error while validating fields, error: %v\n", errorsString[0])
		ctx.JSON(400, gin.H{"message" : errorsString[0]})
		return
	}

	if pasval1, pasval2, pasval3, pasval4 := verifyPassword(user.Password); pasval1 == false || pasval2 == false || pasval3 == false || pasval4 == false {
		r.logger.Logger.Errorf("error while veryfing password, error: password not matching pattern")
		ctx.JSON(400, gin.H{"message" : "Password must have minimum 1 uppercase letter, 1 lowercase letter, 1 digit and 1 special character and needs to be minimum 8 characters long"})
		return
	}


	if user.Birthday == "" {
		r.logger.Logger.Errorf("error while registrating user, error: no birthday")
		ctx.JSON(400, gin.H{"message" : "Enter birthday!"})
		return
	}

	if strings.Contains(user.Username, " ") {
		r.logger.Logger.Errorf("error while registrating user, error: username not in valid format")
		ctx.JSON(400, gin.H{"message" : "Username is not in valid format!"})
		return
	}


	if r.RegistrationUsecase.IsAlreadyRegistered(ctx, user.Username, user.Email) {
		r.logger.Logger.Errorf("error while registrating user, error: user already exists")
		ctx.JSON(402, gin.H{"message" : "User already exists"})
		return
	}

	if err := r.RegistrationUsecase.RegisterAgent(ctx, user); err != nil {
		r.logger.Logger.Errorf("error while registrating user, error: %v\n", err)
		ctx.JSON(402, gin.H{"message" : err.Error()})
	}

	ctx.JSON(http.StatusOK, gin.H{"message" : "Please check your email to confirm registration"})
}

func (r *registrationHandler) ValidateAgentAccount(ctx *gin.Context) {

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

	if err := r.RegistrationUsecase.ValidateAgentAccount(ctx, dto.Code, dto.Email); err != nil {
		ctx.JSON(400, gin.H{"message" : err.Error()})
		return
	}


	ctx.JSON(http.StatusOK, gin.H{"message" : "Registration successful"})
	return
}

func (r *registrationHandler) GetAgentRequests(ctx *gin.Context) {

	keys, err := r.RegistrationUsecase.GetAgentRequests(ctx)

	if err != nil {
		ctx.JSON(400, gin.H{"message" : "error"})
		return
	}

	ctx.JSON(200, keys)
}

func (r *registrationHandler) ConfirmAgentAccount(ctx *gin.Context) {

	decoder := json.NewDecoder(ctx.Request.Body)
	var dto dto.ConfirmAgentAccountDto
	if err := decoder.Decode(&dto); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message" : "Error decoding json"})
		return
	}


	err := r.RegistrationUsecase.ConfirmAgentAccount(ctx, dto.Email, dto.Confirm)
	if err != nil {
		ctx.JSON(400, gin.H{"message" : "Error confirming account"})
		return
	}

	ctx.JSON(400, gin.H{"message" : "Account confirmed successfully"})
}