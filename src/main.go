package main

import (
	"auth-service/http/middleware"
	router2 "auth-service/http/router"
	"auth-service/infrastructure/postgresqldb"
	"auth-service/infrastructure/seeder"
	interactor2 "auth-service/interactor"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	logger "github.com/jelena-vlajkov/logger/logger"
	"time"
)

func main() {

	logger := logger.InitializeLogger("auth-service", context.Background())
	postgreConn := postgresqldb.NewDBConnection(logger)

	seeder.SeedData(postgreConn)

	interactor := interactor2.NewInteractor(postgreConn, logger)
	appHandler := interactor.NewAppHandler()

	router := router2.NewRouter(appHandler, logger)
	router.Use(gin.Logger())
	router.Use(middleware.CORSMiddleware())

	redis := interactor.NewRedisUsecase()
	redis.AddKeyValueSet(context.Background(), "aaa", "111", time.Duration(1000000000000000000))
	value, _ := redis.GetValueByKey(context.Background(), "aaa")
	fmt.Println(value)


	logger.Logger.Info("server auth-service listening on port 8091")
	router.RunTLS(":8091", "src/certificate/cert.pem", "src/certificate/key.pem")
}
