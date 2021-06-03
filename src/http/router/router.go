package router

import (
	"auth-service/src/interactor"
	"github.com/gin-gonic/gin"
)

func NewRouter(handler interactor.AppHandler) *gin.Engine{
	router := gin.Default()
	router.Use(CORSMiddleware())
	router.POST("/validateToken", handler.ValidateToken)
	router.POST("/login", handler.Login)
	router.POST("/logout", handler.Logout)

	return router
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Header("Access-Control-Allow-Origin", "http://localhost:8080")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH, OPTIONS, GET, PUT")


		if c.Request.Method == "OPTIONS" {
			c.JSON(204, gin.H{})
			return
		}

		c.Next()
	}
}
