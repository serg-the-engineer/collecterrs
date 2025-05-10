package otp

import (
	"time"
)

type (
	Request struct {
		Initiator       string
		Action          string
		Payload         []byte
		NewAttemptUntil time.Time // Время до запроса нового кода в секундах
		ValidUntil      time.Time

		LastAttemptID  string
		AttemptsCount  int // Число сгенерированных кодов (попыток) для подтверждения
		CodeValidUntil time.Time

		Code            string
		CodeChecksCount int // Число проверок введенного кода в рамках одной попытки
	}
)
