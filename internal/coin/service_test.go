package coin

import (
	"avito-intern/internal/auth"
	"avito-intern/internal/common"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockAuthService struct {
	getUserByUsernameFunc func(ctx context.Context, username string) (*auth.User, error)
}

func (m *mockAuthService) AuthUser(_ context.Context, _, _ string) (auth.Token, error) {
	return "", nil
}

func (m *mockAuthService) GetUserFromToken(_ context.Context, _ auth.Token) (*auth.User, error) {
	return nil, nil
}

func (m *mockAuthService) GetUserByUsername(ctx context.Context, username string) (*auth.User, error) {
	return m.getUserByUsernameFunc(ctx, username)
}

type mockRepository struct {
	saveTransactionFunc  func(ctx context.Context, tx *Transaction) (*Transaction, error)
	getIncomingTransfers func(ctx context.Context, userID auth.UserID) ([]*Transaction, error)
	getOutgoingTransfers func(ctx context.Context, userID auth.UserID) ([]*Transaction, error)
}

func (m *mockRepository) SaveTransaction(ctx context.Context, tx *Transaction) (*Transaction, error) {
	return m.saveTransactionFunc(ctx, tx)
}

func (m *mockRepository) GetIncomingTransfers(ctx context.Context, userID auth.UserID) ([]*Transaction, error) {
	return m.getIncomingTransfers(ctx, userID)
}

func (m *mockRepository) GetOutgoingTransfers(ctx context.Context, userID auth.UserID) ([]*Transaction, error) {
	return m.getOutgoingTransfers(ctx, userID)
}

func TestTransfer_Success(t *testing.T) {
	fromUser := &auth.User{
		ID:          1,
		Username:    "from",
		CoinBalance: 100,
	}
	toUser := &auth.User{
		ID:          2,
		Username:    "to",
		CoinBalance: 50,
	}

	repo := &mockRepository{
		saveTransactionFunc: func(_ context.Context, tx *Transaction) (*Transaction, error) {
			tx.ID = 1
			return tx, nil
		},
	}

	svc := NewService(&mockAuthService{}, repo)

	tx, err := svc.Transfer(context.Background(), fromUser, toUser, 50)
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, TransactionID(1), tx.ID)
	assert.Equal(t, fromUser, tx.FromUser)
	assert.Equal(t, toUser, tx.ToUser)
	assert.Equal(t, 50, tx.Amount)
	assert.Equal(t, Transfer, tx.Type)
}

func TestTransfer_InsufficientFunds(t *testing.T) {
	fromUser := &auth.User{
		ID:          1,
		Username:    "from",
		CoinBalance: 40,
	}
	toUser := &auth.User{
		ID:          2,
		Username:    "to",
		CoinBalance: 50,
	}

	svc := NewService(&mockAuthService{}, &mockRepository{})

	tx, err := svc.Transfer(context.Background(), fromUser, toUser, 50)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "not enough coins")
}

func TestTransfer_MissingUsers(t *testing.T) {
	svc := NewService(&mockAuthService{}, &mockRepository{})

	tx, err := svc.Transfer(context.Background(), nil, nil, 50)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "missing required data")
}

func TestPurchase_Success(t *testing.T) {
	buyer := &auth.User{
		ID:          1,
		Username:    "buyer",
		CoinBalance: 100,
	}

	repo := &mockRepository{
		saveTransactionFunc: func(_ context.Context, tx *Transaction) (*Transaction, error) {
			tx.ID = 1
			return tx, nil
		},
	}

	svc := NewService(&mockAuthService{}, repo)

	tx, err := svc.Purchase(context.Background(), buyer, 50)
	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, TransactionID(1), tx.ID)
	assert.Equal(t, buyer, tx.FromUser)
	assert.Nil(t, tx.ToUser)
	assert.Equal(t, 50, tx.Amount)
	assert.Equal(t, Purchase, tx.Type)
}

func TestPurchase_InsufficientFunds(t *testing.T) {
	buyer := &auth.User{
		ID:          1,
		Username:    "buyer",
		CoinBalance: 40,
	}

	svc := NewService(&mockAuthService{}, &mockRepository{})

	tx, err := svc.Purchase(context.Background(), buyer, 50)
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "not enough coins")
}

func TestListTransfers(t *testing.T) {
	user := &auth.User{
		ID:       1,
		Username: "user",
	}

	now := time.Now()
	incomingTx := []*Transaction{
		{
			ID:        1,
			FromUser:  &auth.User{ID: 2},
			ToUser:    user,
			Amount:    50,
			Type:      Transfer,
			CreatedAt: now,
		},
	}

	outgoingTx := []*Transaction{
		{
			ID:        2,
			FromUser:  user,
			ToUser:    &auth.User{ID: 3},
			Amount:    30,
			Type:      Transfer,
			CreatedAt: now,
		},
	}

	repo := &mockRepository{
		getIncomingTransfers: func(_ context.Context, _ auth.UserID) ([]*Transaction, error) {
			return incomingTx, nil
		},
		getOutgoingTransfers: func(_ context.Context, _ auth.UserID) ([]*Transaction, error) {
			return outgoingTx, nil
		},
	}

	svc := NewService(&mockAuthService{}, repo)

	incoming, outgoing, err := svc.ListTransfers(context.Background(), user)
	assert.NoError(t, err)
	assert.Equal(t, incomingTx, incoming)
	assert.Equal(t, outgoingTx, outgoing)
}

func TestListTransfers_NotFound(t *testing.T) {
	user := &auth.User{
		ID:       1,
		Username: "user",
	}

	repo := &mockRepository{
		getIncomingTransfers: func(_ context.Context, _ auth.UserID) ([]*Transaction, error) {
			return nil, common.ErrNotFound
		},
		getOutgoingTransfers: func(_ context.Context, _ auth.UserID) ([]*Transaction, error) {
			return nil, common.ErrNotFound
		},
	}

	svc := NewService(&mockAuthService{}, repo)

	incoming, outgoing, err := svc.ListTransfers(context.Background(), user)
	assert.NoError(t, err)
	assert.Empty(t, incoming)
	assert.Empty(t, outgoing)
}

func TestListTransfers_Error(t *testing.T) {
	user := &auth.User{
		ID:       1,
		Username: "user",
	}

	expectedErr := errors.New("database error")
	repo := &mockRepository{
		getIncomingTransfers: func(_ context.Context, _ auth.UserID) ([]*Transaction, error) {
			return nil, expectedErr
		},
		getOutgoingTransfers: func(_ context.Context, _ auth.UserID) ([]*Transaction, error) {
			return nil, expectedErr
		},
	}

	svc := NewService(&mockAuthService{}, repo)

	incoming, outgoing, err := svc.ListTransfers(context.Background(), user)
	assert.Error(t, err)
	assert.Nil(t, incoming)
	assert.Nil(t, outgoing)
}

func TestGetUserByUsername(t *testing.T) {
	expectedUser := &auth.User{
		ID:       1,
		Username: "testuser",
	}

	authSvc := &mockAuthService{
		getUserByUsernameFunc: func(_ context.Context, username string) (*auth.User, error) {
			if username == "testuser" {
				return expectedUser, nil
			}
			return nil, auth.ErrUserNotFound
		},
	}

	svc := NewService(authSvc, &mockRepository{})

	user, err := svc.GetUserByUsername(context.Background(), "testuser")
	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)

	user, err = svc.GetUserByUsername(context.Background(), "nonexistent")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.ErrorIs(t, err, auth.ErrUserNotFound)
}
