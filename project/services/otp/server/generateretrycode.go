package server

import (
	"context"

	pb "your-company.com/project/specs/proto/otp"

	"your-company.com/project/services/otp/entity"
)

func (s *Server) GenerateRetryCode(ctx context.Context, req *pb.GenerateRetryCodeReq) (*pb.GenerateCodeResp, error) {
	generateRetryCodeEntity := entity.MakeGenerateRetryCodePbToEntity(req)

	generateRetryCode, err := s.useCase.GenerateRetryCode(ctx, generateRetryCodeEntity)
	if err != nil {
		return nil, err
	}

	return entity.MakeGenerateRetryCodeEntityToPb(generateRetryCode, s.cfg.App), nil
}
