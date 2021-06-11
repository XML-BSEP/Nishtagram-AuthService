package usecase

import (
	"auth-service/domain"
	"auth-service/gateway"
	"auth-service/helper"
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
	UserGateway gateway.UserGateway
}


type RegistrationUsecase interface {
	Register(context context.Context, user domain.User) error
	ConfirmAccount(context context.Context, code string, email string) error
	IsAlreadyRegistered(context context.Context, username, email string) bool
	ResendCode(ctx context.Context, email string) (error, string, string)
}

func NewRegistrationUsecase(redisUsecase RedisUsecase, profileInfoUsecase ProfileInfoUsecase, gateway gateway.UserGateway) RegistrationUsecase{
	return &registrationUsecase{
		RedisUsecase: redisUsecase,
		ProfileInfoUsecase: profileInfoUsecase,
		UserGateway: gateway,
		}
}

func (s *registrationUsecase) Register(context context.Context, user domain.User) error{
	redisKey := redisKeyPattern + user.Email

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
	err = s.RedisUsecase.AddKeyValueSet(context, redisKey, serializedUser, time.Duration(expiration));
	if err != nil {
		return err
	}

	go SendMail(user.Email, user.Username, confirmationCode)
	return nil
}


func (s *registrationUsecase) ConfirmAccount(context context.Context, code string, email string) error {
	key := redisKeyPattern + email
	bytes, err := s.RedisUsecase.GetValueByKey(context, key)
	if err != nil {
		return err
	}


	user, err := deserialize(bytes)
	if err != nil {
		return err
	}
	fmt.Println("User id : " + user.ID)
	if err := helper.Verify(code, user.ConfirmationCode); err != nil {
		return err
	}
	if err := s.RedisUsecase.DeleteValueByKey(context, key); err != nil {
		return err
	}
	if err := s.ProfileInfoUsecase.Create(context, userToProfleInfo(user)); err != nil {
		return err
	}

	if err := s.UserGateway.SaveRegisteredUser(context, user); err != nil {
		return err
	}

	return nil
}

func (s *registrationUsecase) IsAlreadyRegistered(context context.Context, username, email string) bool {
	redisKey := redisKeyPattern + email
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
		ID: user.ID,
		Username: user.Username,
		Email: user.Email,
		Password: user.Password,
		RoleId : 3,
	}

}

func (s *registrationUsecase) ResendCode(ctx context.Context ,email string) (error, string, string) {
	rediskey := redisKeyPattern + email

	if !s.RedisUsecase.ExistsByKey(ctx,rediskey) {
		return fmt.Errorf("invalid email"), "", ""
	}
	bytes, err := s.RedisUsecase.GetValueByKey(ctx, rediskey)
	if err != nil {
		return err, "", ""
	}


	user, err := deserialize(bytes)
	if err != nil {
		return err, "", ""
	}

	code := helper.RandomStringGenerator(8)

	expiration  := 1000000000 * 3600 * 2 //2h
	hash, _ := helper.Hash(code)

	user.ConfirmationCode = string(hash)

	redisKey := redisKeyPattern + email

	serializedUser, err := serialize(*user)
	if err != nil {
		return err, "" ,""
	}
	err = s.RedisUsecase.AddKeyValueSet(ctx, redisKey, serializedUser, time.Duration(expiration))
	if err != nil {
		return err, "", ""
	}

	return nil, user.Email, code
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




