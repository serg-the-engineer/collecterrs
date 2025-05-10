package tests

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
	"your-company.com/project/pkg/core/memorystore"
	"your-company.com/project/pkg/otp"
	"your-company.com/project/pkg/otp/mockotp"
)

var (
	validReq = otp.Request{
		Initiator:       "some initiator",
		Action:          "sign-in",
		Payload:         []byte("test"),
		LastAttemptID:   attemptID,
		CodeValidUntil:  time.Now(),
		AttemptsCount:   0,
		Code:            "test_code",
		CodeChecksCount: 0,
		NewAttemptUntil: passDate,
	}

	longValidDate = time.Date(2177, 12, 31, 23, 59, 0, 0, time.UTC)
	passDate      = time.Date(2024, 10, 15, 12, 0, 0, 0, time.UTC)
	attemptID     = "some_attempt_id"
	otpReqKey     = "some_otp_req_key"
)

func initTestOtp(r otp.Redis) otp.ProviderOtp {
	cfg := otp.Config{
		Env:             "dev",
		OtpRequestTTL:   5 * time.Minute,
		CodeTTL:         5 * time.Minute,
		NewAttemptDelay: 30 * time.Second,
		MaxAttempts:     2,
		MaxCodeChecks:   4,
	}
	codeGenerator := otp.NewCodeGenerator(cfg.Env)
	return &otp.ProviderOtpImpl{
		Cfg:           cfg,
		Redis:         r,
		CodeGenerator: codeGenerator,
	}
}

