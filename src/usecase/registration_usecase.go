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
	logger "github.com/jelena-vlajkov/logger/logger"
	"time"
)

const (
	redisKeyPattern = "registrationRequest/"
	agentRegistrationRequest = "agentRegistrationRequest/"
	agent = "agent/"
)
type registrationUsecase struct {
	RedisUsecase RedisUsecase
	ProfileInfoUsecase ProfileInfoUsecase
	UserGateway gateway.UserGateway
	logger *logger.Logger
}


type RegistrationUsecase interface {
	Register(context context.Context, user domain.User) error
	ConfirmAccount(context context.Context, code string, email string) error
	IsAlreadyRegistered(context context.Context, username, email string) bool
	ResendCode(ctx context.Context ,email string) (error, string, string)
	ValidateAgentAccount(context context.Context, code string, email string) error
	RegisterAgent(context context.Context, user domain.User) error
	ConfirmAgentAccount(context context.Context, email string, confirm bool) error
	GetAgentRequests(context context.Context) ([]domain.User ,error)
}

func NewRegistrationUsecase(redisUsecase RedisUsecase, profileInfoUsecase ProfileInfoUsecase, gateway gateway.UserGateway, logger *logger.Logger) RegistrationUsecase{
	return &registrationUsecase{
		logger: logger,
		RedisUsecase: redisUsecase,
		ProfileInfoUsecase: profileInfoUsecase,
		UserGateway: gateway,
		}
}

