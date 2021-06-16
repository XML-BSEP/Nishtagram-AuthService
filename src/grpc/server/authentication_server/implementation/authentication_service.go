package implementation

import (
	"auth-service/domain"
	helper2 "auth-service/grpc/helper"
	pb "auth-service/grpc/server/authentication_server"
	"auth-service/helper"
	"auth-service/usecase"
	"context"
	"fmt"
	"github.com/microcosm-cc/bluemonday"
	"strings"
	"time"
)

const (
	redisKeyPattern = "passwordResetRequest"
)


type AuthenticationServer struct {
	pb.UnimplementedAuthenticationServer
	ProfileInfoUsecase usecase.ProfileInfoUsecase
	TotpUsecase usecase.TotpUsecase
	JwtUsecase usecase.JwtUsecase
	AuthenticationUsecase usecase.AuthenticationUsecase
	RedisUsecase usecase.RedisUsecase
}



func NewAuthenticationServiceImpl(profileInfoUsecase usecase.ProfileInfoUsecase, totpUsecase usecase.TotpUsecase, jwtUsecase usecase.JwtUsecase, authenticationUsecase usecase.AuthenticationUsecase, redisUsecase usecase.RedisUsecase) *AuthenticationServer {
	return &AuthenticationServer{ProfileInfoUsecase: profileInfoUsecase, TotpUsecase: totpUsecase, JwtUsecase: jwtUsecase, AuthenticationUsecase: authenticationUsecase, RedisUsecase: redisUsecase}
}

func (s *AuthenticationServer) Login(ctx context.Context, in *pb.LoginCredentials) (*pb.LoginResponse, error) {


	policy := bluemonday.UGCPolicy()
	in.Username = strings.TrimSpace(policy.Sanitize(in.Username))
	in.Password = strings.TrimSpace(policy.Sanitize(in.Password))

	profileInfo, err := s.ProfileInfoUsecase.GetProfileInfoByUsername(ctx, in.Username)

	if err != nil {
		return nil, err
	}

	if err := usecase.VerifyPassword(ctx, in.Password, profileInfo.Password); err != nil {
		return nil, err
	}

	_, err = s.TotpUsecase.GetSecretByProfileInfoId(ctx, profileInfo.ID)

	if err == nil {
		userInfo, err := s.CreateTemporaryToken(ctx, profileInfo)
		if err != nil {
			return nil, err
		}
		userInfo.Role = "temporary_user"

		return userInfo, err
	}

	val, err := s.generateToken(ctx, profileInfo, true)


	if err != nil {
		return nil, err
	}

	return val, nil
}

func (s *AuthenticationServer) Logout(ctx context.Context, in *pb.Tokens) (*pb.BooleanResponse, error) {

	if err := s.AuthenticationUsecase.DeleteAuthToken(ctx, in.Token); err != nil {
		return &pb.BooleanResponse{Success: false}, err
	}

	if err := s.AuthenticationUsecase.DeleteRefreshToken(ctx, in.RefreshToken); err != nil {
		return &pb.BooleanResponse{Success: false}, err
	}

	return &pb.BooleanResponse{Success: true}, nil

}

func (s *AuthenticationServer) ValidateToken(ctx context.Context, in *pb.Tokens) (*pb.TokenValidationResponse, error) {

	policy := bluemonday.UGCPolicy()
	in.Token = strings.TrimSpace(policy.Sanitize(in.Token))
	in.RefreshToken = strings.TrimSpace(policy.Sanitize(in.RefreshToken))

	at, errAt := s.AuthenticationUsecase.FetchAuthToken(ctx, in.Token)

	token, errAtValidation := s.JwtUsecase.ValidateToken(ctx, string(at))

	//If token exists, but is not valid
	if errAt == nil && errAtValidation != nil {
		return nil, errAt
	}

	newToken := &pb.TokenValidationResponse{
		AccessToken: token,
		AccessTokenUuid: in.Token,
		RefreshTokenUuid: in.RefreshToken,
	}

	var errRt error
	if errAt != nil || token == "" {

		newToken, errRt = s.RefreshToken(ctx, in.RefreshToken)
		if errRt != nil {
			return nil, errRt
		}
	}

	return newToken, nil
}

