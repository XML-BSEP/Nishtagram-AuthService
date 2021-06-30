package router

import (
	"auth-service/http/middleware"
	"auth-service/interactor"
	logger "github.com/jelena-vlajkov/logger/logger"

	"github.com/gin-gonic/gin"
)

func NewRouter(handler interactor.AppHandler, logger *logger.Logger) *gin.Engine {
	router := gin.Default()
	router.Use(middleware.AuthMiddleware(logger))

	router.POST("/validateToken", handler.ValidateToken)
	router.POST("/login", handler.Login)
	router.POST("/logout", handler.Logout)
	router.POST("/register", handler.Register)
	router.POST("/confirmAccount", handler.ConfirmAccount)
	router.GET("/generateSecret", handler.GenerateSecret)
	router.POST("/verifySecret", handler.Verify)
	router.POST("/isTotpEnabled", handler.IsEnabled)
	router.POST("/disableTotp", handler.Disable)
	router.POST("/validateTotp", handler.Validate)
	router.POST("/validateTemporaryToken", handler.ValidateTemporaryToken)
	router.POST("/resendRegistrationCode", handler.ResendCode)
	router.POST("/resetPasswordMail", handler.SendResetMail)
	router.POST("/resetPassword", handler.ResetPassword)
	router.POST("refreshToken", handler.RefreshToken)

	router.POST("/agent", handler.RegisterAgent)
	router.POST("/agent/validate", handler.ValidateAgentAccount)
	router.GET("/agent", handler.GetAgentRequests)
	router.POST("/confirmAgentAccount", handler.ConfirmAgentAccount)

	return router
}
