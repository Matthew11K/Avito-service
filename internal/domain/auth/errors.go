package auth

// ErrInvalidRole ошибка при неверной роли пользователя.
type ErrInvalidRole struct{}

func (e ErrInvalidRole) Error() string {
	return "неверная роль пользователя"
}

// ErrUserAlreadyExists ошибка при попытке создать существующего пользователя.
type ErrUserAlreadyExists struct{}

func (e ErrUserAlreadyExists) Error() string {
	return "пользователь с таким email уже существует"
}

// ErrUserNotFound ошибка когда пользователь не найден.
type ErrUserNotFound struct{}

func (e ErrUserNotFound) Error() string {
	return "пользователь не найден"
}

// ErrInvalidCredentials ошибка при неверных учетных данных.
type ErrInvalidCredentials struct{}

func (e ErrInvalidCredentials) Error() string {
	return "неверный email или пароль"
}

// ErrEmptyToken ошибка при пустом токене аутентификации.
type ErrEmptyToken struct{}

func (e ErrEmptyToken) Error() string {
	return "пустой токен"
}

// ErrEmailEmpty ошибка при пустом email.
type ErrEmailEmpty struct{}

func (e ErrEmailEmpty) Error() string {
	return "email не может быть пустым"
}

// ErrPasswordEmpty ошибка при пустом пароле.
type ErrPasswordEmpty struct{}

func (e ErrPasswordEmpty) Error() string {
	return "пароль не может быть пустым"
}

// ErrRoleEmpty ошибка при пустой роли.
type ErrRoleEmpty struct{}

func (e ErrRoleEmpty) Error() string {
	return "роль не может быть пустой"
}

// ValidationError ошибка валидации пользователя.
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}
