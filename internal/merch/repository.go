package merch

import (
	"avito-intern/internal/auth"
	"context"
)

type Repository interface {
	GetMerchByID(ctx context.Context, merchName string) (*Merch, error)
	SavePurchase(ctx context.Context, purchase *Purchase) error
	ListPurchasesByUserID(ctx context.Context, userID auth.UserID) ([]*Purchase, error)
}
