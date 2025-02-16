package auth

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

// UserID — идентификатор пользователя в системе.
type UserID int64

func NewUserID(userID int64) (UserID, error) {
	if userID <= 0 {
		return 0, NewErrInvalidUserID(userID)
	}
	return UserID(userID), nil
}

// User — представление пользователя в системе.
type User struct {
	ID          UserID
	Username    string
	Password    string
	CoinBalance int
	CreatedAt   time.Time
}

// Claims определяет наши собственные JWT claims, включая идентификатор пользователя.
type Claims struct {
	UserID int64
	jwt.RegisteredClaims
}

// Token представляет JWT токен для аутентификации пользователя.
type Token string

// NewToken создает подписанный JWT токен с идентификатором пользователя.
// Токен будет действителен в течение 24 часов.
func NewToken(secretKey string, expireTime time.Duration, user *User) (Token, error) {
	claims := &Claims{
		UserID: int64(user.ID),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", NewErrUnableToSignToken(err)
	}
	return Token(tokenString), nil
}

// UserID парсит JWT токен и возвращает содержащийся в нем идентификатор пользователя.
func (t Token) UserID(secretKey string) (UserID, error) {
	token, err := jwt.ParseWithClaims(string(t), &Claims{}, func(_ *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return 0, NewErrUnableToParseToken(err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return 0, ErrInvalidToken
	}
	userID, err := NewUserID(claims.UserID)
	if err != nil {
		return 0, NewErrInvalidUserID(claims.UserID)
	}

	return userID, nil
}
