package entity

import (
	cfgOtp "your-company.com/project/config/services/otp"
	pb "your-company.com/project/specs/proto/otp"
)

type (
	GenerateRetryCodeReq struct {
		AttemptID string
	}

	GenerateRetryCode struct {
		OtpRequest
	}
)

// MakeGenerateRetryCodePbToEntity создает объект из pb.GenerateRetryCodeReq в GenerateRetryCodeReq
func MakeGenerateRetryCodePbToEntity(req *pb.GenerateRetryCodeReq) *GenerateRetryCodeReq {
	return &GenerateRetryCodeReq{
		AttemptID: req.AttemptId,
	}
}

// MakeGenerateRetryCodeEntityToPb создает объект из GenerateRetryCode в pb.GenerateCodeResp
func MakeGenerateRetryCodeEntityToPb(req *GenerateRetryCode, config *cfgOtp.Application) *pb.GenerateCodeResp {
	//nolint:gosec
	return &pb.GenerateCodeResp{
		AttemptId:       req.LastAttemptID,
		Code:            req.Code,
		CodeTtl:         int32(config.CodeTTL),
		CodeChecksLeft:  int32(config.MaxCodeChecks),
		AttemptsLeft:    int32(config.MaxAttempts - req.AttemptsCount),
		AttemptsTimeout: int32(config.CodeTTL),
		NewAttemptDelay: int32(config.NewAttemptDelay),
	}
}
