package coin

import (
	"avito-intern/internal/auth"
	"context"
)

type Repository interface {
	SaveTransaction(context.Context, *Transaction) (*Transaction, error)
	GetIncomingTransfers(context.Context, auth.UserID) ([]*Transaction, error)
	GetOutgoingTransfers(context.Context, auth.UserID) ([]*Transaction, error)
}
