package storage

import "errors"

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrStudentNotFound  = errors.New("student not found")
	ErrStudentExists    = errors.New("student already exists")
	ErrRelationNotFound = errors.New("relation not found")
	ErrFailedToGetTX    = errors.New("failed to get the transaction")
)
