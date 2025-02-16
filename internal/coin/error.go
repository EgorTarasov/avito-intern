package coin

import (
	"errors"
	"fmt"
)

var (
	Err                 = errors.New("coin")
	ErrInvalidRecipient = fmt.Errorf("%v: invalid recipient users can't send coins to them self", Err)
	ErrNotEnoughCoins   = fmt.Errorf("%v: not enough coins for transfer", Err)
)

type ErrInvalidTransactionID struct {
	itemID int64
}

func (e ErrInvalidTransactionID) Error() string {
	return fmt.Sprintf("%v: недопустимый transaction %d: должен быть > 0", Err, e.itemID)
}

func NewErrInvalidItemID(itemID int64) error {
	return &ErrInvalidTransactionID{itemID: itemID}
}
