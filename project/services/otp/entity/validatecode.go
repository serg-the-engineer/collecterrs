package entity

import (
	pb "your-company.com/project/specs/proto/otp"
)

type (
	ValidateCodeReq struct {
		AttemptID string
		Code      string
	}

	ValidateCode struct {
		Initiator   string
		IsValid     bool
		RetriesLeft int
		Payload     []byte
	}
)

// MakeValidateCodePbToEntity создает объект из pb.ValidateCodeReq в ValidateCodeReq
func MakeValidateCodePbToEntity(req *pb.ValidateCodeReq) *ValidateCodeReq {
	return &ValidateCodeReq{
		AttemptID: req.AttemptId,
		Code:      req.Code,
	}
}

// MakeValidateCodeEntityToPb создает объект из ValidateCode в pb.ValidateCodeResp
func MakeValidateCodeEntityToPb(req *ValidateCode) *pb.ValidateCodeResp {
	return &pb.ValidateCodeResp{
		Success:     req.IsValid,
		Initiator:   req.Initiator,
		RetriesLeft: int32(req.RetriesLeft), //nolint: gosec //not critical
		Payload:     req.Payload,
	}
}
