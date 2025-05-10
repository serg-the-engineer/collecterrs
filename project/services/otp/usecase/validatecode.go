package usecase

import (
	"context"
	"errors"
	"your-company.com/project/errs/errsOtp"
	"your-company.com/project/pkg/otp"

	"your-company.com/project/services/otp/entity"
)

// ValidateCode выполняет проверку переданного кода OTP
func (u *useCasesImpl) ValidateCode(ctx context.Context, req *entity.ValidateCodeReq) (*entity.ValidateCode, error) {
	otpRequest, err := u.Providers.ProviderOtp.GetOtpRequestByAttemptID(ctx, req.AttemptID)
	if err != nil {
		if errors.Is(err, otp.ErrAttemptNotFound) {
			return nil, errsOtp.AttemptNotFoundError
		}
		return nil, err
	}

	isValid, err := u.Providers.ProviderOtp.ValidateCode(ctx, otpRequest, req.Code)
	if err != nil {
		if errors.Is(err, otp.ErrInvalidCode) {
			return nil, errsOtp.InvalidCodeError
		}
		if errors.Is(err, otp.ErrMaxCodeChecksExceeded) {
			return nil, errsOtp.MaxCodeChecksExceededError
		}
		return nil, err
	}
	if !isValid {
		return nil, errsOtp.InvalidCodeError
	}
	return &entity.ValidateCode{
		IsValid:     true,
		RetriesLeft: 0,
		Payload:     otpRequest.Payload,
	}, nil
}
