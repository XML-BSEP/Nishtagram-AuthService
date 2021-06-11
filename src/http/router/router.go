package router

import (
	"auth-service/interactor"
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
)

func NewRouter(handler interactor.AppHandler) *gin.Engine {
	router := gin.Default()

	router.Use(secure.New(secure.DefaultConfig()))

	router.POST("/validateToken", handler.ValidateToken)
	router.POST("/login", handler.Login)
	router.POST("/logout", handler.Logout)
	router.POST("/register", handler.Register)
	router.POST("/confirmAccount", handler.ConfirmAccount)
	router.POST("/generateSecret", handler.GenerateSecret)
	router.POST("/verifySecret", handler.Verify)
	router.GET("/isTotpEnabled", handler.IsEnabled)
	router.POST("/disableTotp", handler.Disable)
	router.POST("/resendRegistrationCode", handler.ResendCode)


	return router
}


