package reception

// ErrActiveReceptionExists ошибка, когда уже есть активная приемка.
type ErrActiveReceptionExists struct{}

func (e ErrActiveReceptionExists) Error() string {
	return "уже есть незакрытая приемка"
}

// ErrNoActiveReception ошибка, когда нет активной приемки.
type ErrNoActiveReception struct{}

func (e ErrNoActiveReception) Error() string {
	return "нет активной приемки"
}

// ErrReceptionNotFound ошибка, когда приемка не найдена.
type ErrReceptionNotFound struct{}

func (e ErrReceptionNotFound) Error() string {
	return "приемка не найдена"
}

// ErrReceptionClosed ошибка, когда приемка уже закрыта.
type ErrReceptionClosed struct{}

func (e ErrReceptionClosed) Error() string {
	return "приемка уже закрыта"
}

// ValidationError ошибка валидации приемки.
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}
