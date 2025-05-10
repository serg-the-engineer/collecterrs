package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"your-company.com/project/errs/errsUsers"

	pb "your-company.com/project/specs/proto/otp"

	"your-company.com/project/services/users/entity"
)

func (u *useCasesImpl) ConfirmLogin(ctx context.Context, req *entity.ConfirmLoginReq) (*entity.ConfirmLogin, error) {
	otpReq := &pb.ValidateCodeReq{
		AttemptId: req.AttemptID,
		Code:      req.Code,
	}

	resp, err := u.Providers.Otp.ValidateCode(ctx, otpReq)
	if err != nil {
		return nil, err
	}

	var loginPayload entity.OtpPayload
	if err = json.Unmarshal(resp.Payload, &loginPayload); err != nil {
		return nil, fmt.Errorf("users usecase ConfirmLogin payload unmarshall err: %w", err)
	}

	// получаем данные пользователя
	user, err := u.Providers.Storage.GetUser(ctx, loginPayload.Phone)
	if err != nil && !errors.Is(err, errsUsers.UserNotFoundError) {
		return nil, err
	}
	if errors.Is(err, errsUsers.UserNotFoundError) {
		// создаем нового пользователя
		user, err = u.Providers.Storage.CreateUser(ctx, entity.ParamsCreateUser{
			Status: entity.UserStatusActive,
			Phone:  loginPayload.Phone,
		})
		if err != nil {
			return nil, fmt.Errorf("users usecase ConfirmLogin create user err: %w", err)
		}
	}

	if user.IsBlocked() {
		return nil, errsUsers.UserBlockedError
	}

	return &entity.ConfirmLogin{
		AccessToken:  "accessToken",
		RefreshToken: "refreshToken",
	}, nil
}
