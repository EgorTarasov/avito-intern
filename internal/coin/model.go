package coin

import (
	"avito-intern/internal/auth"
	"time"
)

type TransactionID int64

type Type string

const (
	Purchase Type = "purchase"
	Transfer Type = "transfer"
)

type Transaction struct {
	ID              TransactionID
	FromUser        *auth.User
	ToUser          *auth.User
	Amount          int
	Type            Type
	PrevTransaction *int64
	CreatedAt       time.Time
}
