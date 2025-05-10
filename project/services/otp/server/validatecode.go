package server

import (
	"context"

	pb "your-company.com/project/specs/proto/otp"

	"your-company.com/project/services/otp/entity"
)

func (s *Server) ValidateCode(ctx context.Context, req *pb.ValidateCodeReq) (*pb.ValidateCodeResp, error) {
	validateCodeEntity := entity.MakeValidateCodePbToEntity(req)

	validateCode, err := s.useCase.ValidateCode(ctx, validateCodeEntity)
	if err != nil {
		return nil, err
	}

	return entity.MakeValidateCodeEntityToPb(validateCode), nil
}
