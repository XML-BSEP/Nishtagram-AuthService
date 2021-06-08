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



	router.RunTLS(":8091", "src/certificate/cert.pem", "src/certificate/key.pem")
}
