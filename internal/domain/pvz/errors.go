package pvz

// ErrInvalidCity ошибка при неверном городе.
type ErrInvalidCity struct{}

func (e ErrInvalidCity) Error() string {
	return "город может быть только Москва, Санкт-Петербург или Казань"
}

// ErrPVZNotFound ошибка когда ПВЗ не найден.
type ErrPVZNotFound struct{}

func (e ErrPVZNotFound) Error() string {
	return "ПВЗ не найден"
}

// ErrInvalidPaginationParams ошибка при неверных параметрах пагинации.
type ErrInvalidPaginationParams struct{}

func (e ErrInvalidPaginationParams) Error() string {
	return "неверные параметры пагинации"
}

// ErrCityEmpty ошибка при пустом городе.
type ErrCityEmpty struct{}

func (e ErrCityEmpty) Error() string {
	return "город не может быть пустым"
}

// ValidationError ошибка валидации ПВЗ.
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}