func TestProviderOtpImpl_CreateNewAttempt(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRedis := mockotp.NewMockRedis(ctrl)
	provider := initTestOtp(mockRedis)

	type args struct {
		ctx        context.Context
		otpRequest *otp.Request
	}
	tests := []struct {
		name        string
		args        args
		want        *otp.Request
		wantErr     bool
		pretestFunc func()
	}{
		{
			name: "01 - positive case",
			args: args{
				ctx:        context.Background(),
				otpRequest: &validReq,
			},
			pretestFunc: func() {
				mockRedis.EXPECT().Set(gomock.Any(), "otp-request-sign-in-some initiator",
					gomock.Any(), gomock.Any()).Return(nil)

				mockRedis.EXPECT().Set(gomock.Any(), gomock.Any(),
					gomock.Any(), gomock.Any()).Return(nil)

				mockRedis.EXPECT().Delete(gomock.Any(), []string{"otp-attempt-some_attempt_id-request-key"}).Return(1, nil)
			},
			want: &otp.Request{
				Initiator:       validReq.Initiator,
				Action:          validReq.Action,
				Payload:         validReq.Payload,
				LastAttemptID:   "attemptID",
				AttemptsCount:   1,
				Code:            "",
				CodeChecksCount: 0,
			},
			wantErr: false,
		},
		{
			name: "02 - negative case - max attempts exceeded",
			args: args{
				ctx: context.Background(),
				otpRequest: &otp.Request{
					Initiator:       validReq.Initiator,
					Action:          validReq.Action,
					Payload:         validReq.Payload,
					ValidUntil:      validReq.ValidUntil,
					LastAttemptID:   validReq.LastAttemptID,
					AttemptsCount:   4,
					CodeValidUntil:  validReq.CodeValidUntil,
					Code:            validReq.Code,
					CodeChecksCount: 0,
				},
			},
			want:    nil,
			wantErr: true,
			pretestFunc: func() {
				return
			},
		},
		{
			name: "03 - negative case - new attempt time not exceeded",
			args: args{
				ctx: context.Background(),
				otpRequest: &otp.Request{
					Initiator:       validReq.Initiator,
					Action:          validReq.Action,
					Payload:         validReq.Payload,
					ValidUntil:      longValidDate,
					LastAttemptID:   validReq.LastAttemptID,
					AttemptsCount:   0,
					CodeValidUntil:  validReq.CodeValidUntil,
					Code:            validReq.Code,
					CodeChecksCount: 0,
					NewAttemptUntil: longValidDate,
				},
			},
			want:    nil,
			wantErr: true,
			pretestFunc: func() {
				return
			},
		},
		{
			name: "04 - negative case - OtpRequest save err",
			args: args{
				ctx:        context.Background(),
				otpRequest: &validReq,
			},
			pretestFunc: func() {
				mockRedis.EXPECT().Set(gomock.Any(), "otp-request-sign-in-some initiator",
					gomock.Any(), gomock.Any()).Return(errors.New("some error"))
			},
			want: &otp.Request{
				Initiator:       validReq.Initiator,
				Action:          validReq.Action,
				Payload:         validReq.Payload,
				LastAttemptID:   "attemptID",
				AttemptsCount:   1,
				Code:            "",
				CodeChecksCount: 0,
			},
			wantErr: true,
		},
		{
			name: "05 - negative case - OtpAttempt save err",
			args: args{
				ctx:        context.Background(),
				otpRequest: &validReq,
			},
			pretestFunc: func() {
				mockRedis.EXPECT().Set(gomock.Any(), "otp-request-sign-in-some initiator",
					gomock.Any(), gomock.Any()).Return(nil)

				mockRedis.EXPECT().Set(gomock.Any(), gomock.Any(),
					"otp-request-sign-in-some initiator", gomock.Any()).
					Return(errors.New("some OtpAttempt save err"))
			},
			want: &otp.Request{
				Initiator:       validReq.Initiator,
				Action:          validReq.Action,
				Payload:         validReq.Payload,
				LastAttemptID:   "attemptID",
				AttemptsCount:   1,
				Code:            "",
				CodeChecksCount: 0,
			},
			wantErr: true,
		},
		{
			name: "06 - negative case - OtpAttempt delete err",
			args: args{
				ctx:        context.Background(),
				otpRequest: &validReq,
			},
			pretestFunc: func() {
				mockRedis.EXPECT().Set(gomock.Any(), "otp-request-sign-in-some initiator",
					gomock.Any(), gomock.Any()).Return(nil)

				mockRedis.EXPECT().Set(gomock.Any(), gomock.Any(),
					"otp-request-sign-in-some initiator", gomock.Any()).Return(nil)

				mockRedis.EXPECT().Delete(gomock.Any(), []string{"otp-attempt-some_attempt_id-request-key"}).
					Return(0, errors.New("OtpAttempt delete err"))
			},
			want: &otp.Request{
				Initiator:       validReq.Initiator,
				Action:          validReq.Action,
				Payload:         validReq.Payload,
				LastAttemptID:   "attemptID",
				AttemptsCount:   1,
				Code:            "",
				CodeChecksCount: 0,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pretestFunc != nil {
				tt.pretestFunc()
			}
			got, err := provider.CreateNewAttempt(tt.args.ctx, tt.args.otpRequest)
			if tt.wantErr {
				assert.Error(t, err)
				ctrl.Finish()
				return
			}
			assert.Nil(t, err)
			assert.Equalf(t, tt.want.AttemptsCount, got.AttemptsCount, "CreateNewAttempt() = %v, want %v", got.AttemptsCount, tt.want.AttemptsCount)
			assert.NotEqualf(t, tt.want.LastAttemptID, got.LastAttemptID, "CreateNewAttempt() should not have equal LastAttemptID with req")
			assert.Equalf(t, tt.want.CodeChecksCount, got.CodeChecksCount, "CreateNewAttempt() = %v, want %v", got.CodeChecksCount, tt.want.CodeChecksCount)
			ctrl.Finish()
		})
	}
}

