package gateway

import (
	"auth-service/src/domain"
	"context"
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
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
	if err != nil {
		return err
	}
	response, err := u.RestyClient.R().SetBody(json).Post("https://localhost:8082/saveNewUser")
	if response.StatusCode() != 200 {
		return errors.New("User failed")
	}

	return nil
}
