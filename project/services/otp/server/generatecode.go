package server

import (
	"context"

	pb "your-company.com/project/specs/proto/otp"

	"your-company.com/project/services/otp/entity"
)

func (s *Server) GenerateCode(ctx context.Context, req *pb.GenerateCodeReq) (*pb.GenerateCodeResp, error) {
	generateCodeEntity := entity.MakeGenerateCodePbToEntity(req)

	generateCode, err := s.useCase.GenerateCode(ctx, generateCodeEntity)
	if err != nil {
		return nil, err
	}

	return entity.MakeGenerateCodeEntityToPb(generateCode, s.cfg.App), nil
}
