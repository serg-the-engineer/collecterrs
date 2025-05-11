package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"your-company.com/project/errs/errsUsers"

	"your-company.com/project/services/users/entity"
	"your-company.com/project/specs/proto/otp"
)

const OtpActionLogin = "login-"

func (u *useCasesImpl) Login(ctx context.Context, req *entity.LoginReq) (*entity.Login, error) {
	user, err := u.Providers.Storage.GetUser(ctx, req.Phone)
	if err != nil && !errors.Is(err, errsUsers.UserNotFoundError) {
		return nil, err
	}

	if user != nil && user.IsBlocked() {
		return nil, errsUsers.UserBlockedError
	}

	otpPayload, err := json.Marshal(entity.OtpPayload{
		Phone:  req.Phone,
		Device: req.Device,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка формирования payload для otp: %w", err)
	}
	otpReq := &otp.GenerateCodeReq{
		Action:  OtpActionLogin + req.Phone,
		Payload: otpPayload,
	}

	code, err := u.Providers.Otp.GenerateCode(ctx, otpReq)
	if err != nil {
		return nil, err
	}

	return entity.MakeLogin(code), nil
}
