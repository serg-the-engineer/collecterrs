// Package errs
// для работы с ошибками, классификация и перехват причин ошибок.
//
// Для того, чтобы классифицированные ошибки могли быть перехвачены в любом слое приложения.
//
// Неупакованными можно оставлять системные ошибки, например сбой ввода, вывода, ошибка соединения и т.д...
package errs

import (
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

var _ error = (*ServiceError)(nil)

// Type тип (класс) сбоя.
type Type string

const (
	TypeInternalError    Type = "INTERNAL_ERROR"
	TypeUserRelatedError Type = "USER_RELATED_ERROR"
)

// ServiceError служит для классификации возникающих ошибок внутри бэкэнда. Поддерживает интерфейс error. В месте возникновения нужно создать
// структуру ошибки с нужным кодом и типом. Далее эту ошибку можно перехватить с помощью стандартной библиотеки работы
// с ошибками: errors.Is() и errors.As().
//
// Например:
//
//	err := errs.IncorrectBodyError  // такую ошибку можно вернуть из функции.
//
//	if errors.Is(err, errs.IncorrectBodyError) {
//		// обрабатываем эту ошибку ...
//	}
//
type ServiceError struct {
	Code        string
	Description string
	Type        Type
	Details     map[string]string
}

func (r ServiceError) Error() string {
	if r.Code == "" {
		return string(r.Type)
	}

	return r.Code
}

func (r ServiceError) WithDetails(details map[string]string) ServiceError {
	r.Details = details

	return r
}

// Is проверяет, что аргумент типа error является эквивалентной ошибкой.
// Применяется в том числе для проверки ответов из сервисов по GRPC
// Например:
//
//	err := someSearch() // может вернуть errs.ErrNotFound
//	if err != nil {
//		if errs.ErrNotFound.Is(err) {
//			// перехватываем ошибку ...
//		}
//		// какая то другая ошибка ...
//	}
//
// .
func (r ServiceError) Is(err error) bool {
	var target ServiceError
	if errors.As(err, &target) {
		return r.Equals(target)
	}

	var targetPtr *ServiceError
	if errors.As(err, &targetPtr) {
		return targetPtr != nil && r.Equals(*targetPtr)
	}

	s := BuildFromGRPCStatus(status.Convert(err))
	return r.Equals(s)
}

// Equals проверяет, что описания ошибок равнозначны.
func (r ServiceError) Equals(v ServiceError) bool {
	if r.Code != "" {
		return r.Code == v.Code
	}

	return r.Type == v.Type
}

// NewServiceError создает ServiceError с указанным кодом и типом.
func NewServiceError(code string, t Type, description string) ServiceError {
	var err ServiceError
	err.Code = code
	err.Type = t
	err.Description = description

	return err
}

// GRPCStatus() создает представление ServiceError для передачи по GRPC
func (r ServiceError) GRPCStatus() *status.Status {
	// Определяем код GRPC на основе типа ошибки
	var code codes.Code
	switch r.Type {
	case TypeUserRelatedError:
		code = codes.InvalidArgument
	case TypeInternalError:
		code = codes.Internal
	default:
		code = codes.Unknown
	}

	// Создаем статус с кодом - описание и тип при желании мы можем получать по коду из project.errs
	st := status.New(code, r.Code)

	// Конвертируем детали в proto-совместимый формат
	detailsProto := &structpb.Struct{
		Fields: make(map[string]*structpb.Value),
	}

	// Добавляем описание как отдельное поле
	if r.Description != "" {
		detailsProto.Fields["_description"] = structpb.NewStringValue(r.Description)
	}

	// Если есть детали, добавляем их как дополнительную информацию
	if len(r.Details) > 0 {
		for k, v := range r.Details {
			detailsProto.Fields[k] = structpb.NewStringValue(v)
		}
	}

	if len(detailsProto.Fields) > 0 {
		st, _ = st.WithDetails(detailsProto)
	}

	return st
}

// BuildFromGRPCStatus создает ServiceError из представления GRPC Status
func BuildFromGRPCStatus(st *status.Status) ServiceError {
	if st == nil {
		return ServiceError{}
	}

	// Определяем тип ошибки на основе кода GRPC
	var errType Type
	switch st.Code() {
	case codes.InvalidArgument:
		errType = TypeUserRelatedError
	case codes.Internal:
		errType = TypeInternalError
	default:
		errType = TypeInternalError
	}

	err := ServiceError{
		Code: st.Message(),
		Type: errType,
	}

	// Извлекаем детали из статуса, если они есть
	details := st.Details()
	if len(details) > 0 {
		for _, detail := range details {
			if s, ok := detail.(*structpb.Struct); ok {
				// Создаем map для деталей, если он еще не создан
				if err.Details == nil {
					err.Details = make(map[string]string)
				}

				// Обрабатываем все поля
				for k, v := range s.Fields {
					if k == "_description" {
						// Особая обработка для поля описания
						err.Description = v.GetStringValue()
					} else {
						// Остальные поля добавляем в детали
						err.Details[k] = v.GetStringValue()
					}
				}
			}
		}
	}

	return err
}

// Формат ошибок, отдающиеся вовне с гейтвеев
type ServerError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

func (r ServerError) Error() string {
	if r.Code == "" {
		return r.Message
	}
	return r.Code
}

func BuildFromServiceError(e ServiceError) ServerError {
	se := ServerError{
		Code:    e.Code,
		Message: e.Description,
	}
	if e.Details != nil && len(e.Details) > 0 {
		se.Details = e.Details
	}
	// для неименных ошибок сообщение пишется в Code, приводим к общему виду
	if e.Description == "" {
		se.Message = e.Code
		se.Code = "InternalServiceError"
	}
	return se
}
