package cache

import "errors"

var (
	ErrMarksNotFound = errors.New("marks not found")
	ErrTokenNotFound = errors.New("token not found")
)
