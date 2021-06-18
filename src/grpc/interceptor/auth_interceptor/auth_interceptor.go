package auth_interceptor

import (
	"auth-service/infrastructure/tracer"
	"auth-service/usecase"
	"context"
	"fmt"
	"github.com/casbin/casbin/v2"
	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"os"
	"strings"
)

type authUnaryInterceptor struct {
	Authenticationusecase usecase.AuthenticationUsecase
}

type AuthUnaryInterceptor interface {
	UnaryAuthorizationInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error)
	ExtractUserRole(ctx context.Context, info *grpc.UnaryServerInfo) (string, error)
}

func NewAuthUnaryInterceptor(authenticationUsecase usecase.AuthenticationUsecase) AuthUnaryInterceptor {
	return &authUnaryInterceptor{authenticationUsecase}
}
func (a *authUnaryInterceptor) UnaryAuthorizationInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	role, err := a.ExtractUserRole(ctx, info)

	if err != nil {
		return nil, err
	}

	if role == "" {
		return nil, err
	}

	fullMethod := info.FullMethod
	ok, err := enforce(role, fullMethod, "*")

	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, err
	}
	log.Println("USAO U INTERCEPTOR: --> unary interceptor: ", fullMethod)
	return handler(ctx, req)
}


func (a *authUnaryInterceptor) ExtractUserRole(ctx context.Context, info *grpc.UnaryServerInfo) (string, error) {
	span := tracer.StartSpanFromContext(ctx, "middleware/ExtractUserRole")
	defer span.Finish()

	//ctx1 := tracer.ContextWithSpan(ctx, span)

	tokenString := a.ExtractToken(ctx, info)
	if tokenString == nil {
		return "ANONYMOUS", nil
	}

	token, err := jwt.Parse(*tokenString, func(token *jwt.Token) (interface{}, error) {
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

func (a *authUnaryInterceptor) ExtractToken(ctx context.Context, info *grpc.UnaryServerInfo) *string {
	span := tracer.StartSpanFromContext(ctx, "middleware/ExtractToken")
	defer span.Finish()

	headers, ok  := metadata.FromIncomingContext(ctx)

	if !ok {
		return nil
	}

	authHeaders := headers["authorization"]
	if len(authHeaders) != 1 {
		return nil
	}

	if authHeaders == nil {
		return nil
	}
	tokenId := authHeaders[0]

	token, err := a.FetchToken(ctx, tokenId, info)

	if err != nil {
		return nil
	}

	return token
}

func (a *authUnaryInterceptor) FetchToken(ctx context.Context, tokenId string, info *grpc.UnaryServerInfo) (*string, error){

	if info.FullMethod == "/Authentication/ValidateTotp" {
		token, err := a.Authenticationusecase.FetchTemporaryToken(ctx, tokenId)
		if err != nil {
			return nil, err
		}

		tokenStr := string(token)
		return &tokenStr, nil
	}

	token, err := a.Authenticationusecase.FetchAuthToken(ctx, tokenId)
	if err != nil {
		return nil, err
	}

	tokenStr := string(token)

	return &tokenStr, nil
}

func enforce(role string, obj string, act string) (bool, error) {
	m, _ := os.Getwd()

	if !strings.HasSuffix(m, "src")  {
		splits := strings.Split(m, "src")
		wd := splits[0] + "/src"
		if err := os.Chdir(wd); err != nil {
			return false, err
		}
	}

	enforcer, err := casbin.NewEnforcer("grpc/interceptor/auth_interceptor/rbac_model.conf", "grpc/interceptor/auth_interceptor/rbac_policy.csv")
	if err != nil {
		return false, fmt.Errorf("failed to load policy from DB: %w", err)
	}

	err = enforcer.LoadPolicy()
	if err != nil {
		return false, fmt.Errorf("failed to load policy from DB: %w", err)
	}

	ok, _ := enforcer.Enforce(role, obj, act)

	return ok, nil
}
