package movieservice

import (
	"github.com/go-playground/validator"
	"github.com/labstack/echo"
)

//используется go-playground/validator для проверки входящих данных
//реализуем обёртку в echo.Validator
//можно было бы использовать другую библиотеку,
//или проверить единственный случай вручную
type myValidator struct {
	validator *validator.Validate
}

func (v myValidator) Validate(value interface{}) error {
	return v.validator.Struct(value)
}

func Validator() echo.Validator {
	return myValidator{
		validator: validator.New(),
	}
}
