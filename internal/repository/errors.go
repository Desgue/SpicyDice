package repository

import "errors"

var (
	ErrNegativeBalance   = errors.New("negative balance")
	ErrUnaffectedRows    = errors.New("failed to affect row(s)")
	ErrTransactionCommit = errors.New("failed to commit transaction")
)
