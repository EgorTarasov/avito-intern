package common

import (
	"errors"
	"fmt"
)

var (
	Err         = errors.New("common")
	ErrNotFound = fmt.Errorf("%v: запись не найдена", Err)
)
