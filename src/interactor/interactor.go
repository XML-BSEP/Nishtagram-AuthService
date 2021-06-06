package interactor

import (
	"auth-service/http/handler"
	"auth-service/infrastructure/redisdb"
	"auth-service/repository"
	"auth-service/usecase"
	"gorm.io/gorm"
)

type interactor struct {
	Conn *gorm.DB
}

type Interactor interface {
	NewProfileInfoRepository() repository.ProfileInfoRepository
	NewRoleRepository() repository.RoleRepository


	NewRedisUsecase() usecase.RedisUsecase
	NewAuthenticationUsecase() usecase.AuthenticationUsecase
	NewJwtUsecase() usecase.JwtUsecase
	NewProfileInfoUsecase() usecase.ProfileInfoUsecase
	NewRegistrationUsecase() usecase.RegistrationUsecase

	NewAppHandler() AppHandler
	NewAuthenticationHandler() handler.AuthenticationHandler
	NewRegistrationHandler() handler.RegistrationHandler

}

type appHandler struct {
	handler.AuthenticationHandler
	handler.RegistrationHandler
}

type AppHandler interface {
	handler.AuthenticationHandler
	handler.RegistrationHandler
}

func NewInteractor(conn *gorm.DB) Interactor {
	return &interactor{Conn: conn}
}

func (i *interactor) NewAppHandler() AppHandler {
	appHandler := &appHandler{}
	appHandler.AuthenticationHandler = i.NewAuthenticationHandler()
	appHandler.RegistrationHandler = i.NewRegistrationHandler()
	return appHandler
}
func (i *interactor) NewProfileInfoRepository() repository.ProfileInfoRepository {
	return repository.NewProfileInfoRepository(i.Conn)
}

func (i *interactor) NewRoleRepository() repository.RoleRepository {
	return repository.NewRoleRepository(i.Conn)
}

func (i *interactor) NewRedisUsecase() usecase.RedisUsecase {
	redis := redisdb.NewReddisConn()
	return usecase.NewRedisUsecase(redis)
}

func (i *interactor) NewAuthenticationUsecase() usecase.AuthenticationUsecase {
	return usecase.NewAuthenticationUsecase(i.NewRedisUsecase())
}

func (i *interactor) NewJwtUsecase() usecase.JwtUsecase {
	return usecase.NewJwtUsecase(i.NewRedisUsecase())
}

func (i *interactor) NewAuthenticationHandler() handler.AuthenticationHandler {
	return handler.NewAuthenticationHandler(i.NewAuthenticationUsecase(), i.NewJwtUsecase(), i.NewProfileInfoUsecase())
}

func (i *interactor) NewProfileInfoUsecase() usecase.ProfileInfoUsecase {
	return usecase.NewProfileInfoUsecase(i.NewProfileInfoRepository())
}

func (i *interactor) NewRegistrationUsecase() usecase.RegistrationUsecase {
	return usecase.NewRegistrationUsecase(i.NewRedisUsecase(), i.NewProfileInfoUsecase())
}

func (i *interactor) NewRegistrationHandler() handler.RegistrationHandler {
	return handler.NewRegistrationHandler(i.NewRegistrationUsecase())
}
