package entity

import (
	"your-company.com/project/specs/proto/otp"
	pb "your-company.com/project/specs/proto/users"
)

type (
	DeviceInfo struct {
		InstallationID string
		AppVersion     string
		DeviceModel    string
		SystemType     string
		SystemVersion  string
	}

	LoginReq struct {
		Phone  string
		Device *DeviceInfo
	}

	Login struct {
		AttemptID       string
		Code            string
		CodeTTL         int32
		CodeChecksLeft  int32
		AttemptsLeft    int32
		AttemptsTimeout int32
	}

	OtpPayload struct {
		Phone  string
		Device *DeviceInfo
	}
)

func MakeLogin(r *otp.GenerateCodeResp) *Login {
	return &Login{
		AttemptID:       r.AttemptId,
		Code:            r.Code,
		CodeTTL:         r.CodeTtl,
		CodeChecksLeft:  r.CodeChecksLeft,
		AttemptsLeft:    r.AttemptsLeft,
		AttemptsTimeout: r.AttemptsTimeout,
	}
}

// MakeLoginPbToEntity создает объект из pb.LoginReq в LoginReq
func MakeLoginPbToEntity(req *pb.LoginReq) *LoginReq {
	return &LoginReq{
		Phone:  req.Phone,
		Device: &DeviceInfo{},
	}
}

// MakeLoginEntityToPb создает объект из Login в pb.LoginResp
func MakeLoginEntityToPb(l *Login) *pb.LoginResp {
	return &pb.LoginResp{
		AttemptID: l.AttemptID,
		RetryTime: l.AttemptsTimeout,
	}
}
