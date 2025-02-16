package merch

import (
	"avito-intern/internal/auth"
	"avito-intern/internal/coin"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService is a mock implementation of the auth.Service interface.
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) AuthUser(ctx context.Context, username, password string) (auth.Token, error) {
	args := m.Called(ctx, username, password)
	return args.Get(0).(auth.Token), args.Error(1)
}

func (m *MockAuthService) GetUserFromToken(ctx context.Context, rawToken auth.Token) (*auth.User, error) {
	args := m.Called(ctx, rawToken)
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockAuthService) GetUserByUsername(ctx context.Context, username string) (*auth.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*auth.User), args.Error(1)
}

// MockCoinService is a mock implementation of the coin.Service interface.
type MockCoinService struct {
	mock.Mock
}

func (m *MockCoinService) GetUserByUsername(ctx context.Context, username string) (*auth.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockCoinService) Transfer(ctx context.Context, from, to *auth.User, amount int) (*coin.Transaction, error) {
	args := m.Called(ctx, from, to, amount)
	return args.Get(0).(*coin.Transaction), args.Error(1)
}

func (m *MockCoinService) Purchase(ctx context.Context, user *auth.User, amount int) (*coin.Transaction, error) {
	args := m.Called(ctx, user, amount)
	return args.Get(0).(*coin.Transaction), args.Error(1)
}

func (m *MockCoinService) ListTransfers(ctx context.Context, user *auth.User) (incoming, outgoing []*coin.Transaction, err error) {
	args := m.Called(ctx, user)
	return args.Get(0).([]*coin.Transaction), args.Get(1).([]*coin.Transaction), args.Error(2)
}

// MockRepository is a mock implementation of the merch.Repository interface.
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetMerchByID(ctx context.Context, merchName string) (*Merch, error) {
	args := m.Called(ctx, merchName)
	return args.Get(0).(*Merch), args.Error(1)
}

func (m *MockRepository) SavePurchase(ctx context.Context, purchase *Purchase) error {
	args := m.Called(ctx, purchase)
	return args.Error(0)
}

func (m *MockRepository) ListPurchasesByUserID(ctx context.Context, userID auth.UserID) ([]*Purchase, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*Purchase), args.Error(1)
}

func TestService_Purchase(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockCoinService := new(MockCoinService)
	mockRepo := new(MockRepository)

	service := NewService(mockAuthService, mockCoinService, mockRepo)

	user := &auth.User{ID: 1, CoinBalance: 100}
	merchItem := &Merch{ID: 1, Name: "T-Shirt", Price: 50}

	mockRepo.On("GetMerchByID", mock.Anything, merchItem.Name).Return(merchItem, nil)
	mockCoinService.On("Purchase", mock.Anything, user, merchItem.Price).Return(&coin.Transaction{}, nil)
	mockRepo.On("SavePurchase", mock.Anything, mock.AnythingOfType("*merch.Purchase")).Return(nil)

	err := service.Purchase(context.Background(), user, merchItem.Name)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockCoinService.AssertExpectations(t)
}

func TestService_ListPurchases(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockCoinService := new(MockCoinService)
	mockRepo := new(MockRepository)

	service := NewService(mockAuthService, mockCoinService, mockRepo)

	user := &auth.User{ID: 1}
	purchases := []*Purchase{
		{ID: 1, UserID: user.ID, MerchID: 1, Quantity: 1, PurchasedAt: time.Now()},
		{ID: 2, UserID: user.ID, MerchID: 2, Quantity: 2, PurchasedAt: time.Now()},
	}

	mockRepo.On("ListPurchasesByUserID", mock.Anything, user.ID).Return(purchases, nil)

	result, err := service.ListPurchases(context.Background(), user)
	assert.NoError(t, err)
	assert.Equal(t, purchases, result)
	mockRepo.AssertExpectations(t)
}

func TestService_ListTransfers(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockCoinService := new(MockCoinService)
	mockRepo := new(MockRepository)

	service := NewService(mockAuthService, mockCoinService, mockRepo)

	user := &auth.User{ID: 1}
	incoming := []*coin.Transaction{
		{ID: 1, FromUser: &auth.User{Username: "user1"}, Amount: 100},
	}
	outgoing := []*coin.Transaction{
		{ID: 2, ToUser: &auth.User{Username: "user2"}, Amount: 50},
	}

	mockCoinService.On("ListTransfers", mock.Anything, user).Return(incoming, outgoing, nil)

	in, out, err := service.ListTransfers(context.Background(), user)
	assert.NoError(t, err)
	assert.Equal(t, incoming, in)
	assert.Equal(t, outgoing, out)
	mockCoinService.AssertExpectations(t)
}
