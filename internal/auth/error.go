package auth

import (
	"errors"
	"fmt"
)

var (
	Err             = errors.New("auth")
	ErrInvalidToken = fmt.Errorf("%v: недопустимый токен", Err)
	ErrUserNotFound = fmt.Errorf("%v: пользователь не найден", Err)
	ErrUnauthorized = fmt.Errorf("%v: пользователь не авторизован", Err)
)

func NewErrInvalidUserID(userID int64) error {
	return ErrInvalidUserID{userID: userID}
}

type ErrInvalidUserID struct {
	userID int64
}

func (e ErrInvalidUserID) Error() string {
	return fmt.Sprintf("%v: недопустимый userID %d: должен быть > 0", Err, e.userID)
}

func NewErrUnableToSignToken(err error) error {
	return ErrUnableToSignToken{err: err}
}

type ErrUnableToSignToken struct {
	err error
}

func (e ErrUnableToSignToken) Error() string {
	return fmt.Sprintf("%v: не удалось подписать токен: %v", Err, e.err)
}

// NewErrUnableToParseToken создает новый error для ошибок парсинга токена.
func NewErrUnableToParseToken(err error) error {
	return ErrUnableToParseToken{err: err}
}

type ErrUnableToParseToken struct {
	err error
}

func (e ErrUnableToParseToken) Error() string {
	return fmt.Sprintf("%v: не удалось распарсить токен: %v", Err, e.err)
}

func NewErrInternal(err error) error {
	return ErrInternal{err: err}
}

type ErrInternal struct {
	err error
}

func (e ErrInternal) Error() string {
	return fmt.Sprintf("%v: internal err: %v", Err, e.err)
}
