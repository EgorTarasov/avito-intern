package auth

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type service struct {
	users UserRepo
	cfg   *Config
}

func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func checkPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func NewService(cfg *Config, ur UserRepo) Service {
	return &service{
		cfg:   cfg,
		users: ur,
	}
}

func (s *service) AuthUser(ctx context.Context, username, password string) (Token, error) {
	user, err := s.users.GetUserByUsername(ctx, username)
	if err == nil && !checkPassword(user.Password, password) {
		return "", ErrUnauthorized
	}

	if errors.Is(err, ErrUserNotFound) {
		var errCreate error
		hashed, errCreate := hashPassword(password)
		if errCreate != nil {
			return "", NewErrInternal(errCreate)
		}

		user, errCreate = s.users.CreateUser(ctx, username, hashed)
		if errCreate != nil {
			return "", NewErrInternal(errCreate)
		}
	}
	token, err := NewToken(s.cfg.JWTSecret, s.cfg.TokenExpireDuration, user)
	if err != nil {
		return "", ErrUnauthorized
	}
	return token, nil
}

// GetUserFromToken интерфейс для получение данных пользователя из jwt токена.
func (s *service) GetUserFromToken(ctx context.Context, rawToken Token) (*User, error) {
	uid, err := rawToken.UserID(s.cfg.JWTSecret)
	if err != nil {
		return nil, ErrUnauthorized
	}
	u, err := s.users.GetUserByID(ctx, uid)
	if err != nil {
		return nil, ErrUnauthorized
	}
	return u, nil
}