func (s *AuthenticationServer) ResendEmail(ctx context.Context, in *pb.ResendEmailRequest) (*pb.BooleanResponse, error) {


	policy := bluemonday.UGCPolicy()
	in.Email = strings.TrimSpace(policy.Sanitize(in.Email))


	exists := s.ProfileInfoUsecase.ExistsByUsernameOrEmail(ctx, "", in.Email)
	if !exists {
		return nil, fmt.Errorf("User does not exists")
	}

	user, _ := s.ProfileInfoUsecase.GetProfileInfoByEmail(ctx, in.Email)

	code := helper.RandomStringGenerator(8)

	expiration := 1000000000 * 3600 * 2 //2h

	hash, _ := helper.Hash(code)

	go usecase.SendRestartPasswordMail(user.Email, code)

	redisKey := redisKeyPattern + user.Email

	err := s.RedisUsecase.AddKeyValueSet(ctx, redisKey, hash, time.Duration(expiration))
	if err != nil {
		return nil, err
	}

	return &pb.BooleanResponse{Success: true}, nil
}
func (s *AuthenticationServer) CreateTemporaryToken(ctx context.Context, profileInfo domain.ProfileInfo) (*pb.LoginResponse, error) {



	temporaryToken, err := s.JwtUsecase.CreateTemporaryToken(ctx, profileInfo.Role.RoleName, profileInfo.ID)

	if err != nil {
		return nil, err
	}

	loginResponse := &pb.LoginResponse{
		AccessToken: temporaryToken.TokenUuid,
		Role:  profileInfo.Role.RoleName,
		Id:    profileInfo.ID,
	}

	if err := s.AuthenticationUsecase.SaveTemporaryToken(ctx, temporaryToken); err != nil {

		return nil, err
	}



	return loginResponse, nil
}

func (s *AuthenticationServer) generateToken(ctx context.Context, profileInfo domain.ProfileInfo, refresh bool) (*pb.LoginResponse, error) {


	token, err := s.JwtUsecase.CreateToken(ctx, profileInfo.Role.RoleName, profileInfo.ID, true)
	if err != nil {

		return nil, err
	}
	authenticatedUserInfo := pb.LoginResponse{
		AccessToken: token.TokenUuid,
		Role:  profileInfo.Role.RoleName,
		Id:    profileInfo.ID,
		RefreshToken: token.RefreshUuid,
	}

	fmt.Print("KREIRAN TOKEN ID: " + authenticatedUserInfo.AccessToken)
	return &authenticatedUserInfo, nil
}

func (s *AuthenticationServer) ValidateTemporaryToken(ctx context.Context, in *pb.AccessToken) (*pb.AccessToken, error) {

	at, err := s.AuthenticationUsecase.FetchTemporaryToken(ctx, in.AccessToken)

	if err != nil {
		return nil, err
	}

	token, err := s.JwtUsecase.ValidateToken(ctx, string(at))

	if err != nil{
		return nil, err
	}
	if token == "" {
		return nil, fmt.Errorf("")
	}


	return &pb.AccessToken{AccessToken: token}, nil

}

func (s *AuthenticationServer) ValidateTotp(ctx context.Context, in *pb.TotpValidation) (*pb.LoginResponse, error) {

	at, err := s.AuthenticationUsecase.FetchTemporaryToken(ctx, in.AccessToken.AccessToken)

	if err != nil {
		return nil, err
	}

	token, err := s.JwtUsecase.ValidateToken(ctx, string(at))

	if err != nil{
		return nil, err
	}
	if token == "" {
		return nil, fmt.Errorf("")
	}

	userId, err := helper2.ExtractUserIdFromToken(token)

	if err != nil {
		return nil, err
	}

	if !s.TotpUsecase.Validate(ctx, *userId, in.Passcode) {
		return nil, err
	}

	profileInfo, err := s.ProfileInfoUsecase.GetProfileInfoById(ctx, *userId)
	//TODO: add refresh bool in structure
	val, err := s.generateToken(ctx, *profileInfo, true)

	if err != nil {
		return nil, err
	}

	tokenUuid, err := helper2.ExtractTokenUuid(token)

	if err != nil {
		return nil, err
	}

	if err := s.AuthenticationUsecase.DeleteTemporaryToken(ctx, *tokenUuid); err != nil {
		return nil, err
	}

	return val, nil
}

func (s *AuthenticationServer) RefreshToken(context context.Context, refreshTokenUuid string) (*pb.TokenValidationResponse, error) {


	rt, err := s.AuthenticationUsecase.FetchRefreshToken(context, refreshTokenUuid)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.JwtUsecase.ValidateToken(context, string(rt))

	if err != nil || refreshToken == "" {
		return nil, err
	}

	token, uuid, err := s.JwtUsecase.RefreshToken(context, refreshToken)
	if err != nil {
		return nil, err
	}

	ret := &pb.TokenValidationResponse{
		AccessToken: *token,
		AccessTokenUuid: *uuid,
		RefreshTokenUuid: refreshTokenUuid,
	}

	return ret, nil
}
