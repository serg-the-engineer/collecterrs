package errsDummy

import (
	"your-company.com/project/pkg/errs"
)

var (
	DummyError                = errs.NewServiceError("DummyError", errs.TypeUserRelatedError, "Ы")
	FromVar1Error             = errs.NewServiceError("FromVar1Error", errs.TypeUserRelatedError, "Ы")
	FromVar2Error             = errs.NewServiceError("FromVar2Error", errs.TypeUserRelatedError, "Ы")
	FromDepthError            = errs.NewServiceError("FromDepthError", errs.TypeUserRelatedError, "Ы")
	FromStorageHandledError   = errs.NewServiceError("FromStorageHandledError", errs.TypeUserRelatedError, "Ы")
	FromStorageUnhandledError = errs.NewServiceError("FromStorageUnhandledError", errs.TypeUserRelatedError, "Ы")
	WithDetailsError          = errs.NewServiceError("WithDetailsError", errs.TypeUserRelatedError, "Ы")
)
