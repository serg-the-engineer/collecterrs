package errsUsers

import (
	"your-company.com/project/pkg/errs"
)

var (
	UserNotFoundError = errs.NewServiceError("UserNotFound", errs.TypeUserRelatedError, "Пользователь не найден")
	UserBlockedError  = errs.NewServiceError("UserBlocked", errs.TypeUserRelatedError, "Пользователь заблокирован")
)
