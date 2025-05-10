package otp

import (
	"fmt"
	"time"
)

const (
	otpTestCode = "1111"
	envDev      = "dev"
)

type (
	CodeGeneratorImpl struct {
		env string
	}

	CodeGenerator interface {
		Generate() string
	}
)

func NewCodeGenerator(env string) CodeGenerator {
	return &CodeGeneratorImpl{env: env}
}

// Generate формирует символьный код OTP для проверки пользовательского действия
func (g *CodeGeneratorImpl) Generate() string {
	if g.env == envDev {
		return otpTestCode
	}
	length := 4
	return fmt.Sprint(time.Now().Nanosecond())[:length]
}
