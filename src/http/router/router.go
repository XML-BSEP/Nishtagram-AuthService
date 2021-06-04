package router

import (
	"auth-service/src/interactor"
	"github.com/gin-gonic/gin"
)

func NewRouter(handler interactor.AppHandler) *gin.Engine {
	router := gin.Default()

	router.POST("/validateToken", handler.ValidateToken)
	router.POST("/login", handler.Login)
	router.POST("/logout", handler.Logout)

	return router
}


