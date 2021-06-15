package helper

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"os"
)

func ExtractUserIdFromToken(tokenString string) (*string, error) {

	if tokenString == "" {
		return nil, fmt.Errorf("", "message= %s", "Authorization header does not exist")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})

	claims, ok := token.Claims.(jwt.MapClaims)

	if ok  {
		userId, ok := claims["user_id"].(string)
		if !ok {
			return nil, err
		}

		return &userId, nil
	}
	return nil, err
}

func ExtractTokenUuid (tokenString string) (*string, error) {

	if tokenString == "" {
		return nil, fmt.Errorf("", "message= %s", "Authorization header does noe exist")

	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})

	claims, ok := token.Claims.(jwt.MapClaims)

	if ok  {
		userId, ok := claims["token_uuid"].(string)
		if !ok {
			return nil, err
		}

		return &userId, nil
	}
	return nil, err
}
