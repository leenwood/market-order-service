package domain

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrEmptyCart         = errors.New("cart is empty")
	ErrCancelNotAllowed  = errors.New("cancel not allowed in current status")
)
