package errsUsers

import (
	"your-company.com/project/pkg/errs"
)

var (
	UserNotFoundError = errs.ServiceError("UserNotFound", errs.TypeUserRelatedError, "Пользователь не найден")
	UserBlockedError  = errs.ServiceError("UserBlocked", errs.TypeUserRelatedError, "Пользователь заблокирован")
)
