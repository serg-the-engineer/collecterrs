package grpcx

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"your-company.com/project/pkg/errs"
)

// SentryInterceptor is a gRPC UnaryServerInterceptor for capturing errors and additional context into Sentry.
func SentryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// ...
	return nil, nil
}

// ProjectErrorInterceptor Интерцептор для обработки ошибок GRPC возвращением ошибки в формате grpc.Status.
func ProjectErrorInterceptor(
	ctx context.Context,
	req interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	resp, err := handler(ctx, req)
	if err != nil {
		var projectErr errs.ServiceError
		// Проверяем, является ли ошибка нашей кастомной ошибкой errs.ServiceError
		if errors.As(err, &projectErr) {
			// Если да, то преобразуем ее в gRPC статус
			return resp, projectErr.GRPCStatus().Err()
		}
		// Если это другая ошибка, возвращаем ее как есть
		// gRPC автоматически преобразует ее в статус codes.Unknown,
		// либо можно здесь добавить свою логику для других типов ошибок.
		return resp, err
	}
	return resp, nil

}
