package service

import "errors"

var (
	ErrStudentNotFound = errors.New("student not found")
	ErrUserNotFound    = errors.New("user not found")
)
