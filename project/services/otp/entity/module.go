package entity

import (
	"time"

	o "your-company.com/project/pkg/otp"
)

type (
	OtpRequest struct {
		Initiator       string
		Action          string
		Payload         []byte
		ValidUntil      time.Time
		NewAttemptUntil time.Time // Время до запроса нового кода в секундах

		LastAttemptID  string
		AttemptsCount  int // Число сгенерированных кодов (попыток) для подтверждения
		CodeValidUntil time.Time

		Code            string
		CodeChecksCount int // Число проверок введенного кода в рамках одной попытки
	}
)

func (o *OtpRequest) Convert(req *o.Request) {
	if req != nil {
		o.Initiator = req.Initiator
		o.Action = req.Action
		o.Payload = req.Payload
		o.ValidUntil = req.ValidUntil

		o.LastAttemptID = req.LastAttemptID
		o.AttemptsCount = req.AttemptsCount
		o.CodeValidUntil = req.CodeValidUntil

		o.Code = req.Code
		o.CodeChecksCount = req.CodeChecksCount
		o.NewAttemptUntil = req.NewAttemptUntil
	}
}