func (s *registrationUsecase) Register(context context.Context, user domain.User) error{
	s.logger.Logger.Infof("registering user with email %v\n", user.Email)
	redisKey := redisKeyPattern + user.Email

	confirmationCode := helper.RandomStringGenerator(8)
	hashedConfirmationCode, err := Hash(confirmationCode)
	if err != nil {
		s.logger.Logger.Errorf("error while registering user, error %v\n", err)
		return err
	}

	hashedPassword, err := helper.Hash(user.Password)
	if err != nil {
		s.logger.Logger.Errorf("error while registering user, error %v\n", err)
		return err
	}

	user.ID = uuid.NewString()
	user.ConfirmationCode = string(hashedConfirmationCode)
	user.Password = string(hashedPassword)


	expiration  := 1000000000 * 3600 * 2 //2h
	serializedUser, err := serialize(user)
	if err != nil {
		s.logger.Logger.Errorf("error while registering user, error %v\n", err)
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
	s.logger.Logger.Infof("confirming account for email %v\n", email)
	key := redisKeyPattern + email
	bytes, err := s.RedisUsecase.GetValueByKey(context, key)
	if err != nil {
		return err
	}


	user, err := deserialize(bytes)
	if err != nil {
		s.logger.Logger.Errorf("error while confirming account, error %v\n", err)
		return err
	}
	fmt.Println("User id : " + user.ID)
	if err := helper.Verify(code, user.ConfirmationCode); err != nil {
		s.logger.Logger.Errorf("error while confirming account, error %v\n", err)
		return err
	}
	if err := s.RedisUsecase.DeleteValueByKey(context, key); err != nil {
		s.logger.Logger.Errorf("error while confirming account, error %v\n", err)
		return err
	}
	if err := s.ProfileInfoUsecase.Create(context, userToProfleInfo(user)); err != nil {
		s.logger.Logger.Errorf("error while confirming account, error %v\n", err)
		return err
	}

	if err := s.UserGateway.SaveRegisteredUser(context, user); err != nil {
		return err
	}

	return nil
}

func (s *registrationUsecase) IsAlreadyRegistered(context context.Context, username, email string) bool {
	s.logger.Logger.Infof("checking if user already exists")
	redisKey := redisKeyPattern + email
	if s.RedisUsecase.ExistsByKey(context, redisKey) {
		return true
	}
	//TODO:Prefix scan with pattern and check if username already exists
	agentRegistrationRequestKey := agentRegistrationRequest + email
	if s.RedisUsecase.ExistsByKey(context, agentRegistrationRequestKey) {
		return true
	}

	redisAgentKey := agent + email
	if s.RedisUsecase.ExistsByKey(context, redisAgentKey) {
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
func agentToProfleInfo(user *domain.User) *domain.ProfileInfo{
	return &domain.ProfileInfo{
		ID: user.ID,
		Username: user.Username,
		Email: user.Email,
		Password: user.Password,
		RoleId : 2,
	}

}

func (s *registrationUsecase) ResendCode(ctx context.Context ,email string) (error, string, string) {
	s.logger.Logger.Infof("resending code for email %v\n", email)
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
		s.logger.Logger.Errorf("error while confirming account, error %v\n", err)
		return err, "", ""
	}

	code := helper.RandomStringGenerator(8)

	expiration  := 1000000000 * 3600 * 2 //2h
	hash, _ := helper.Hash(code)

	user.ConfirmationCode = string(hash)

	redisKey := redisKeyPattern + user.Email

	serializedUser, err := serialize(*user)
	if err != nil {
		s.logger.Logger.Errorf("error while confirming account, error %v\n", err)
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

func (s *registrationUsecase) ValidateAgentAccount(context context.Context, code string, email string) error {
	s.logger.Logger.Infof("confirming account for email %v\n", email)
	key := agent + email
	bytes, err := s.RedisUsecase.GetValueByKey(context, key)
	if err != nil {
		return err
	}


	user, err := deserialize(bytes)
	if err != nil {
		s.logger.Logger.Errorf("error while confirming account, error %v\n", err)
		return err
	}
	fmt.Println("User id : " + user.ID)
	if err := helper.Verify(code, user.ConfirmationCode); err != nil {
		s.logger.Logger.Errorf("error while confirming account, error %v\n", err)
		return err
	}
	if err := s.RedisUsecase.DeleteValueByKey(context, key); err != nil {
		s.logger.Logger.Errorf("error while confirming account, error %v\n", err)
		return err
	}

	regRequestKey := agentRegistrationRequest + email

	if err := s.RedisUsecase.AddKeyValueSet(context, regRequestKey, bytes, time.Duration(0)); err != nil {
		return err
	}



	return nil

}

func (s *registrationUsecase) RegisterAgent(context context.Context, user domain.User) error {
	s.logger.Logger.Infof("registering user with email %v\n", user.Email)
	redisKey := agent + user.Email

	confirmationCode := helper.RandomStringGenerator(8)
	hashedConfirmationCode, err := Hash(confirmationCode)
	if err != nil {
		s.logger.Logger.Errorf("error while registering user, error %v\n", err)
		return err
	}

	hashedPassword, err := helper.Hash(user.Password)
	if err != nil {
		s.logger.Logger.Errorf("error while registering user, error %v\n", err)
		return err
	}

	user.ID = uuid.NewString()
	user.ConfirmationCode = string(hashedConfirmationCode)
	user.Password = string(hashedPassword)


	expiration  := 1000000000 * 3600 * 2 //2h
	serializedUser, err := serialize(user)
	if err != nil {
		s.logger.Logger.Errorf("error while registering user, error %v\n", err)
		return err
	}
	err = s.RedisUsecase.AddKeyValueSet(context, redisKey, serializedUser, time.Duration(expiration));
	if err != nil {
		return err
	}

	go SendMail(user.Email, user.Username, confirmationCode)
	return nil
}

func (s *registrationUsecase) ConfirmAgentAccount(context context.Context, email string, confirm bool) error {
	key := agentRegistrationRequest + email
	bytes, err := s.RedisUsecase.GetValueByKey(context, key)
	if err != nil {
		return err
	}


	user, err := deserialize(bytes)
	if err != nil {
		s.logger.Logger.Errorf("error while confirming account, error %v\n", err)
		return err
	}
	if err := s.RedisUsecase.DeleteValueByKey(context, key); err != nil {
		s.logger.Logger.Errorf("error while confirming account, error %v\n", err)
		return err
	}
	if confirm {
		if err := s.ProfileInfoUsecase.Create(context, agentToProfleInfo(user)); err != nil {
			s.logger.Logger.Errorf("error while confirming account, error %v\n", err)
			return err
		}

		if err := s.UserGateway.SaveRegisteredUser(context, user); err != nil {
			return err
		}
	}

	return nil
}

func (s *registrationUsecase) GetAgentRequests(context context.Context) ([]domain.User, error) {


	keys, err := s.RedisUsecase.ScanKeyByPattern(context, agentRegistrationRequest + "*")
	if err != nil {
		return nil, err
	}

	values, err := s.getValuesByKeys(context, keys)
	if err != nil {
		return nil, err
	}

	return values, nil
}

func (s *registrationUsecase) getValuesByKeys(context context.Context, keys []string) ([]domain.User, error) {


	var values []domain.User

	for _, key := range keys {
		val, err := s.RedisUsecase.GetValueByKey(context, key)
		if err != nil {
			continue
		}
		user, err  := deserialize(val)
		if err != nil {
			continue
		}
		values = append(values, *user)
	}

	return values, nil
}


