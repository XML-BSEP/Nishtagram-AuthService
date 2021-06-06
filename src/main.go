package main

import (
	"auth-service/http/middleware"
	router2 "auth-service/http/router"
	"auth-service/infrastructure/postgresqldb"
	"auth-service/infrastructure/seeder"
	interactor2 "auth-service/interactor"
	"github.com/gin-gonic/gin"
)

func main() {


	postgreConn := postgresqldb.NewDBConnection()

	seeder.SeedData(postgreConn)

	interactor := interactor2.NewInteractor(postgreConn)
	appHandler := interactor.NewAppHandler()

	router := router2.NewRouter(appHandler)
	router.Use(gin.Logger())
	router.Use(middleware.CORSMiddleware())



	router.RunTLS(":8091", "certificate/cert.pem", "certificate/key.pem")

}
