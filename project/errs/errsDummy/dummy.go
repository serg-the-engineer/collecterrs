package errsDummy

import (
	"your-company.com/project/pkg/errs"
)

var (
	DummyError                = errs.ServiceError("DummyError", errs.TypeUserRelatedError, "Ы")
	FromVar1Error             = errs.ServiceError("FromVar1Error", errs.TypeUserRelatedError, "Ы")
	FromVar2Error             = errs.ServiceError("FromVar2Error", errs.TypeUserRelatedError, "Ы")
	FromDepthError            = errs.ServiceError("FromDepthError", errs.TypeUserRelatedError, "Ы")
	FromStorageHandledError   = errs.ServiceError("FromStorageHandledError", errs.TypeUserRelatedError, "Ы")
	FromStorageUnhandledError = errs.ServiceError("FromStorageUnhandledError", errs.TypeUserRelatedError, "Ы")
	WithDetailsError          = errs.ServiceError("WithDetailsError", errs.TypeUserRelatedError, "Ы")
)
