package usecase

import (
	"context"
	"errors"
	"your-company.com/project/errs/errsOtp"

	"your-company.com/project/pkg/otp"
	"your-company.com/project/services/otp/entity"
)

// GenerateCode формирует OTP для проверки пользовательского действия
func (u *useCasesImpl) GenerateCode(ctx context.Context, req *entity.GenerateCodeReq) (*entity.GenerateCode, error) {
	//return nil, errsOtp.MaxAttemptsExceededError
	var otpRequest *otp.Request

	otpRequest, err := u.Providers.ProviderOtp.GetOtpRequestByAction(ctx, "user", req.Action)
	if err != nil {
		return nil, err
	}
	if otpRequest == nil {
		otpRequest, err = u.Providers.ProviderOtp.CreateNewOtp(ctx, "user", req.Action, req.Payload)
		if err != nil {
			return nil, err
		}
	} else {
		otpRequest, err = u.Providers.ProviderOtp.CreateNewAttempt(ctx, otpRequest)
		if err != nil {
			if errors.Is(err, otp.ErrMaxAttemptsExceeded) {
				return nil, errsOtp.MaxAttemptsExceededError.WithDetails(map[string]string{"max": "не более 3х попыток!"})
			}
			if errors.Is(err, otp.ErrNewAttemptTimeNotExceeded) {
				return nil, errsOtp.NewAttemptTimeNotExceededError
			}
		}
	}

	result := &entity.OtpRequest{}
	result.Convert(otpRequest)

	return &entity.GenerateCode{OtpRequest: *result}, nil
}
