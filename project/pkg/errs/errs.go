// Package errs
// для работы с ошибками, классификация и перехват причин ошибок.
//
// При классификации ошибок, связанных с логикой крайне рекомендуется упаковывать ошибки с помощью функций пакета.
// Например:
//
//	err := someSearch()
//	if err != nil {
//		return errs.Wrapf(errs.ErrNotFound, err)
//	}
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

var _ error = (*Error)(nil)

// Type тип (класс) сбоя.
type Type string

const (
	TypeInternalError    Type = "INTERNAL_ERROR"
	TypeUserRelatedError Type = "USER_RELATED_ERROR"
)

// Error служит для классификации возникающих ошибок. Поддерживает интерфейс error. В месте возникновения нужно создать
// структуру ошибки с нужным кодом и типом. Далее эту ошибку можно перехватить с помощью стандартной библиотеки работы
// с ошибками: errors.Is() и errors.As().
//
// Например:
//
//	err := errs.Error{Code: "MARK"} // такую ошибку можно вернуть из функции.
//
//	if errors.Is(err, errs.Error{Code: "MARK"}) {
//		// обрабатываем эту ошибку ...
//	}
//
// Для создания необходимой ошибки, крайне рекомендуется пользоваться функциями хелперами или статическими ошибками
// пакета errs или определенными в приложении.
type Error struct {
	Code        string
	Description string
	Type        Type
	Details     map[string]string
}

func (r Error) Error() string {
	if r.Code == "" {
		return string(r.Type)
	}

	return r.Code
}

func (r Error) WithDetails(details map[string]string) Error {
	r.Details = details

	return r
}

// Is проверяет, что аргумент типа error является эквивалентной ошибкой.
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
func (r Error) Is(err error) bool {
	var target Error
	if errors.As(err, &target) {
		return r.Equals(target)
	}

	var targetPtr *Error
	if errors.As(err, &targetPtr) {
		return targetPtr != nil && r.Equals(*targetPtr)
	}

	return false
}

// Equals проверяет, что описания ошибок равнозначны.
func (r Error) Equals(v Error) bool {
	if r.Code != "" {
		return r.Code == v.Code
	}

	return r.Type == v.Type
}

// ServiceError создает Error с указанным кодом и типом.
func ServiceError(code string, t Type, description string) Error {
	var err Error
	err.Code = code
	err.Type = t
	err.Description = description

	return err
}

// GRPCStatus() создает представление Error для передачи по GRPC
func (r Error) GRPCStatus() *status.Status {
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

	// Если есть детали, добавляем их как дополнительную информацию
	if len(r.Details) > 0 {
		// Конвертируем детали в proto-совместимый формат
		detailsProto := &structpb.Struct{
			Fields: make(map[string]*structpb.Value),
		}
		for k, v := range r.Details {
			detailsProto.Fields[k] = structpb.NewStringValue(v)
		}

		st, _ = st.WithDetails(detailsProto)
	}

	return st
}
