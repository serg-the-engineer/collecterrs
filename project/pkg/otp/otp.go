// Package otp реализует генерацию и проверку одноразовых паролей (OTP) с использованием кеша (Redis).
// Он предоставляет интерфейс для создания OTP-запросов, проверки кодов и управления попытками.
package otp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"your-company.com/project/pkg/memorystore"
)

const (
	otpRequestKeyFormat     = "otp-request-%s-%s"
	otpAttemptRequestFormat = "otp-attempt-%s-request-key"
)

type (
	// ProviderOtpImpl - Основная структура, которая реализует интерфейс ProviderOtp для работы с OTP.
	ProviderOtpImpl struct {
		Cfg           Config
		Redis         Redis
		CodeGenerator CodeGenerator
	}

	Redis interface {
		Set(ctx context.Context, key string, value any, expiration time.Duration) error
		Get(ctx context.Context, key string) (*memorystore.Value, error)
		GetList(ctx context.Context, keys ...string) ([]*memorystore.Value, error)
		Delete(ctx context.Context, keys ...string) (int, error)
		Close(ctx context.Context)
	}

	ProviderOtp interface {
		GetOtpRequestByAction(ctx context.Context, initiator, action string) (*Request, error)
		GetOtpRequestByAttemptID(ctx context.Context, attemptID string) (*Request, error)
		CreateNewOtp(ctx context.Context, initiator, action string, payload []byte) (*Request, error)
		CreateNewAttempt(ctx context.Context, otpRequest *Request) (*Request, error)
		ValidateCode(ctx context.Context, otpRequest *Request, code string) (bool, error)
	}
)

func NewProviderOtp(cfg Config, redis Redis) ProviderOtp {
	return &ProviderOtpImpl{
		Cfg:           cfg,
		Redis:         redis,
		CodeGenerator: NewCodeGenerator(cfg.Env),
	}
}

// GetOtpRequestByAction получает OtpRequest соответствующий конкретному действию пользователя action.
func (o *ProviderOtpImpl) GetOtpRequestByAction(ctx context.Context, initiator, action string) (*Request, error) {
	var otpRequest Request
	otpRequestKey := fmt.Sprintf(otpRequestKeyFormat, action, initiator)
	v, err := o.Redis.Get(ctx, otpRequestKey)
	if err != nil && !errors.Is(err, memorystore.ErrKeyNotFound) {
		return nil, err
	}
	if v != nil {
		err = v.Struct(&otpRequest)
		if err != nil {
			return nil, err
		}
	}
	if v == nil || otpRequest.ValidUntil.Before(time.Now()) {
		// Может быть невалидный otp в кеше, т.к. каждый раз при новой попытке мы обновляем ttl
		return nil, nil
	}
	return &otpRequest, nil
}

// GetOtpRequestByAttemptID получает OtpRequest соответствующей конкретной попытке.
func (o *ProviderOtpImpl) GetOtpRequestByAttemptID(ctx context.Context, attemptID string) (*Request, error) {
	var otpRequest Request

	otpAttemptKey := fmt.Sprintf(otpAttemptRequestFormat, attemptID)
	v, err := o.Redis.Get(ctx, otpAttemptKey)
	if err != nil {
		if errors.Is(err, memorystore.ErrKeyNotFound) {
			return nil, ErrAttemptNotFound
		}
		return nil, fmt.Errorf("ошибка получения данных otpAttempt из кеша: %v", err)
	}
	otpRequestKey, err := v.String()
	if err != nil {
		return nil, fmt.Errorf("невалидные данные otpAttempt получены из кеша: %v", err)
	}
	v, err = o.Redis.Get(ctx, otpRequestKey)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения данных otpRequest из кеша: %v", err)
	}
	err = v.Struct(&otpRequest)
	if err != nil {
		return nil, fmt.Errorf("невалидные данные otpRequest получены из кеша: %v", err)
	}
	if otpRequest.ValidUntil.Before(time.Now()) {
		return nil, fmt.Errorf(
			"для актуальной попытки %s найдены данные по неактуальному запросу ОТП [%s %s]",
			attemptID, otpRequest.Action, otpRequest.Initiator)
	}
	return &otpRequest, nil
}

// CreateNewOtp генерирует сущности OtpRequest и OtpAttempt и сохраняет их в кеш.
func (o *ProviderOtpImpl) CreateNewOtp(ctx context.Context, initiator, action string, payload []byte) (*Request, error) {
	attemptID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации attemptID: %v", err)
	}

	otpRequest := &Request{
		Action:          action,
		Initiator:       initiator,
		Payload:         payload,
		ValidUntil:      time.Now().Add(o.Cfg.OtpRequestTTL),
		NewAttemptUntil: time.Now().Add(o.Cfg.NewAttemptDelay),
		LastAttemptID:   attemptID.String(),
		AttemptsCount:   1,
		Code:            o.CodeGenerator.Generate(),
		CodeChecksCount: 0,
		CodeValidUntil:  time.Now().Add(o.Cfg.CodeTTL),
	}

	err = o.saveOtpRequest(ctx, otpRequest)
	if err != nil {
		return nil, err
	}
	return otpRequest, nil
}

