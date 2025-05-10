package entity

import (
	cfgOtp "your-company.com/project/config/services/otp"

	pb "your-company.com/project/specs/proto/otp"
)

type (
	GenerateCodeReq struct {
		Action  string
		Payload []byte
	}

	GenerateCode struct {
		OtpRequest
	}
)

// MakeGenerateCodePbToEntity создает объект из pb.GenerateCodeReq в GenerateCodeReq
func MakeGenerateCodePbToEntity(req *pb.GenerateCodeReq) *GenerateCodeReq {
	return &GenerateCodeReq{
		Action:  req.Action,
		Payload: req.Payload,
	}
}

// MakeGenerateCodeEntityToPb создает объект из GenerateCode в pb.GenerateCodeResp
func MakeGenerateCodeEntityToPb(req *GenerateCode, config *cfgOtp.Application) *pb.GenerateCodeResp {
	//nolint:gosec
	return &pb.GenerateCodeResp{
		AttemptId:       req.LastAttemptID,
		Code:            req.Code,
		CodeTtl:         int32(config.CodeTTL.Seconds()),
		CodeChecksLeft:  int32(config.MaxCodeChecks),
		AttemptsLeft:    int32(config.MaxAttempts - req.AttemptsCount),
		AttemptsTimeout: int32(config.CodeTTL.Seconds()),
		NewAttemptDelay: int32(config.NewAttemptDelay.Seconds()),
	}
}
