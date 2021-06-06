package main

import (
	"auth-service/src/http/middleware"
	router2 "auth-service/src/http/router"
	"auth-service/src/infrastructure/postgresqldb"
	"auth-service/src/infrastructure/seeder"
	interactor2 "auth-service/src/interactor"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

func main() {


	postgreConn := postgresqldb.NewDBConnection()

	seeder.SeedData(postgreConn)

	interactor := interactor2.NewInteractor(postgreConn)
	appHandler := interactor.NewAppHandler()

	router := router2.NewRouter(appHandler)
	router.Use(gin.Logger())
	router.Use(middleware.CORSMiddleware())

	redis := interactor.NewRedisUsecase()
	redis.AddKeyValueSet(context.Background(), "aaa", "111", time.Duration(1000000000000000000))
	value, _ := redis.GetValueByKey(context.Background(), "aaa")
	fmt.Println(value)



	router.RunTLS("localhost:8091", "certificate/cert.pem", "certificate/key.pem")
}
