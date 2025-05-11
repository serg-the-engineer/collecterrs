package errsOtp

import (
	"your-company.com/project/pkg/errs"
)

var (
	MaxAttemptsExceededError       = errs.NewServiceError("MaxAttemptsExceeded", errs.TypeUserRelatedError, "Превышено количество запросов кода")
	InvalidCodeError               = errs.NewServiceError("InvalidCode", errs.TypeUserRelatedError, "Некорректный код подтверждения")
	MaxCodeChecksExceededError     = errs.NewServiceError("MaxCodeChecksExceeded", errs.TypeUserRelatedError, "Превышено количество проверок кода")
	NewAttemptTimeNotExceededError = errs.NewServiceError("NewAttemptTimeNotExceeded", errs.TypeUserRelatedError, "Новый код возможен после ожидания")
	AttemptNotFoundError           = errs.NewServiceError("AttemptNotFound", errs.TypeUserRelatedError, "Запрос для проверки кода не найден")
)
