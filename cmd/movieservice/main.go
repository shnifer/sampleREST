package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	_ "github.com/lib/pq"
	"github.com/shnifer/sampleREST/db"
	"log"
)

func main() {

	//Устанавливаем соединение с БД, паникуем если что-то пошло не так
	err := db.Open("postgres", Params.dbSource)
	if err != nil {
		log.Fatal(err)
	}

	e := echo.New()
	//валидатор используется (несколько избыточно) для проверки запросов на создание пользователя
	e.Validator = Validator()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	//добавление пользователя
	e.POST("/users", postUsersHandler)
	//авторизация
	e.POST("/login", postLoginHandler)
	//список фильмов
	e.GET("/movies", getMoviesHandler)

	//группа под авторизацией
	authGroup := e.Group("/rents")
	authGroup.Use(middleware.JWT(Params.tokenSecret))
	{
		//список аредованный пользователем фильмов
		authGroup.GET("", getRentsHandler)
		//добавить аренду
		authGroup.PUT("/:movieId", putRentsHandler)
		//завершить аренду
		authGroup.DELETE("/:movieId", delRentsHandler)
	}

	//не упомянутое в задание, но IMHO нужное
	//если мы не хотим потом изменять список жанров,
	//захардкоженный в 5 клиентах
	e.GET("/genres", getGenresHandler)

	e.Logger.Fatal(e.Start(Params.serverAddr))
}
