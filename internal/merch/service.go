package merch

import (
	"avito-intern/internal/auth"
	"avito-intern/internal/coin"
	"context"
	"errors"
	"time"
)

type service struct {
	authService auth.Service
	coinService coin.Service
	repo        Repository
}

func NewService(authService auth.Service, coinService coin.Service, repo Repository) Service {
	return &service{
		authService: authService,
		coinService: coinService,
		repo:        repo,
	}
}

func (s *service) Purchase(ctx context.Context, user *auth.User, merchName string) error {
	merch, err := s.repo.GetMerchByID(ctx, merchName)
	if err != nil {
		return err
	}

	totalCost := merch.Price * 1
	if user.CoinBalance < totalCost {
		return errors.New("not enough coins")
	}

	_, err = s.coinService.Purchase(ctx, user, totalCost)
	if err != nil {
		return err
	}

	err = s.repo.SavePurchase(ctx, &Purchase{
		UserID:      user.ID,
		MerchID:     merch.ID,
		Quantity:    1,
		PurchasedAt: time.Now(),
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *service) ListPurchases(ctx context.Context, user *auth.User) ([]*Purchase, error) {
	return s.repo.ListPurchasesByUserID(ctx, user.ID)
}

func (s *service) ListTransfers(ctx context.Context, user *auth.User) (incoming, outgoing []*coin.Transaction, err error) {
	return s.coinService.ListTransfers(ctx, user)
}
