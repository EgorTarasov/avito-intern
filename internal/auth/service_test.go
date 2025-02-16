package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Dummy Types and Fake Repo ---

// fakeUserRepo implements the UserRepo interface.
type fakeUserRepo struct {
	getUserByUsernameFunc func(ctx context.Context, username string) (*User, error)
	createUserFunc        func(ctx context.Context, username, password string) (*User, error)
	getUserByIDFunc       func(ctx context.Context, userID UserID) (*User, error)
}

func (f *fakeUserRepo) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	if f.getUserByUsernameFunc != nil {
		return f.getUserByUsernameFunc(ctx, username)
	}
	return nil, ErrUserNotFound
}

func (f *fakeUserRepo) CreateUser(ctx context.Context, username, password string) (*User, error) {
	if f.createUserFunc != nil {
		return f.createUserFunc(ctx, username, password)
	}
	return nil, errors.New("not implemented")
}

func (f *fakeUserRepo) GetUserByID(ctx context.Context, userID UserID) (*User, error) {
	if f.getUserByIDFunc != nil {
		return f.getUserByIDFunc(ctx, userID)
	}
	return nil, errors.New("user not found")
}

// createValidToken is a helper to create a token for testing using NewToken.
func createValidToken(t *testing.T, cfg *Config, user *User) Token {
	token, err := NewToken(cfg.JWTSecret, cfg.TokenExpireDuration, user)
	require.NoError(t, err)
	return token
}

// --- Tests ---

func TestAuthUser_CorrectPassword(t *testing.T) {
	cfg := &Config{
		JWTSecret:           testSecretKey,
		TokenExpireDuration: time.Minute,
	}

	plainPassword := "secret123"
	hashed, err := hashPassword(plainPassword)
	require.NoError(t, err)

	// Existing user with correct hashed password.
	user := &User{
		ID:       1,
		Username: "user1",
		Password: hashed,
	}

	repo := &fakeUserRepo{
		getUserByUsernameFunc: func(_ context.Context, username string) (*User, error) {
			if username == user.Username {
				return user, nil
			}
			return nil, ErrUserNotFound
		},
	}

	svc := NewService(cfg, repo)
	token, err := svc.AuthUser(context.Background(), user.Username, plainPassword)
	require.NoError(t, err, "authUser should succeed")
	assert.NotEmpty(t, token, "expected token not to be empty")
}

func TestAuthUser_WrongPassword(t *testing.T) {
	cfg := &Config{
		JWTSecret:           testSecretKey,
		TokenExpireDuration: time.Minute,
	}

	plainPassword := "correctPass"
	hashed, err := hashPassword(plainPassword)
	require.NoError(t, err)

	// Existing user but provided wrong password.
	user := &User{
		ID:       2,
		Username: "user2",
		Password: hashed,
	}

	repo := &fakeUserRepo{
		getUserByUsernameFunc: func(_ context.Context, _ string) (*User, error) {
			return user, nil
		},
	}

	svc := NewService(cfg, repo)
	token, err := svc.AuthUser(context.Background(), user.Username, "wrongPass")
	assert.Error(t, err, "expected error for wrong password")
	assert.Equal(t, ErrUnauthorized, err, "error should be ErrUnauthorized")
	assert.Empty(t, token, "expected token to be empty")
}

func TestAuthUser_UserNotFound_CreateUserSuccess(t *testing.T) {
	cfg := &Config{
		JWTSecret:           testSecretKey,
		TokenExpireDuration: time.Minute,
	}

	username := "newuser"
	plainPassword := "newpassword"

	repo := &fakeUserRepo{
		getUserByUsernameFunc: func(_ context.Context, _ string) (*User, error) {
			// Simulate user not found.
			return nil, ErrUserNotFound
		},
		createUserFunc: func(_ context.Context, username, password string) (*User, error) {
			hashed, err := hashPassword(password)
			require.NoError(t, err)
			// Simulate successful user creation.
			return &User{ID: 3, Username: username, Password: hashed}, nil
		},
	}

	svc := NewService(cfg, repo)
	token, err := svc.AuthUser(context.Background(), username, plainPassword)
	require.NoError(t, err, "expected authUser to create user successfully")
	assert.NotEmpty(t, token, "expected token not to be empty")
}

func TestAuthUser_CreateUserFailure(t *testing.T) {
	cfg := &Config{
		JWTSecret:           testSecretKey,
		TokenExpireDuration: time.Minute,
	}

	username := "failuser"
	plainPassword := "failpassword"

	repo := &fakeUserRepo{
		getUserByUsernameFunc: func(_ context.Context, _ string) (*User, error) {
			return nil, ErrUserNotFound
		},
		createUserFunc: func(_ context.Context, _, _ string) (*User, error) {
			return nil, errors.New("failed to create user")
		},
	}

	svc := NewService(cfg, repo)
	token, err := svc.AuthUser(context.Background(), username, plainPassword)
	assert.Error(t, err, "expected error creating user")
	assert.Empty(t, token, "token should be empty on error")
}

func TestValidateToken_Success(t *testing.T) {
	cfg := &Config{
		JWTSecret:           testSecretKey,
		TokenExpireDuration: time.Minute,
	}

	plainPassword := "validpass"
	hashed, err := hashPassword(plainPassword)
	require.NoError(t, err)

	// Create a dummy user.
	user := &User{
		ID:       4,
		Username: "validuser",
		Password: hashed,
	}

	// Create a valid token for the user.
	token := createValidToken(t, cfg, user)

	repo := &fakeUserRepo{
		getUserByIDFunc: func(_ context.Context, userID UserID) (*User, error) {
			if userID == user.ID {
				return user, nil
			}
			return nil, ErrUserNotFound
		},
	}

	svc := NewService(cfg, repo)
	retUser, err := svc.GetUserFromToken(context.Background(), token)
	require.NoError(t, err, "ValidateToken should succeed with a valid token")
	assert.Equal(t, user, retUser, "the returned user should match")
}

func TestValidateToken_InvalidToken(t *testing.T) {
	cfg := &Config{
		JWTSecret:           testSecretKey,
		TokenExpireDuration: time.Minute,
	}

	repo := &fakeUserRepo{}
	svc := NewService(cfg, repo)

	// Use an obviously invalid token.
	invalidToken := Token("this.is.invalid")
	u, err := svc.GetUserFromToken(context.Background(), invalidToken)
	assert.Error(t, err, "ValidateToken should return an error for an invalid token")
	assert.Nil(t, u, "expected returned user to be nil")
}

func TestValidateToken_UserNotFound(t *testing.T) {
	cfg := &Config{
		JWTSecret:           testSecretKey,
		TokenExpireDuration: time.Minute,
	}

	plainPassword := "testpass"
	hashed, err := hashPassword(plainPassword)
	require.NoError(t, err)

	// Create a dummy user.
	user := &User{
		ID:       5,
		Username: "missinguser",
		Password: hashed,
	}

	token := createValidToken(t, cfg, user)

	repo := &fakeUserRepo{
		getUserByIDFunc: func(_ context.Context, _ UserID) (*User, error) {
			// Simulate user not found.
			return nil, ErrUserNotFound
		},
	}

	svc := NewService(cfg, repo)
	u, err := svc.GetUserFromToken(context.Background(), token)
	assert.Error(t, err, "expected error when user not found")
	assert.Nil(t, u, "expected nil user when lookup fails")
}
