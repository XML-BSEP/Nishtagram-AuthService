package saga

import (
	"auth-service/usecase"
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"log"
)

type authSaga struct {
	profileInfoUsecase usecase.ProfileInfoUsecase
	registartionUsecase usecase.RegistrationUsecase
	redisClient *redis.Client
}


type AuthSaga interface {
	SagaAuth(context context.Context)
}

func NewAuthSaga(profileInfoUsecase usecase.ProfileInfoUsecase, registrationUsecase usecase.RegistrationUsecase, redisClient *redis.Client) AuthSaga {
	return &authSaga{
		profileInfoUsecase: profileInfoUsecase,
		registartionUsecase: registrationUsecase,
		redisClient: redisClient,
	}
}


func (a *authSaga) SagaAuth(context context.Context) {

	pubsub := a.redisClient.Subscribe(context, AuthChannel, ReplyChannel)
	if _, err := pubsub.Receive(context); err != nil {
		log.Fatalf("error subscribing %s", err)
	}

	defer func() { _ = pubsub.Close() }()
	ch := pubsub.Channel()

	for {
		select {
		case msg := <- ch:
			m := Message{}
			err := json.Unmarshal([]byte(msg.Payload), &m)
			if err != nil {
				log.Println(err)
				continue
			}

			switch msg.Channel{
			case AuthChannel:
				if m.Action == ActionStart {
					user := m.Payload
					newUser, err := a.registartionUsecase.ConfirmAgentAccount(context, user.Email, m.Confirm)
					m.Payload = *newUser
					if err != nil {
						break
					}

					sendToReplyChannel(context, a.redisClient, m, ActionDone, AuthService, UserService)

				}
				if m.Action == ActionRollback {
					user := m.Payload
					if err := a.profileInfoUsecase.DeleteProfileInfo(context, user.Username); err != nil {
						break
					}
				}
			}
		}
	}
}

func sendToReplyChannel(context context.Context, client *redis.Client, m Message, action string, service string, senderService string) {
	var err error
	m.Action = action
	m.Service = service
	m.SenderService = senderService
	if err = client.Publish(context, ReplyChannel, m).Err(); err != nil {
		log.Printf("error publishing done-message to %s channel", ReplyChannel)
	}
	log.Printf("done message published to channel :%s", ReplyChannel)
}

