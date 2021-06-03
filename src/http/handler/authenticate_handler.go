package handler

import (
	"auth-service/src/infrastructure/dto"
	"auth-service/src/usecase"
	"encoding/json"
	"github.com/gin-gonic/gin"
)

type authenticateHandler struct {
	AuthenticationUsecase usecase.AuthenticationUsecase
	JwtUsecase usecase.JwtUsecase
}

type AuthenticationHandler interface {
	Login(ctx *gin.Context)
	ValidateToken(ctx *gin.Context)
	Logout(ctx *gin.Context)
}

func NewAuthenticationHandler(authUsecase usecase.AuthenticationUsecase, jwtUSecase usecase.JwtUsecase) AuthenticationHandler {
	return &authenticateHandler{authUsecase, jwtUSecase}
}

func (a *authenticateHandler) Login(ctx *gin.Context) {
	token, err := a.JwtUsecase.CreateToken(ctx, 12)
	if err != nil {
		ctx.JSON(400, gin.H{"message" : "Can not create token"})
		ctx.Abort()
		return
	}

	a.AuthenticationUsecase.SaveAuthToken(ctx, 12, token)

	ctx.JSON(200, gin.H{"token_id" : token.TokenUuid})
}

func (a *authenticateHandler) ValidateToken(ctx *gin.Context) {
	var tokenDto dto.TokenDto

	decoder := json.NewDecoder(ctx.Request.Body)

	if err := decoder.Decode(&tokenDto); err != nil {
		ctx.JSON(400, gin.H{"message" : "Token decoding error"})
		ctx.Abort()
		return
	}

	at, err := a.AuthenticationUsecase.FetchAuthToken(ctx, tokenDto.TokenId)

	if err != nil {
		ctx.JSON(401, gin.H{"message" : "Invalid token"})
		ctx.Abort()
		return
	}

	if err := a.JwtUsecase.ValidateToken(ctx, at); err != nil {
		ctx.JSON(401, gin.H{"message" : "Invalid token"})
		ctx.Abort()
		return
	}

	ctx.JSON(200, gin.H{"message" : "Token is valid"})

}

func (a *authenticateHandler) Logout(ctx *gin.Context) {
	var tokenDto dto.TokenDto

	decoder := json.NewDecoder(ctx.Request.Body)

	if err := decoder.Decode(&tokenDto); err != nil {
		ctx.JSON(400, gin.H{"message" : "Token decoding error"})
		ctx.Abort()
		return
	}

	if err := a.AuthenticationUsecase.DeleteAuthToken(ctx, tokenDto.TokenId); err != nil {
		ctx.JSON(400, gin.H{"message" : "Token deleting error"})
		ctx.Abort()
		return
	}

	ctx.JSON(200, gin.H{"message" : "Sucessful logout"})
}