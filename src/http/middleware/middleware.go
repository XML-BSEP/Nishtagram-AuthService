package middleware

import (
	"auth-service/infrastructure/tracer"
	"context"
	"fmt"
	"github.com/casbin/casbin/v2"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	logger "github.com/jelena-vlajkov/logger/logger"
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

func AuthMiddleware(logger *logger.Logger) gin.HandlerFunc {
	return func (c *gin.Context) {
		role, err := ExtractUserRole(context.Background(), c.Request, logger)
		if err != nil {
			logger.Logger.Warnf("unauthorized request from IP address: %v", c.Request.Host)
			c.JSON(401, gin.H{"message" : "Unauthorized"})
			c.Abort()
			return
		}

		if role == "" {
			logger.Logger.Warnf("unauthorized request from IP address: %v", c.Request.Host)
			c.JSON(401, gin.H{"message" : "Unauthorized"})
			c.Abort()
			return
		}

		ok, err := enforce(role, c.Request.URL.Path, c.Request.Method, logger)

		if err != nil {
			c.JSON(500, gin.H{"message" : "error occurred when authorizing user"})
			c.Abort()
			return
		}

		if !ok {
			logger.Logger.Errorf("forbidden request from IP address: %v", c.Request.Referer())
			c.JSON(403, gin.H{"message" : "forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}


func enforce(role string, obj string, act string, logger *logger.Logger) (bool, error) {
	m, _ := os.Getwd()
	fmt.Println(m)
	fmt.Println(role)

	if !strings.HasSuffix(m, "src")  {
		splits := strings.Split(m, "src")
		wd := splits[0] + "/src"
		fmt.Println(wd)
		os.Chdir(wd)
	}
	enforcer, err := casbin.NewEnforcer("http/middleware/rbac_model.conf", "http/middleware/rbac_policy.csv")
	if err != nil {
		logger.Logger.Errorf("failed to load policy from file: %v", err)
		return false, fmt.Errorf("failed to load policy from DB: %w", err)
	}
	err = enforcer.LoadPolicy()
	if err != nil {
		logger.Logger.Errorf("failed to load policy from file: %v", err)
		return false, fmt.Errorf("failed to load policy from DB: %w", err)
	}
	ok, _ := enforcer.Enforce(role, obj, act)
	return ok, nil
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
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2{
		return strArr[1]
	} else {
		if len(strArr) == 1 {
			if strArr[0] != "" {
				strArr2 := strings.Split(strArr[0], "\"")
				if len(strArr2) == 1 {
					return strArr2[0]
				}
				return strArr2[1]
			}
		}
	}
	return bearToken
}

func ExtractTokenFromCookie(ctx context.Context, r *http.Request) string {
	span := tracer.StartSpanFromContext(ctx, "middleware/ExtractTokenFromCookie")
	defer span.Finish()


	cookie := r.Header.Get("Cookie")
	tokens := strings.Split(cookie, "jwt=")
	if len(tokens) < 2 {
		tracer.LogError(span, fmt.Errorf("message= %s", "Token does not exists"))
		return ""
	}

	return tokens[1]
}

func ExtractUserIdFromCookie(ctx context.Context, r *http.Request) (string, error) {
	span := tracer.StartSpanFromContext(ctx, "middleware/ExtractUserIdFromCookie")
	defer span.Finish()

	ctx1 := tracer.ContextWithSpan(ctx, span)

	tokenString := ExtractTokenFromCookie(ctx1, r)

	if tokenString == "" {
		tracer.LogError(span, fmt.Errorf("message= %s", "Authorization header does noe exist"))
		return "", fmt.Errorf("", "message= %s", "Token does not exist")
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
func ExtractUserId(ctx context.Context, r *http.Request) (string, error) {
	span := tracer.StartSpanFromContext(ctx, "middleware/ExtractUserId")
	defer span.Finish()

	ctx1 := tracer.ContextWithSpan(ctx, span)

	tokenString := ExtractToken(ctx1, r)

	if tokenString == "" {
		tracer.LogError(span, fmt.Errorf("message= %s", "Authorization header does noe exist"))
		return "", fmt.Errorf("", "message= %s", "Authorization header does not exist")
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

func ExtractTokenUuid(ctx context.Context, r *http.Request) (string, error) {
	span := tracer.StartSpanFromContext(ctx, "middleware/ExtractTokenUuid")
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
		userId, ok := claims["token_uuid"].(string)
		if !ok {
			return "", err
		}

		return userId, nil
	}
	return "", err
}
func ExtractUserRole(ctx context.Context, r *http.Request, logger *logger.Logger) (string, error) {
	span := tracer.StartSpanFromContext(ctx, "middleware/ExtractUserRole")
	defer span.Finish()

	ctx1 := tracer.ContextWithSpan(ctx, span)

	tokenString := ExtractToken(ctx1, r)
	if tokenString == "" {
		return "ANONYMOUS", nil
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})

	claims, ok := token.Claims.(jwt.MapClaims)

	if ok  {
		userId, ok := claims["role"].(string)
		if !ok {
			return "ANONYMOUS", err
		}

		return strings.ToUpper(userId), nil
	}
	return  "ANONYMOUS", err
}


