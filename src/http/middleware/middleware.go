package middleware

import (
	"auth-service/infrastructure/tracer"
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strings"
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

func ExtractToken(ctx context.Context, r *http.Request) string {
	span := tracer.StartSpanFromContext(ctx, "middleware/ExtractToken")
	defer span.Finish()


	bearToken := r.Header.Get("Authorization")
	if bearToken == "" {
		tracer.LogError(span, fmt.Errorf("message= %s", "Authorization header does noe exist"))
		return ""
	}
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2{
		return strArr[1]
	} else {
		if len(strArr) == 1 {
			if strArr[0] != "" {
				strArr2 := strings.Split(strArr[0], "\"")

				return strArr2[1]
			}
		}
	}
	return ""
}

func ExtractUserId(ctx context.Context, r *http.Request) (string, error) {
	span := tracer.StartSpanFromContext(ctx, "middleware/ExtractToken")
	defer span.Finish()

	ctx1 := tracer.ContextWithSpan(ctx, span)

	tokenString := ExtractToken(ctx1, r)

	if tokenString == "" {
		tracer.LogError(span, fmt.Errorf("message= %s", "Authorization header does noe exist"))
		return "", fmt.Errorf("", "message= %s", "Authorization header does noe exist")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})

	claims, ok := token.Claims.(jwt.MapClaims)

	if ok  {
		userId, ok := claims["user_id"].(string)
		if !ok {
			return "", err
		}

		return userId, nil
	}
	return "", err
}

func ExtractUserRole(ctx context.Context, r *http.Request) (string, error) {
	span := tracer.StartSpanFromContext(ctx, "middleware/ExtractUserRole")
	defer span.Finish()

	ctx1 := tracer.ContextWithSpan(ctx, span)

	tokenString := ExtractToken(ctx1, r)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})

	claims, ok := token.Claims.(jwt.MapClaims)

	if ok  {
		userId, ok := claims["role"].(string)
		if !ok {
			return "", err
		}

		return userId, nil
	}
	return "", err
}