package merch

import (
	"avito-intern/internal/auth"
	"avito-intern/internal/coin"
	"context"
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

type Service interface {
	Purchase(ctx context.Context, user *auth.User, merchName string) error
	ListPurchases(ctx context.Context, user *auth.User) ([]*Purchase, error)
	ListTransfers(ctx context.Context, user *auth.User) (incoming, outgoing []*coin.Transaction, err error)
}

type Handler struct {
	svc          Service
	authHandlers AuthHandler
}

type AuthHandler interface {
	Verify(c *fiber.Ctx) error
}

func NewMerchHandler(svc Service, authHandler AuthHandler) *Handler {
	return &Handler{
		svc:          svc,
		authHandlers: authHandler,
	}
}

func (h *Handler) Init(router fiber.Router) {
	router.Get("/info", h.authHandlers.Verify, h.info)
	router.Get("/buy/:item", h.authHandlers.Verify, h.buyItem)
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

type InventoryItem struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type InfoResponse struct {
	Coins       int             `json:"coins"`
	Inventory   []InventoryItem `json:"inventory"`
	CoinHistory History         `json:"coinHistory"`
}

func (h *Handler) info(c *fiber.Ctx) error {
	ctx := c.UserContext()
	user, ok := auth.GetUser(ctx)
	if !ok {
		return errors.New("failed to get user data")
	}
	in, out, err := h.svc.ListTransfers(ctx, user)
	if err != nil {
		slog.Error("failed to get transfers", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"errors": err.Error(),
		})
	}
	purchases, err := h.svc.ListPurchases(ctx, user)
	if err != nil {
		slog.Error("failed to get purchases", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"errors": err.Error(),
		})
	}

	var (
		received = make([]ReceivedTx, len(in))
		sent     = make([]SentTx, len(out))
	)

	for idx, row := range in {
		received[idx] = ReceivedTx{
			FromUser: row.FromUser.Username,
			Amount:   row.Amount,
		}
	}
	for idx, row := range out {
		sent[idx] = SentTx{
			FromUser: row.ToUser.Username,
			Amount:   row.Amount,
		}
	}

	inventoryMap := make(map[string]int)
	for _, purchase := range purchases {
		inventoryMap[purchase.MerchName] += purchase.Quantity
	}

	inventory := make([]InventoryItem, 0, len(inventoryMap))
	for merchType, quantity := range inventoryMap {
		inventory = append(inventory, InventoryItem{
			Type:     merchType,
			Quantity: quantity,
		})
	}

	return c.JSON(InfoResponse{
		Coins:     user.CoinBalance,
		Inventory: inventory,
		CoinHistory: History{
			Received: received,
			Sent:     sent,
		},
	})
}

func (h *Handler) buyItem(c *fiber.Ctx) error {
	ctx := c.UserContext()
	user, ok := auth.GetUser(ctx)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": "invalid user data",
		})
	}
	item := c.Params("item")
	if item == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": "item parameter is required",
		})
	}

	err := h.svc.Purchase(ctx, user, item)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"errors": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusOK)
}
