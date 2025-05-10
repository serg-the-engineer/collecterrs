package otp

import "time"

type Config struct {
	Env             string
	OtpRequestTTL   time.Duration // Время жизни запроса ОТП для расчета лимитов попыток
	CodeTTL         time.Duration // Время жизни кода ОТП (попытки)
	NewAttemptDelay time.Duration // Время до запроса нового кода

	MaxAttempts   int // Допустимое количество попыток ОТП (попыток) в рамках жизни одного запроса ОТП
	MaxCodeChecks int // Допустимое количество проверок совпадения кода для одной попытки
}
