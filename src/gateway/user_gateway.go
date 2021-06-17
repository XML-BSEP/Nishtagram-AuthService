package gateway

import (
	"auth-service/domain"
	"context"
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"os"
)

type userGateway struct {
	RestyClient *resty.Client
}

type UserGateway interface {
	SaveRegisteredUser(context context.Context, user *domain.User) error
}

func NewUserGateway(resty *resty.Client) UserGateway {
	return &userGateway{RestyClient: resty}
}

func (u *userGateway) SaveRegisteredUser(context context.Context, user *domain.User) error {
	json, err := json.Marshal(user)
	domain := os.Getenv("USER_DOMAIN")
	if domain == "" {
		domain = "127.0.0.1"
	}
	if err != nil {
		return err
	}
	response, err := u.RestyClient.R().SetBody(json).Post("https://" + domain + ":8082/saveNewUser")

	if response.StatusCode() != 200 {
		return errors.New("User failed")
	}

	return nil
}
