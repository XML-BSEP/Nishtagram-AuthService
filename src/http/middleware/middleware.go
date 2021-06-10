package middleware

import (
	"auth-service/infrastructure/tracer"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

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

func GetTokenId(ctx context.Context, request *http.Request) *string {
	span := tracer.StartSpanFromContext(ctx, "middleware/GetTokenId")
	defer span.Finish()
	authHeader := request.Header.Get("authorization")
	if authHeader == "" {
		tracer.LogError(span, fmt.Errorf("", "Cookie header does not exist"))
		return nil
	}

	return &authHeader

}
