package main

import (
	"auth-service/grpc/server/authentication_server"
	"auth-service/http/middleware"
	router2 "auth-service/http/router"
	"auth-service/infrastructure/postgresqldb"
	"auth-service/infrastructure/seeder"
	interactor2 "auth-service/interactor"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	logger "github.com/jelena-vlajkov/logger/logger"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

func getNetListener(port uint) net.Listener {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		panic(fmt.Sprintf("failed to listen: %v", err))
	}

	return lis
}


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

	port := uint(8079)
	lis := getNetListener(port)
	grpcServer := grpc.NewServer()
	loginServiceImpl := interactor.NewAuthenticationServiceImpl()
	totpServiceImpl := interactor.NewTotpServiceImpl()

	authentication_server.RegisterAuthenticationServer(grpcServer, loginServiceImpl)
	authentication_server.RegisterTotpServer(grpcServer, totpServiceImpl)
	go func() {
		log.Fatalln(grpcServer.Serve(lis))
	}()

	logger.Logger.Info("server auth-service listening on port %s\n", port)

	router.RunTLS(":8091", "certificate/cert.pem", "certificate/key.pem")
}