// CreateNewAttempt генерирует новую попытку (attempt) для существующего OtpRequest и обновляет данные в кеше.
func (o *ProviderOtpImpl) CreateNewAttempt(ctx context.Context, otpRequest *Request) (*Request, error) {
	if otpRequest.AttemptsCount >= o.Cfg.MaxAttempts {
		return nil, ErrMaxAttemptsExceeded
	}

	if time.Now().Before(otpRequest.NewAttemptUntil) {
		return nil, ErrNewAttemptTimeNotExceeded
	}

	prevAttemptID := otpRequest.LastAttemptID
	attemptID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации attemptID: %v", err)
	}

	otpRequest.AttemptsCount++
	otpRequest.LastAttemptID = attemptID.String()
	otpRequest.Code = o.CodeGenerator.Generate()
	otpRequest.CodeChecksCount = 0
	otpRequest.CodeValidUntil = time.Now().Add(o.Cfg.CodeTTL)
	otpRequest.NewAttemptUntil = time.Now().Add(o.Cfg.NewAttemptDelay)

	err = o.saveOtpRequest(ctx, otpRequest)
	if err != nil {
		return nil, err
	}
	err = o.invalidateAttempt(ctx, prevAttemptID)
	if err != nil {
		return nil, err
	}
	return otpRequest, nil
}

// ValidateCode проверяет соответствие переданного кода OTP в конкретной попытке.
// В зависимости от результата вносит нужные изменения в кеш.
func (o *ProviderOtpImpl) ValidateCode(ctx context.Context, otpRequest *Request, code string) (bool, error) {
	if otpRequest.CodeValidUntil.Before(time.Now()) {
		return false, ErrInvalidCode
	}
	if otpRequest.CodeChecksCount >= o.Cfg.MaxCodeChecks {
		return false, ErrMaxCodeChecksExceeded
	}
	if otpRequest.Code == code {
		err := o.invalidateOtpRequest(ctx, otpRequest)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	err := o.incrementCodeChecks(ctx, otpRequest)
	if err != nil {
		return false, err
	}
	return false, nil
}

// SaveOtpRequest сохраняет сущность OtpRequest в кеш, а также ключ для связи по attemptID
func (o *ProviderOtpImpl) saveOtpRequest(ctx context.Context, otpRequest *Request) error {
	otpRequestKey := fmt.Sprintf(otpRequestKeyFormat, otpRequest.Action, otpRequest.Initiator)
	err := o.Redis.Set(ctx, otpRequestKey, otpRequest, o.Cfg.OtpRequestTTL)
	if err != nil {
		return fmt.Errorf("ошибка сохранения значения OtpRequest в redis: %v", err)
	}

	// устанавливаем связь с request для возможности делать retry по одному attemptID
	attemptKey := fmt.Sprintf(otpAttemptRequestFormat, otpRequest.LastAttemptID)
	// попытка живет дольше кода, чтобы можно было сделать retry просроченной
	err = o.Redis.Set(ctx, attemptKey, otpRequestKey, o.Cfg.OtpRequestTTL)
	if err != nil {
		return fmt.Errorf("ошибка сохранения значения OtpAttempt в redis: %v", err)
	}
	return nil
}

// InvalidateAttempt удаляет информацию о попытке из кеша.
func (o *ProviderOtpImpl) invalidateAttempt(ctx context.Context, prevAttemptID string) error {
	attemptKey := fmt.Sprintf(otpAttemptRequestFormat, prevAttemptID)
	_, err := o.Redis.Delete(ctx, attemptKey)
	if err != nil && !errors.Is(err, memorystore.ErrKeyNotFound) {
		return fmt.Errorf("ошибка удаления значения OtpAttempt из redis")
	}
	return nil
}

// InvalidateOtpRequest удаляет всю информацию о запросе OTP из кеша.
func (o *ProviderOtpImpl) invalidateOtpRequest(ctx context.Context, otpRequest *Request) error {
	err := o.invalidateAttempt(ctx, otpRequest.LastAttemptID)
	if err != nil {
		return err
	}
	otpRequestKey := fmt.Sprintf(otpRequestKeyFormat, otpRequest.Action, otpRequest.Initiator)
	deleted, err := o.Redis.Delete(ctx, otpRequestKey)
	if err != nil {
		return fmt.Errorf("ошибка удаления значения otpRequest из redis")
	}
	if deleted != 1 {
		return fmt.Errorf("ошибка удаления значения из redis. удалено %d записей вместо 1", deleted)
	}
	return nil
}

// incrementCodeChecks сохраняет в кеш неуспешный факт сверки кода.
func (o *ProviderOtpImpl) incrementCodeChecks(ctx context.Context, otpRequest *Request) error {
	otpRequest.CodeChecksCount++
	otpRequestKey := fmt.Sprintf(otpRequestKeyFormat, otpRequest.Action, otpRequest.Initiator)
	err := o.Redis.Set(ctx, otpRequestKey, otpRequest, o.Cfg.OtpRequestTTL)
	if err != nil {
		return fmt.Errorf("ошибка сохранения значения в redis: %v", err)
	}
	return nil
}
