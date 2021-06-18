package implementation

import (
	pb "auth-service/grpc/server/authentication_server"
	"auth-service/usecase"
	"context"
	"fmt"
	"github.com/microcosm-cc/bluemonday"
	"strings"
)

type TotpServer struct {
	pb.UnimplementedTotpServer
	TotpUsecase usecase.TotpUsecase
	ProfileInfoUsecase usecase.ProfileInfoUsecase
}

func NewTotpServer (totpUsecase usecase.TotpUsecase, profileInfoUsecase usecase.ProfileInfoUsecase) *TotpServer {
	return &TotpServer{TotpUsecase: totpUsecase, ProfileInfoUsecase: profileInfoUsecase}
}

func(t *TotpServer) Verify(ctx context.Context, in *pb.TotpSecret) (*pb.BoolWrapper, error) {

	policy := bluemonday.UGCPolicy()

	in.Passcode = strings.TrimSpace(policy.Sanitize(in.Passcode))
	in.UserId = strings.TrimSpace(policy.Sanitize(in.UserId))

	if !t.TotpUsecase.Verify(ctx, in.Passcode, in.UserId) {
		return nil, fmt.Errorf("")
	}

	if err := t.TotpUsecase.SaveSecret(ctx, in.UserId); err != nil {

		return nil, err
	}

	return &pb.BoolWrapper{Value: true}, nil
}

func(t *TotpServer) IsEnabled(ctx context.Context, in *pb.Username) (*pb.BoolWrapper, error) {
	policy := bluemonday.UGCPolicy()
	in.Username = strings.TrimSpace(policy.Sanitize(in.Username))

	secret, err := t.TotpUsecase.GetSecretByProfileInfoId(ctx, in.Username)

	if err != nil {
		return nil, fmt.Errorf("")
	}

	if secret != nil {
		return &pb.BoolWrapper{Value: true}, nil
	}

	return &pb.BoolWrapper{Value: false}, nil
}

func(t *TotpServer) Disable(ctx context.Context, in *pb.TotpSecret) (*pb.BoolWrapper, error) {

	policy := bluemonday.UGCPolicy()
	in.Passcode = strings.TrimSpace(policy.Sanitize(in.Passcode))
	in.UserId = strings.TrimSpace(policy.Sanitize(in.UserId))

	if !t.TotpUsecase.Validate(ctx, in.UserId, in.Passcode) {
		return nil, fmt.Errorf("")
	}

	if err := t.TotpUsecase.DeleteSecretByProfileId(ctx, in.UserId); err != nil {
		return nil, err
	}

	return &pb.BoolWrapper{Value: true}, nil
}

