package auth

import (
	"context"
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
)

type userKey struct{}

func SetUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userKey{}, user)
}

func GetUser(ctx context.Context) (*User, bool) {
	if u, ok := ctx.Value(userKey{}).(*User); ok {
		return u, true
	}
	return nil, false
}

type Service interface {
	AuthUser(ctx context.Context, username, password string) (Token, error)
	GetUserFromToken(ctx context.Context, rawToken Token) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
}

type Handlers struct {
	svc Service
}

func NewAuthHandlers(svc Service) *Handlers {
	return &Handlers{
		svc: svc,
	}
}

func (h *Handlers) Init(router fiber.Router) {
	router.Post("/auth", h.auth)
}

// TokenRequest represents the expected request body.
type TokenRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// TokenResponse is the response containing the JWT token.
type TokenResponse struct {
	Token string `json:"token"`
}

// auth Аутентификация и получение JWT-токена.
// При первой аутентификации пользователь создается автоматически.
func (h *Handlers) auth(c *fiber.Ctx) error {
	var req TokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": "Неверный запрос: " + err.Error(),
		})
	}

	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": "Поля username и password обязательны",
		})
	}

	token, err := h.svc.AuthUser(c.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, ErrUnauthorized) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"errors": "Не авторизован",
			})
		}
		if errors.Is(err, ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"errors": "Пользователь не найден",
			})
		}
		var internalErr ErrInternal
		if errors.As(err, &internalErr) {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"errors": "Внутренняя ошибка сервера",
			})
		}
		log.Printf("auth error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"errors": "Неизвестная ошибка",
		})
	}

	return c.Status(fiber.StatusOK).JSON(TokenResponse{Token: string(token)})
}

// verify middleware.
func (h *Handlers) Verify(c *fiber.Ctx) error {
	ctx := context.Background()
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"errors": "Отсутствует заголовок Authorization",
		})
	}

	const prefix = "Bearer "
	if len(authHeader) < len(prefix) || authHeader[:len(prefix)] != prefix {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"errors": "Неверный формат токена",
		})
	}

	token := authHeader[len(prefix):]

	// Проверка токена через метод сервиса.
	user, err := h.svc.GetUserFromToken(c.Context(), Token(token))
	if err != nil || user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"errors": "Пользователь не прошел проверку",
		})
	}
	ctx = SetUser(ctx, user)
	c.SetUserContext(ctx)
	return c.Next()
}
