package main

import (
	"auth-service/src/http/middleware"
	router2 "auth-service/src/http/router"
	"auth-service/src/infrastructure/postgresqldb"
	"auth-service/src/infrastructure/seeder"
	interactor2 "auth-service/src/interactor"
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



	router.RunTLS("localhost:8091", "certificate/cert.pem", "certificate/key.pem")
}
