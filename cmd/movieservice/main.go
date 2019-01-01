package main

import (
	"database/sql"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	_ "github.com/lib/pq"
	"log"
)

func main() {

	//Получаем параметры из переменных окружения
	params := getParams()

	//Устанавливаем соединение с БД, паникуем если что-то пошло не так
	db, err := sql.Open("postgres", params.dbSource)
	if err != nil {
		log.Fatal(err)
	}

	e := echo.New()
	//валидатор используется (несколько избыточно) для проверки запросов на создание пользователя
	e.Validator = Validator()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	addAllHandlers(e, db, params)

	e.Logger.Fatal(e.Start(params.serverAddr))
}
