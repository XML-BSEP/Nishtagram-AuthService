package usecase

import (
	"auth-service/infrastructure/tracer"
	"context"
	"golang.org/x/crypto/bcrypt"
)

func Hash(password string) ([]byte, error) {

	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func VerifyPassword(context context.Context, password, hashedPassword string) error {
	span := tracer.StartSpanFromContext(context, "VerifyPassword")
	defer span.Finish()

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		tracer.LogError(span, err)
		return err
	}
	return nil
}
