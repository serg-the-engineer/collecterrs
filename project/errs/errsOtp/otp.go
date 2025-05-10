package errsOtp

import (
	"your-company.com/project/pkg/errs"
)

const (
	MaxCodeAttemptsExceeded   = "MaxCodeAttemptsExceeded"
	InvalidCode               = "InvalidCode"
	MaxCodeChecksExceeded     = "MaxCodeChecksExceeded"
	NewAttemptTimeNotExceeded = "NewAttemptTimeNotExceeded"
	AttemptNotFoundInCache    = "AttemptNotFoundInCache"
)

var (
	MaxAttemptsExceededError       = errs.ServiceError(MaxCodeAttemptsExceeded, errs.TypeUserRelatedError, "Превышено количество запросов кода")
	InvalidCodeError               = errs.ServiceError(InvalidCode, errs.TypeUserRelatedError, "Некорректный код подтверждения")
	MaxCodeChecksExceededError     = errs.ServiceError(MaxCodeChecksExceeded, errs.TypeUserRelatedError, "Превышено количество проверок кода")
	NewAttemptTimeNotExceededError = errs.ServiceError(NewAttemptTimeNotExceeded, errs.TypeUserRelatedError, "Новый код возможен после ожидания")
	AttemptNotFoundError           = errs.ServiceError(AttemptNotFoundInCache, errs.TypeUserRelatedError, "Запрос для проверки кода не найден")
)