func TestProviderOtpImpl_CreateNewOtp(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRedis := mockotp.NewMockRedis(ctrl)
	provider := initTestOtp(mockRedis)

	type args struct {
		ctx       context.Context
		initiator string
		action    string
		payload   []byte
	}
	tests := []struct {
		name        string
		args        args
		want        *otp.Request
		wantErr     bool
		pretestFunc func()
	}{
		{
			name: "01 - positive case",
			args: args{
				ctx:       context.Background(),
				initiator: validReq.Initiator,
				action:    validReq.Action,
				payload:   validReq.Payload,
			},
			want: &otp.Request{
				Action:          validReq.Action,
				Payload:         validReq.Payload,
				Initiator:       validReq.Initiator,
				AttemptsCount:   1,
				CodeChecksCount: 0,
			},
			wantErr: false,
			pretestFunc: func() {
				mockRedis.EXPECT().Set(gomock.Any(), "otp-request-sign-in-some initiator",
					gomock.Any(), gomock.Any()).Return(nil)

				mockRedis.EXPECT().Set(gomock.Any(), gomock.Any(),
					"otp-request-sign-in-some initiator", gomock.Any()).Return(nil)
			},
		},
		{
			name: "02 - negative case - OtpRequest save err",
			args: args{
				ctx:       context.Background(),
				initiator: validReq.Initiator,
				action:    validReq.Action,
				payload:   validReq.Payload,
			},
			want:    nil,
			wantErr: true,
			pretestFunc: func() {
				mockRedis.EXPECT().Set(gomock.Any(), "otp-request-sign-in-some initiator",
					gomock.Any(), gomock.Any()).Return(errors.New("some error"))
			},
		},
		{
			name: "03 - negative case - OtpAttempt save err",
			args: args{
				ctx:       context.Background(),
				initiator: validReq.Initiator,
				action:    validReq.Action,
				payload:   validReq.Payload,
			},
			want:    nil,
			wantErr: true,
			pretestFunc: func() {
				mockRedis.EXPECT().Set(gomock.Any(), "otp-request-sign-in-some initiator",
					gomock.Any(), gomock.Any()).Return(nil)

				mockRedis.EXPECT().Set(gomock.Any(), gomock.Any(),
					"otp-request-sign-in-some initiator", gomock.Any()).
					Return(errors.New("some OtpAttempt save err"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pretestFunc != nil {
				tt.pretestFunc()
			}
			got, err := provider.CreateNewOtp(tt.args.ctx, tt.args.initiator, tt.args.action, tt.args.payload)
			if tt.wantErr {
				assert.Error(t, err)
				ctrl.Finish()
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, tt.want.Initiator, got.Initiator)
			assert.Equal(t, tt.want.Action, got.Action)
			assert.Equal(t, tt.want.Payload, got.Payload)
			assert.Equal(t, tt.want.CodeChecksCount, validReq.CodeChecksCount)
			assert.Equal(t, tt.want.AttemptsCount, tt.want.AttemptsCount)
			assert.NotEqual(t, tt.want.LastAttemptID, validReq.LastAttemptID)
			ctrl.Finish()
		})
	}
}

func TestProviderOtpImpl_GetOtpRequestByAction(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRedis := mockotp.NewMockRedis(ctrl)
	provider := initTestOtp(mockRedis)

	exampleReq := &otp.Request{
		Initiator:       validReq.Initiator,
		Action:          validReq.Action,
		Payload:         validReq.Payload,
		ValidUntil:      longValidDate,
		LastAttemptID:   validReq.LastAttemptID,
		AttemptsCount:   validReq.AttemptsCount,
		CodeValidUntil:  validReq.CodeValidUntil,
		Code:            validReq.Code,
		CodeChecksCount: validReq.CodeChecksCount,
	}

	memStoreBytes, err := json.Marshal(exampleReq)
	require.NoError(t, err)
	redisResp := memorystore.NewValue(memStoreBytes)

	memStoreErrCaseBytes, err := json.Marshal(validReq)
	redisErrCaseResp := memorystore.NewValue(memStoreErrCaseBytes)

	type args struct {
		ctx       context.Context
		initiator string
		action    string
	}
	tests := []struct {
		name        string
		args        args
		want        *otp.Request
		wantErr     bool
		pretestFunc func()
	}{
		{
			name: "01 - positive case",
			args: args{
				ctx:       context.Background(),
				initiator: validReq.Initiator,
				action:    validReq.Action,
			},
			want:    exampleReq,
			wantErr: false,
			pretestFunc: func() {
				mockRedis.EXPECT().Get(gomock.Any(), "otp-request-sign-in-some initiator").Return(
					redisResp, nil)
			},
		},
		{
			name: "02 - negative case - redis nil resp",
			args: args{
				ctx:       context.Background(),
				initiator: validReq.Initiator,
				action:    validReq.Action,
			},
			want:    nil,
			wantErr: false,
			pretestFunc: func() {
				mockRedis.EXPECT().Get(gomock.Any(), "otp-request-sign-in-some initiator").Return(
					nil, nil)
			},
		},
		{
			name: "03 - negative case - otpReq validUntil exceeded",
			args: args{
				ctx:       context.Background(),
				initiator: validReq.Initiator,
				action:    validReq.Action,
			},
			want:    nil,
			wantErr: false,
			pretestFunc: func() {
				mockRedis.EXPECT().Get(gomock.Any(), "otp-request-sign-in-some initiator").Return(
					redisErrCaseResp, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pretestFunc != nil {
				tt.pretestFunc()
			}

			got, err := provider.GetOtpRequestByAction(tt.args.ctx, tt.args.initiator, tt.args.action)
			if tt.wantErr {
				assert.Error(t, err)
				ctrl.Finish()
				return
			}

			assert.Nil(t, err)
			if got == nil {
				assert.Equal(t, got, tt.want)
				ctrl.Finish()
				return
			}

			assert.Equal(t, tt.want.Initiator, got.Initiator)
			assert.Equal(t, tt.want.Action, got.Action)
			ctrl.Finish()
		})
	}
}

func TestProviderOtpImpl_GetOtpRequestByAttemptID(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRedis := mockotp.NewMockRedis(ctrl)
	provider := initTestOtp(mockRedis)

	exampleReq := &otp.Request{
		Initiator:       validReq.Initiator,
		Action:          validReq.Action,
		Payload:         validReq.Payload,
		ValidUntil:      longValidDate,
		LastAttemptID:   validReq.LastAttemptID,
		AttemptsCount:   validReq.AttemptsCount,
		CodeValidUntil:  validReq.CodeValidUntil,
		Code:            validReq.Code,
		CodeChecksCount: validReq.CodeChecksCount,
	}

	memStoreBytes, err := json.Marshal(exampleReq)
	require.NoError(t, err)
	redisOtpResp := memorystore.NewValue(memStoreBytes)

	redisOtpKeyBytes, err := json.Marshal(map[string]string{"key": otpReqKey})
	redisOtpKeyResp := memorystore.NewValue(redisOtpKeyBytes)

	memStoreErrCaseBytes, err := json.Marshal(validReq)
	redisErrCaseResp := memorystore.NewValue(memStoreErrCaseBytes)

	type args struct {
		ctx       context.Context
		attemptID string
	}
	tests := []struct {
		name        string
		args        args
		want        *otp.Request
		pretestFunc func()
		wantErr     bool
	}{
		{
			name: "01 - positive case",
			args: args{
				ctx:       context.Background(),
				attemptID: validReq.LastAttemptID,
			},
			want:    exampleReq,
			wantErr: false,
			pretestFunc: func() {
				mockRedis.EXPECT().Get(gomock.Any(), gomock.Any()).
					Return(redisOtpKeyResp, nil)

				mockRedis.EXPECT().Get(gomock.Any(), gomock.Any()).
					Return(redisOtpResp, nil)
			},
		},
		{
			name: "02 - negative case - otpAttemptKey not found",
			args: args{
				ctx:       context.Background(),
				attemptID: validReq.LastAttemptID,
			},
			want: nil,
			pretestFunc: func() {
				mockRedis.EXPECT().Get(gomock.Any(), gomock.Any()).
					Return(nil, memorystore.ErrKeyNotFound)
			},
			wantErr: true,
		},
		{
			name: "03 - negative case - otpRequest ValidUntil exceeded",
			args: args{
				ctx:       context.Background(),
				attemptID: validReq.LastAttemptID,
			},
			want: nil,
			pretestFunc: func() {
				mockRedis.EXPECT().Get(gomock.Any(), gomock.Any()).
					Return(redisOtpKeyResp, nil)

				mockRedis.EXPECT().Get(gomock.Any(), gomock.Any()).
					Return(redisErrCaseResp, nil)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pretestFunc != nil {
				tt.pretestFunc()
			}
			got, err := provider.GetOtpRequestByAttemptID(tt.args.ctx, tt.args.attemptID)
			if tt.wantErr {
				assert.Error(t, err)
				ctrl.Finish()
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, tt.want.Initiator, got.Initiator)
			assert.Equal(t, tt.want.Action, got.Action)
			assert.Equal(t, tt.want.Payload, got.Payload)
			assert.Equal(t, tt.want.CodeChecksCount, validReq.CodeChecksCount)
			assert.Equal(t, tt.want.AttemptsCount, tt.want.AttemptsCount)
			ctrl.Finish()
		})
	}
}

func TestProviderOtpImpl_ValidateCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockRedis := mockotp.NewMockRedis(ctrl)
	provider := initTestOtp(mockRedis)

	type args struct {
		ctx        context.Context
		otpRequest *otp.Request
		code       string
	}
	tests := []struct {
		name        string
		args        args
		pretestFunc func()
		want        bool
		wantErr     bool
	}{
		{
			name: "01 - positive case",
			args: args{
				ctx: context.Background(),
				otpRequest: &otp.Request{
					Initiator:       validReq.Initiator,
					Action:          validReq.Action,
					Payload:         validReq.Payload,
					ValidUntil:      longValidDate,
					LastAttemptID:   validReq.LastAttemptID,
					AttemptsCount:   0,
					CodeValidUntil:  longValidDate,
					Code:            validReq.Code,
					CodeChecksCount: 0,
				},
				code: validReq.Code,
			},
			pretestFunc: func() {
				mockRedis.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(1, nil).AnyTimes()
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "02 - negative case - CodeValidUntil time exceeded",
			args: args{
				ctx: context.Background(),
				otpRequest: &otp.Request{
					CodeValidUntil: time.Now().Add(-1 * time.Hour),
				},
				code: validReq.Code,
			},
			pretestFunc: func() {
				return
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "03 - negative case - CodeChecksCount exceeded",
			args: args{
				ctx: context.Background(),
				otpRequest: &otp.Request{
					Initiator:       validReq.Initiator,
					Action:          validReq.Action,
					Payload:         validReq.Payload,
					ValidUntil:      longValidDate,
					LastAttemptID:   validReq.LastAttemptID,
					AttemptsCount:   0,
					CodeValidUntil:  longValidDate,
					Code:            validReq.Code,
					CodeChecksCount: 4,
				},
				code: validReq.Code,
			},
			pretestFunc: func() {
				return
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "04 - negative case - incrementCodeChecks on invalid code",
			args: args{
				ctx: context.Background(),
				otpRequest: &otp.Request{
					Initiator:       validReq.Initiator,
					Action:          validReq.Action,
					Payload:         validReq.Payload,
					ValidUntil:      longValidDate,
					LastAttemptID:   validReq.LastAttemptID,
					AttemptsCount:   0,
					CodeValidUntil:  longValidDate,
					Code:            validReq.Code,
					CodeChecksCount: 0,
				},
				code: "another_one_code",
			},
			pretestFunc: func() {
				mockRedis.EXPECT().Set(gomock.Any(), "otp-request-sign-in-some initiator", gomock.Any(), gomock.Any()).Return(nil)
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "05- negative case - incrementCodeChecks err",
			args: args{
				ctx: context.Background(),
				otpRequest: &otp.Request{
					Initiator:       validReq.Initiator,
					Action:          validReq.Action,
					Payload:         validReq.Payload,
					ValidUntil:      longValidDate,
					LastAttemptID:   validReq.LastAttemptID,
					AttemptsCount:   0,
					CodeValidUntil:  longValidDate,
					Code:            validReq.Code,
					CodeChecksCount: 0,
				},
				code: "another_one_code",
			},
			pretestFunc: func() {
				mockRedis.EXPECT().Set(gomock.Any(), "otp-request-sign-in-some initiator", gomock.Any(), gomock.Any()).Return(errors.New("some error"))
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pretestFunc != nil {
				tt.pretestFunc()
			}
			got, err := provider.ValidateCode(tt.args.ctx, tt.args.otpRequest, tt.args.code)
			if tt.wantErr {
				assert.Error(t, err)
				ctrl.Finish()
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, tt.want, got)
			ctrl.Finish()
		})
	}
}
