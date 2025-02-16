package auth

import (
	"errors"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testExpireDuration = time.Second * 1
	testSecretKey      = "testsecret"
)

func TestNewUserIDValid(t *testing.T) {
	validID := int64(10)
	uid, err := NewUserID(validID)
	require.NoError(t, err, "неожиданная ошибка")
	assert.Equal(t, UserID(validID), uid, "ожидался UserID %d, получено %d", validID, uid)
}

func TestNewUserIDInvalid(t *testing.T) {
	tests := []struct {
		name  string
		input int64
	}{
		{"zero value", 0},
		{"negative value", -5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewUserID(tc.input)
			assert.Error(t, err, "ожидалась ошибка для userID %d", tc.input)
		})
	}
}

func TestValidToken(t *testing.T) {
	user := &User{
		ID:       1,
		Username: "testuser",
	}

	token, err := NewToken(testSecretKey, testExpireDuration, user)
	require.NoError(t, err, "ошибка при создании токена")

	uid, err := token.UserID(testSecretKey)
	require.NoError(t, err, "ошибка при разборе токена")
	assert.Equal(t, user.ID, uid, "ожидался userID %d, получен %d", user.ID, uid)
}

func TestInvalidSecret(t *testing.T) {
	wrongSecret := "wrongsecret"
	user := &User{
		ID:       2,
		Username: "anotheruser",
	}

	token, err := NewToken(testSecretKey, testExpireDuration, user)
	require.NoError(t, err, "ошибка при создании токена")

	_, err = token.UserID(wrongSecret)
	assert.Error(t, err, "ожидалась ошибка при разборе токена с некорректным секретом")

	var parseErr ErrUnableToParseToken
	assert.True(t, errors.As(err, &parseErr), "ошибка не того типа, ожидалась ErrUnableToParseToken")
}

func TestExpiredToken(t *testing.T) {
	secretKey := "testsecret"

	user := &User{
		ID:       3,
		Username: "expireduser",
	}

	token, err := NewToken(secretKey, testExpireDuration, user)
	require.NoError(t, err, "ошибка при подписании токена")

	// Ждем, пока токен истечет.
	time.Sleep(2 * testExpireDuration)

	_, err = token.UserID(secretKey)
	assert.Error(t, err, "ожидалась ошибка при разборе истекшего токена")
	assert.ErrorAs(t, err, &ErrInvalidToken)
}

func TestInvalidToken(t *testing.T) {
	// Используем строку, не являющуюся корректным JWT.
	token := Token("invalid.token")
	_, err := token.UserID(testSecretKey)
	assert.Error(t, err, "ожидалась ошибка при разборе поврежденного токена")

	var parseErr ErrUnableToParseToken
	assert.True(t, errors.As(err, &parseErr), "ошибка не того типа, ожидалась ErrUnableToParseToken")
}

func TestInvalidClaimsToken(t *testing.T) {
	// Создаем токен с jwt.MapClaims вместо *AuthClaim,
	// чтобы тип claims оказался некорректным.
	mapClaims := jwt.MapClaims{
		"fooBarUserID": 120,
		"exp":          time.Now().Add(24 * time.Hour).Unix(),
		"iat":          time.Now().Unix(),
	}
	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, mapClaims)
	tokenString, err := tokenObj.SignedString([]byte(testSecretKey))
	require.NoError(t, err, "ошибка при подписании токена")
	token := Token(tokenString)

	uid, err := token.UserID(testSecretKey)
	assert.Error(t, err)
	assert.Equal(t, UserID(0), uid, "ожидался uid == 0, получено %d", uid)

	// assert.True(t, errors.Is(err, ErrInvalidToken), "ожидалась ошибка '%v', получена %v", ErrInvalidToken, err)
}

func TestEmptyToken(t *testing.T) {
	token := Token("")
	_, err := token.UserID(testSecretKey)
	assert.Error(t, err, "ожидалась ошибка при разборе пустого токена")

	var parseErr ErrUnableToParseToken
	assert.True(t, errors.As(err, &parseErr), "ошибка не того типа, ожидалась ErrUnableToParseToken")
}
