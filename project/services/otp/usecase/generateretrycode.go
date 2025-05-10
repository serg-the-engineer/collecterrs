package usecase

import (
	"context"
	"errors"
	"your-company.com/project/errs/errsOtp"
	"your-company.com/project/pkg/otp"

	"your-company.com/project/services/otp/entity"
)

// GenerateRetryCode перевыпускает OTP для ранее созданной попытки
func (u *useCasesImpl) GenerateRetryCode(ctx context.Context, req *entity.GenerateRetryCodeReq) (*entity.GenerateRetryCode, error) {
	otpRequest, err := u.Providers.ProviderOtp.GetOtpRequestByAttemptID(ctx, req.AttemptID)
	if err != nil {
		if errors.Is(err, otp.ErrAttemptNotFound) {
			return nil, errsOtp.AttemptNotFoundError
		}
		return nil, err
	}

	otpRequest, err = u.Providers.ProviderOtp.CreateNewAttempt(ctx, otpRequest)
	if err != nil {
		if errors.Is(err, otp.ErrMaxAttemptsExceeded) {
			return nil, errsOtp.MaxAttemptsExceededError
		}
		if errors.Is(err, otp.ErrNewAttemptTimeNotExceeded) {
			return nil, errsOtp.NewAttemptTimeNotExceededError
		}
		return nil, err
	}

	result := &entity.OtpRequest{}
	result.Convert(otpRequest)

	return &entity.GenerateRetryCode{OtpRequest: *result}, nil
}
