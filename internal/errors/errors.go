package errors

import "errors"

var (
	ErrNoConnectionToDb  = errors.New("missing connection to DB")
	ErrNoRepository      = errors.New("missing repository")
	ErrNoController      = errors.New("missing controller")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrBadRequest        = errors.New("wrong data")
)
