package coin

import (
	"avito-intern/internal/auth"
	"avito-intern/internal/common"
	"context"
	"errors"
	"time"

	"golang.org/x/sync/errgroup"
)

type service struct {
	authService  auth.Service
	transactions Repository
}

func NewService(authService auth.Service, transactions Repository) Service {
	return &service{
		authService:  authService,
		transactions: transactions,
	}
}

func (s *service) Transfer(ctx context.Context, from, to *auth.User, amount int) (*Transaction, error) {
	if from == nil || to == nil {
		return nil, errors.New("missing required data")
	}

	if from.ID == to.ID {
		return nil, ErrInvalidRecipient
	}
	if from.CoinBalance < amount {
		return nil, ErrNotEnoughCoins
	}
	return s.transactions.SaveTransaction(ctx, &Transaction{
		ID:        0,
		FromUser:  from,
		ToUser:    to,
		Amount:    amount,
		Type:      Transfer,
		CreatedAt: time.Now(),
	})
}

func (s *service) Purchase(ctx context.Context, buyer *auth.User, amount int) (*Transaction, error) {
	if buyer.CoinBalance < amount {
		return nil, errors.New("not enough coins")
	}
	t := Transaction{
		ID:        0,
		FromUser:  buyer,
		Amount:    amount,
		Type:      Purchase,
		CreatedAt: time.Now(),
	}

	return s.transactions.SaveTransaction(ctx, &t)
}

func (s *service) GetUserByUsername(ctx context.Context, username string) (*auth.User, error) {
	return s.authService.GetUserByUsername(ctx, username)
}

func (s *service) ListTransfers(ctx context.Context, user *auth.User) (incoming, outgoing []*Transaction, err error) {
	var (
		g             errgroup.Group
		inErr, outErr error
	)

	g.Go(func() error {
		incoming, inErr = s.transactions.GetIncomingTransfers(ctx, user.ID)
		if inErr != nil {
			if errors.Is(inErr, common.ErrNotFound) {
				incoming = make([]*Transaction, 0)
				return nil
			}
			return inErr
		}
		return nil
	})
	g.Go(func() error {
		outgoing, outErr = s.transactions.GetOutgoingTransfers(ctx, user.ID)
		if outErr != nil {
			if errors.Is(outErr, common.ErrNotFound) {
				outgoing = make([]*Transaction, 0)
				return nil
			}
			return outErr
		}
		return nil
	})

	if err = g.Wait(); err == nil {
		return incoming, outgoing, nil
	}
	return nil, nil, err
}
