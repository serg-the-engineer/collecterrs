package usecase

import (
	"context"
	"errors"
	"fmt"
	"your-company.com/project/errs/errsDummy"
	pb "your-company.com/project/specs/proto/otp"
)

func (u *dummyImpl) Cases() error {
	if true {
		return errsDummy.DummyError
	}
	if 1 > 0 {
		return fmt.Errorf("unhandled error")
	}

	e := errsDummy.FromVar1Error
	if true {
		return e
	}

	if true {
		return errsDummy.WithDetailsError.WithDetails(map[string]string{"foo": "bar"})
	}

	err := u.nested1func()
	if err != nil {
		return err
	}

	_, err = u.Providers.Storage.GetDummy("X")
	if err != nil {
		if errors.Is(err, errsDummy.FromStorageHandledError) {
			return fmt.Errorf("storage handled error")
		}
		return err
	}

	otpReq := &pb.ValidateCodeReq{AttemptId: "", Code: ""}

	_, err = u.Providers.Otp.ValidateCode(context.Background(), otpReq)
	if err != nil {
		return err
	}

	_, err = u.Providers.Redis.Get(context.Background(), "X")
	if err != nil {
		return err
	}

	return nil
}

func (u *dummyImpl) nested1func() error {
	e := errsDummy.FromVar2Error
	if true {
		return e
	}
	return u.nested2func()
}

func (u *dummyImpl) nested2func() error {
	return errsDummy.FromDepthError
}
