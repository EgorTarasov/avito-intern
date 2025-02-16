package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// fakeService mocks the methods used by the handler.
type fakeService struct {
	// authUserFunc simulates behavior of authUser.
	authUserFunc         func(ctx context.Context, username, password string) (Token, error)
	getUserFromTokenFunc func(ctx context.Context, token Token) (*User, error)
}

func (f *fakeService) AuthUser(ctx context.Context, username, password string) (Token, error) {
	return f.authUserFunc(ctx, username, password)
}

func (f *fakeService) GetUserFromToken(ctx context.Context, token Token) (*User, error) {
	if f.getUserFromTokenFunc != nil {
		return f.getUserFromTokenFunc(ctx, token)
	}
	return nil, nil
}

func (f *fakeService) GetUserByUsername(_ context.Context, _ string) (*User, error) {
	// Return a dummy user or nil as needed by tests.
	return nil, nil
}

// setupTestHandler initializes a Fiber app with our auth handler.
func setupTestHandler(svc *fakeService) *fiber.App {
	app := fiber.New()
	handlers := NewAuthHandlers(svc)
	// register the auth endpoint
	app.Post("/auth", handlers.auth)
	return app
}

func setupVerifyTestHandler(svc Service) *fiber.App {
	app := fiber.New()
	handlers := NewAuthHandlers(svc)
	// Setup a test route using the Verify middleware.
	app.Get("/verify", handlers.Verify, func(c *fiber.Ctx) error {
		return c.SendString("next called")
	})
	return app
}

func TestAuth_Success(t *testing.T) {
	fakeSvc := &fakeService{
		authUserFunc: func(_ context.Context, _, _ string) (Token, error) {
			return "valid-token", nil
		},
	}
	app := setupTestHandler(fakeSvc)

	payload := TokenRequest{
		Username: "user",
		Password: "pass",
	}
	bodyBytes, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/auth", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	// Close body after reading.
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var res TokenResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	assert.NoError(t, err)
	assert.Equal(t, "valid-token", res.Token)
}

func TestAuth_Unauthorized(t *testing.T) {
	fakeSvc := &fakeService{
		authUserFunc: func(_ context.Context, _, _ string) (Token, error) {
			return "", ErrUnauthorized
		},
	}
	app := setupTestHandler(fakeSvc)

	payload := TokenRequest{
		Username: "user",
		Password: "wrongpass",
	}
	bodyBytes, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/auth", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuth_UserNotFound(t *testing.T) {
	fakeSvc := &fakeService{
		authUserFunc: func(_ context.Context, _, _ string) (Token, error) {
			return "", ErrUserNotFound
		},
	}
	app := setupTestHandler(fakeSvc)

	payload := TokenRequest{
		Username: "nonexistent",
		Password: "pass",
	}
	bodyBytes, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/auth", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestAuth_InternalError(t *testing.T) {
	fakeSvc := &fakeService{
		authUserFunc: func(_ context.Context, _, _ string) (Token, error) {
			return "", NewErrInternal(errors.New("database error"))
		},
	}
	app := setupTestHandler(fakeSvc)

	payload := TokenRequest{
		Username: "user",
		Password: "pass",
	}
	bodyBytes, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/auth", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestAuth_BadRequest(t *testing.T) {
	fakeSvc := &fakeService{
		authUserFunc: func(_ context.Context, _, _ string) (Token, error) {
			return "", nil
		},
	}
	app := setupTestHandler(fakeSvc)

	// Invalid JSON body.
	req := httptest.NewRequest("POST", "/auth", bytes.NewReader([]byte("{invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Missing required fields.
	payload := TokenRequest{
		Username: "",
		Password: "",
	}
	bodyBytes, _ := json.Marshal(payload)
	req = httptest.NewRequest("POST", "/auth", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestVerify_MissingHeader(t *testing.T) {
	svc := &fakeService{}
	app := setupVerifyTestHandler(svc)

	req := httptest.NewRequest("GET", "/verify", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var body map[string]string
	err = json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(t, err)
	assert.Equal(t, "Отсутствует заголовок Authorization", body["errors"])
}

func TestVerify_BadFormat(t *testing.T) {
	svc := &fakeService{}
	app := setupVerifyTestHandler(svc)

	req := httptest.NewRequest("GET", "/verify", nil)
	req.Header.Set("Authorization", "Token invalid-format")
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var body map[string]string
	err = json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(t, err)
	assert.Equal(t, "Неверный формат токена", body["errors"])
}

func TestVerify_FailedVerification(t *testing.T) {
	svc := &fakeService{
		getUserFromTokenFunc: func(_ context.Context, _ Token) (*User, error) {
			return nil, errors.New("token invalid")
		},
	}
	app := setupVerifyTestHandler(svc)

	req := httptest.NewRequest("GET", "/verify", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var body map[string]string
	err = json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(t, err)
	assert.Equal(t, "Пользователь не прошел проверку", body["errors"])
}

func TestVerify_Success(t *testing.T) {
	svc := &fakeService{
		getUserFromTokenFunc: func(_ context.Context, _ Token) (*User, error) {
			return &User{
				ID:       1,
				Username: "validUser",
			}, nil
		},
	}
	app := setupVerifyTestHandler(svc)

	req := httptest.NewRequest("GET", "/verify", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	bodyBytes := new(bytes.Buffer)
	_, err = io.Copy(bodyBytes, resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "next called", bodyBytes.String())
}
