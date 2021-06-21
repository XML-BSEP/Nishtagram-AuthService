package main

import (
	"auth-service/grpc/interceptor/auth_interceptor"
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
	"os"
	"strconv"
	"time"
)

func getNetListener(port uint) net.Listener {
	var domain string
	if os.Getenv("DOCKER_ENV") == "" {
		domain = "127.0.0.1"
	} else {
		domain = "authms"
	}
	domain = domain + ":" + strconv.Itoa(int(port))
	lis, err := net.Listen("tcp", domain)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
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
	/*creds, err := credentials.NewServerTLSFromFile("src/certificate/cert.pem", "src/certificate/key.pem")
	if err != nil {
		panic(err)
	}*/

	a := auth_interceptor.NewAuthUnaryInterceptor(interactor.NewAuthenticationUsecase())

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(a.UnaryAuthorizationInterceptor))
	loginServiceImpl := interactor.NewAuthenticationServiceImpl()
	totpServiceImpl := interactor.NewTotpServiceImpl()

	authentication_server.RegisterAuthenticationServer(grpcServer, loginServiceImpl)
	authentication_server.RegisterTotpServer(grpcServer, totpServiceImpl)
	go func() {
		log.Fatalln(grpcServer.Serve(lis))
	}()

	logger.Logger.Info("server auth-service listening on port ", port)
	//logger.Logger.Info("server auth-service listening on port 8091")
	if os.Getenv("DOCKER_ENV") == "" {
		err := router.RunTLS(":8091", "src/certificate/cert.pem", "src/certificate/key.pem")
		if err != nil {
			return
		}
	} else {
		err := router.Run(":8091")
		if err != nil {
			return
		}
	}
}
