package otp

import "errors"

var (
	ErrMaxAttemptsExceeded       = errors.New("max code attempts exceeded")
	ErrNewAttemptTimeNotExceeded = errors.New("new attempt time not exceeded")
	ErrMaxCodeChecksExceeded     = errors.New("max code checks exceeded")
	ErrInvalidCode               = errors.New("invalid code")
	ErrAttemptNotFound           = errors.New("attempt not found")
)
