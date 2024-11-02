package repository

import "errors"

var (
	ErrNegativeBalance = errors.New("negative balance")
	ErrUnaffectedRows  = errors.New("failed to affect row(s)")
)
