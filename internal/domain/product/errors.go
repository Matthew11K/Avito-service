package product

// ErrInvalidProductType ошибка при неверном типе товара.
type ErrInvalidProductType struct{}

func (e ErrInvalidProductType) Error() string {
	return "неверный тип товара"
}

// ErrNoProductsToDelete ошибка, когда нет товаров для удаления.
type ErrNoProductsToDelete struct{}

func (e ErrNoProductsToDelete) Error() string {
	return "нет товаров для удаления"
}

// ErrProductNotFound ошибка, когда товар не найден.
type ErrProductNotFound struct{}

func (e ErrProductNotFound) Error() string {
	return "товар не найден"
}

// ErrTypeEmpty ошибка при пустом типе товара.
type ErrTypeEmpty struct{}

func (e ErrTypeEmpty) Error() string {
	return "тип товара не может быть пустым"
}

// ValidationError ошибка валидации товара.
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}
