package entity

import (
	pb "your-company.com/project/specs/proto/users"
)

type (
	ConfirmLoginReq struct {
		AttemptID string
		Code      string
	}

	ConfirmLogin struct {
		AccessToken  string
		RefreshToken string
		NextStep     string
	}
)

// MakeConfirmLoginPbToEntity создает объект из pb.ConfirmLoginReq в ConfirmLoginReq
func MakeConfirmLoginPbToEntity(req *pb.ConfirmLoginReq) *ConfirmLoginReq {
	return &ConfirmLoginReq{
		AttemptID: req.AttemptID,
		Code:      req.Code,
	}
}

// MakeConfirmLoginEntityToPb создает объект из ConfirmLogin в pb.ConfirmLoginResp
func MakeConfirmLoginEntityToPb(res *ConfirmLogin) *pb.ConfirmLoginResp {
	return &pb.ConfirmLoginResp{
		NextStep: res.NextStep,
		Token: &pb.Token{
			Access:  res.AccessToken,
			Refresh: res.RefreshToken,
		},
	}
}
