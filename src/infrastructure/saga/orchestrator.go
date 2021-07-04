package saga

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"log"
)

const (
	AuthChannel string = "AuthChannel"
	UserChannel string = "UserChannel"
	ReplyChannel string = "ReplyChannel"
	AuthService string = "Auth"
	UserService string = "User"
	ActionStart string = "Start"
	ActionDone string = "DoneMsg"
	ActionError string = "ErrorMsg"
	ActionRollback string = "RollbackMsg"
)

type orchestrator struct {
	redisClient *redis.Client
	redisPubSub *redis.PubSub
}


type Orchestrator interface {
	Start(context context.Context)
	Rollback(context context.Context, m Message)
	Next(context context.Context, channel string, service string, m Message)
}
func NewOrchestrator(context context.Context, c *redis.Client) Orchestrator {

	return &orchestrator{
		redisClient: c,
		redisPubSub: c.Subscribe(context, AuthChannel, UserChannel),
	}
}

func (o *orchestrator) Start(context context.Context) {

	if _, err := o.redisPubSub.Receive(context); err != nil {
		log.Fatalf("error setting up redis %s \n", err)
	}

	ch := o.redisPubSub.Channel()
	defer func() { _ = o.redisPubSub.Close()} ()

	for {
		select {
		case msg := <- ch:
			m := Message{}
			if err := json.Unmarshal([]byte(msg.Payload), &m); err != nil {
				log.Println()
				continue
			}

			switch msg.Channel {
			case ReplyChannel:
				if m.Action != ActionDone {
					o.Rollback(context, m)
				}
			}
		}
	}

}

func (o *orchestrator) Rollback(context context.Context, m Message) {
	var channel string
	switch m.Service {
	case AuthService:
		channel = AuthChannel
	case UserService:
		channel = UserService
	}

	m.Action = ActionRollback
	if err := o.redisClient.Publish(context, channel, m).Err(); err != nil {
		log.Printf("error publishing rollback message to %s channel", channel)
	}

}

func (o *orchestrator) Next(context context.Context, channel string, service string, m Message) {
	m.Action = ActionStart
	m.Service = service
	if err := o.redisClient.Publish(context, channel, m).Err(); err != nil {
		log.Printf("error publishing start-message to %s channel", channel)
		log.Fatal(err)
	}
	log.Printf("start message published to channel :%s", channel)
}


