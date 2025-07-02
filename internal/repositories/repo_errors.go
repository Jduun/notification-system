package repositories

import "errors"

var (
	ErrMaxBatchSizeExceeded = errors.New("batch size exceeds max allowed limit")
	ErrNotFound             = errors.New("not found")
)
