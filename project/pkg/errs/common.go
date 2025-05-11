package errs

var (
	ResponseBuildingError = NewServiceError("ResponseBuildingError", TypeInternalError, "Ошибка построения ответа с ошибкой")
	IncorrectBodyError    = NewServiceError("IncorrectBody", TypeUserRelatedError, "Ошибка парсинга тела запроса")
	ValidationError       = NewServiceError("ValidationError", TypeUserRelatedError, "Ошибка валидации тела запроса")
)
