package usecase

import (
	"auth-service/src/domain"
	"auth-service/src/helper"
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/google/uuid"
	"time"
)

const (
	redisKeyPattern = "registrationRequest"
)
type registrationUsecase struct {
	RedisUsecase RedisUsecase
	ProfileInfoUsecase ProfileInfoUsecase
}


type RegistrationUsecase interface {
	Register(context context.Context, user domain.User) error
	ConfirmAccount(context context.Context, code string, username string) error
	IsAlreadyRegistered(context context.Context, username, email string) bool
}

func NewRegistrationUsecase(redisUsecase RedisUsecase, profileInfoUsecase ProfileInfoUsecase) RegistrationUsecase{
	return &registrationUsecase{RedisUsecase: redisUsecase, ProfileInfoUsecase: profileInfoUsecase}
}

func (s *registrationUsecase) Register(context context.Context, user domain.User) error{
	redisKey := redisKeyPattern + user.Username

	confirmationCode := helper.RandomStringGenerator(8)
	hashedConfirmationCode, err := Hash(confirmationCode)
	fmt.Print("Generisan kod: " + confirmationCode)
	fmt.Print("Hashovan kod: " + string(hashedConfirmationCode))
	if err != nil {
		return err
	}

	hashedPassword, err := helper.Hash(user.Password)
	if err != nil {
		return err
	}

	user.ID = uuid.NewString()
	user.ConfirmationCode = string(hashedConfirmationCode)
	user.Password = string(hashedPassword)


	expiration  := 1000000000 * 3600 * 2 //2h
	serializedUser, err := serialize(user)
	if err != nil {
		return err
	}
	if err := s.RedisUsecase.AddKeyValueSet(context, redisKey, serializedUser, time.Duration(expiration)); err != nil {
		return err
	}

	go SendMail(user.Email, user.Username, confirmationCode)
	return nil
}


func (s *registrationUsecase) ConfirmAccount(context context.Context, code string, username string) error {
	key := redisKeyPattern + username
	bytes, err := s.RedisUsecase.GetValueByKey(context, key)
	if err != nil {
		return err
	}

	fmt.Println("Poslat kod: " + code)
	user, err := deserialize(bytes)
	if err != nil {
		return err
	}

	if err := helper.Verify(code, user.ConfirmationCode); err != nil {
		return err
	}
	if err := s.RedisUsecase.DeleteValueByKey(context, key); err != nil {
		return err
	}
	if err := s.ProfileInfoUsecase.Create(context, userToProfleInfo(user)); err != nil {
		return err
	}
	return nil
}

func (s *registrationUsecase) IsAlreadyRegistered(context context.Context, username, email string) bool {
	redisKey := redisKeyPattern + username
	if s.RedisUsecase.ExistsByKey(context, redisKey) {
		return true
	}

	if s.ProfileInfoUsecase.ExistsByUsernameOrEmail(context, username, email) {
		return true
	}

	return false
}
func userToProfleInfo(user *domain.User) *domain.ProfileInfo{
	return &domain.ProfileInfo{
		Username: user.Username,
		Email: user.Email,
		Password: user.Password,
		RoleId : 3,
	}

}


func serialize(value domain.User) ([]byte, error){
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	if err := e.Encode(value); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func deserialize(bytesArray []byte) (*domain.User, error){
	b := bytes.Buffer{}
	b.Write(bytesArray)
	d := gob.NewDecoder(&b)
	var decoded *domain.User
	if err := d.Decode(&decoded); err != nil {
		return nil, err
	}

	return decoded, nil
}




