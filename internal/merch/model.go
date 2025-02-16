package merch

import (
	"avito-intern/internal/auth"
	"time"
)

type Merch struct {
	ID    int64
	Name  string
	Price int
}

type Purchase struct {
	ID          int
	UserID      auth.UserID
	MerchID     int64
	MerchName   string
	Quantity    int
	PurchasedAt time.Time
}
