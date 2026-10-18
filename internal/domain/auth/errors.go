package auth

import "errors"

var ErrNotFound = errors.New("not found")

type BusinessError struct {
	Code   string
	Detail string
}

func (e *BusinessError) Error() string { return e.Code + ": " + e.Detail }
