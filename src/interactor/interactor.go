package interactor

import (
	"auth-service/gateway"
	"auth-service/grpc/server/authentication_server/implementation"
	"auth-service/http/handler"
	"auth-service/infrastructure/saga"
	"auth-service/infrastructure/tracer"
	"auth-service/repository"
	"auth-service/usecase"
	"github.com/go-redis/redis/v8"
	logger "github.com/jelena-vlajkov/logger/logger"
	"io"

	totp_implementation "auth-service/grpc/server/2fa_server/implementation"
	resty2 "github.com/go-resty/resty/v2"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"
)

const tracing_name = "auth_service"

type interactor struct {
	Conn   *gorm.DB
	Tracer opentracing.Tracer
	Closer io.Closer
	logger *logger.Logger
	RedisClient *redis.Client
	SagaRedisClient *redis.Client
	Orchestrator saga.Orchestrator
}

type Interactor interface {
	NewProfileInfoRepository() repository.ProfileInfoRepository
	NewRoleRepository() repository.RoleRepository
	NewTotpRepository() repository.TotpRepository

	NewRedisUsecase() usecase.RedisUsecase
	NewAuthenticationUsecase() usecase.AuthenticationUsecase
	NewJwtUsecase() usecase.JwtUsecase
	NewProfileInfoUsecase() usecase.ProfileInfoUsecase
	NewRegistrationUsecase() usecase.RegistrationUsecase
	NewTotpUsecase() usecase.TotpUsecase

	NewAppHandler() AppHandler
	NewAuthenticationHandler() handler.AuthenticationHandler
	NewRegistrationHandler() handler.RegistrationHandler
	NewTotpHandler() handler.TotpHandler

	NewUserGateway() gateway.UserGateway

	NewAuthenticationServiceImpl() *implementation.AuthenticationServer
	NewTotpServiceImpl() *totp_implementation.TotpServer
}

type appHandler struct {
	handler.AuthenticationHandler
	handler.RegistrationHandler
	handler.TotpHandler
}

type AppHandler interface {
	handler.AuthenticationHandler
	handler.RegistrationHandler
	handler.TotpHandler
}

func NewInteractor(conn *gorm.DB, logger *logger.Logger, redisClient *redis.Client, sagaRedisClient *redis.Client, orchestrator saga.Orchestrator) Interactor {
	tracer, closer := tracer.Init(tracing_name)
	opentracing.SetGlobalTracer(tracer)
	return &interactor{
		Conn:   conn,
		Tracer: tracer,
		Closer: closer,
		logger: logger,
		RedisClient: redisClient,
		SagaRedisClient: sagaRedisClient,
		Orchestrator: orchestrator,
	}
}

func (i *interactor) NewAppHandler() AppHandler {
	appHandler := &appHandler{}
	appHandler.AuthenticationHandler = i.NewAuthenticationHandler()
	appHandler.RegistrationHandler = i.NewRegistrationHandler()
	appHandler.TotpHandler = i.NewTotpHandler()
	return appHandler
}
func (i *interactor) NewProfileInfoRepository() repository.ProfileInfoRepository {
	return repository.NewProfileInfoRepository(i.Conn, i.logger)
}

func (i *interactor) NewRoleRepository() repository.RoleRepository {
	return repository.NewRoleRepository(i.Conn, i.logger)
}

func (i *interactor) NewRedisUsecase() usecase.RedisUsecase {
	return usecase.NewRedisUsecase(i.RedisClient, i.logger)
}

func (i *interactor) NewAuthenticationUsecase() usecase.AuthenticationUsecase {
	return usecase.NewAuthenticationUsecase(i.NewRedisUsecase(), i.logger)
}

func (i *interactor) NewJwtUsecase() usecase.JwtUsecase {
	return usecase.NewJwtUsecase(i.NewRedisUsecase(), i.logger, i.NewAuthenticationUsecase())
}

func (i *interactor) NewAuthenticationHandler() handler.AuthenticationHandler {

	return handler.NewAuthenticationHandler(i.NewAuthenticationUsecase(), i.NewJwtUsecase(), i.NewProfileInfoUsecase(), i.Tracer, i.NewRedisUsecase(), i.NewTotpUsecase(), i.logger)
}

func (i *interactor) NewProfileInfoUsecase() usecase.ProfileInfoUsecase {
	return usecase.NewProfileInfoUsecase(i.NewProfileInfoRepository(), i.NewRedisUsecase(), i.logger)
}

func (i *interactor) NewRegistrationUsecase() usecase.RegistrationUsecase {
	return usecase.NewRegistrationUsecase(i.NewRedisUsecase(), i.NewProfileInfoUsecase(), i.NewUserGateway(), i.logger)
}

func (i *interactor) NewRegistrationHandler() handler.RegistrationHandler {
	return handler.NewRegistrationHandler(i.NewRegistrationUsecase(), i.logger, i.Orchestrator)
}

func (i *interactor) NewUserGateway() gateway.UserGateway {
	resty := resty2.New()
	return gateway.NewUserGateway(resty)
}

func (i *interactor) NewTotpRepository() repository.TotpRepository {
	return repository.NewTotpRepository(i.Conn, i.logger)
}

func (i *interactor) NewTotpUsecase() usecase.TotpUsecase {
	return usecase.NewTotpUsecase(i.NewTotpRepository(), i.NewRedisUsecase(), i.NewProfileInfoUsecase(), i.logger)
}

func (i *interactor) NewTotpHandler() handler.TotpHandler {
	return handler.NewTotpHandler(i.NewTotpUsecase(), i.Tracer, i.NewProfileInfoUsecase(), i.logger)
}

func (i *interactor) NewAuthenticationServiceImpl() *implementation.AuthenticationServer {
	return implementation.NewAuthenticationServiceImpl(i.NewProfileInfoUsecase(), i.NewTotpUsecase(), i.NewJwtUsecase(), i.NewAuthenticationUsecase(), i.NewRedisUsecase())
}

func (i *interactor) NewTotpServiceImpl() *totp_implementation.TotpServer {
	return totp_implementation.NewTotpServer(i.NewTotpUsecase(), i.NewProfileInfoUsecase())
}
