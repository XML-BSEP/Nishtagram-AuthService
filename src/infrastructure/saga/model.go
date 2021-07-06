package saga

import (
	"auth-service/domain"
	"encoding/json"
)

type Message struct {
	Service string `json:"service"`
	SenderService string `json:"sender_service"`
	Action string `json:"action"`
	Payload domain.User `json:"payload"`
	Confirm bool `json:"confirm"`
	Ok bool `json:"ok"`
}

func MarshalBinary(m *Message) ([]byte, error) {
	return json.Marshal(m)
}