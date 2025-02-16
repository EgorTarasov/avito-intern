package merch

import (
	"errors"
	"fmt"
)

var (
	Err                  = errors.New("merch")
	ErrPurchasesNotFound = fmt.Errorf("%v: purchases not found", Err)
	ErrMerchNotFound     = fmt.Errorf("%v: merch not found", Err)
)
