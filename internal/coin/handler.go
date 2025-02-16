package coin

import (
	"avito-intern/internal/auth"
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
)

type Service interface {
	GetUserByUsername(ctx context.Context, username string) (*auth.User, error)
	Transfer(ctx context.Context, from, to *auth.User, amount int) (*Transaction, error)
	Purchase(ctx context.Context, buyer *auth.User, amount int) (*Transaction, error)
	ListTransfers(ctx context.Context, user *auth.User) (incoming, outgoing []*Transaction, err error)
}

type Handler struct {
	svc          Service
	authHandlers AuthHandler
}

type AuthHandler interface {
	Verify(c *fiber.Ctx) error
}

func NewCoinHandler(svc Service, authHandler AuthHandler) *Handler {
	return &Handler{
		svc:          svc,
		authHandlers: authHandler,
	}
}

type ReceivedTx struct {
	FromUser string `json:"fromUser"`
	Amount   int    `json:"amount"`
}

type SentTx struct {
	FromUser string `json:"fromUser"`
	Amount   int    `json:"amount"`
}
type History struct {
	Received []ReceivedTx `json:"received"`
	Sent     []SentTx     `json:"sent"`
}

type InfoResponse struct {
	Coins       int     `json:"coins"`
	CoinHistory History `json:"coinHistory"`
}

type SendCoinRequest struct {
	ToUsername string `json:"toUser"`
	Amount     int    `json:"amount"`
}

func (h *Handler) Init(router fiber.Router) {
	router.Post("/sendCoin", h.authHandlers.Verify, h.sendCoin)
}

func (h *Handler) sendCoin(c *fiber.Ctx) error {
	ctx := c.UserContext()
	from, ok := auth.GetUser(ctx)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": "invalid user data",
		})
	}
	var coinReq SendCoinRequest
	if err := c.BodyParser(&coinReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": err.Error(),
		})
	}
	to, err := h.svc.GetUserByUsername(ctx, coinReq.ToUsername)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"errors": err.Error(),
		})
	}

	_, err = h.svc.Transfer(ctx, from, to, coinReq.Amount)
	if err != nil {
		if errors.Is(err, ErrInvalidRecipient) || errors.Is(err, ErrNotEnoughCoins) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"errors": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"errors": err.Error(),
		})
	}
	c.Status(fiber.StatusOK)
	return nil
}
