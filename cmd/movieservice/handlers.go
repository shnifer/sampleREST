package main

import (
	"database/sql"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/shnifer/sampleREST/datamodel"
	"github.com/shnifer/sampleREST/db"
	"net/http"
	"strconv"
)

func addAllHandlers(e *echo.Echo, sqlDb *sql.DB, params params) {
	//добавление пользователя
	e.POST("/users", postUsersHandler(sqlDb))
	//авторизация
	e.POST("/login", postLoginHandler(sqlDb, params))
	//список фильмов
	e.GET("/movies", getMoviesHandler(sqlDb))

	//группа под авторизацией
	authGroup := e.Group("/rents")
	authGroup.Use(middleware.JWT(params.tokenSecret))
	{
		//список аредованный пользователем фильмов
		authGroup.GET("", getRentsHandler(sqlDb))
		//добавить аренду
		authGroup.PUT("/:movieId", putRentsHandler(sqlDb))
		//завершить аренду
		authGroup.DELETE("/:movieId", delRentsHandler(sqlDb))
	}

	//не упомянутое в задание, но IMHO нужное
	//если мы не хотим потом изменять список жанров,
	//захардкоженный в 5 клиентах
	e.GET("/genres", getGenresHandler(sqlDb))
}

//POST /user
//Добавление нового пользователя
func postUsersHandler(sqlDb *sql.DB) echo.HandlerFunc {
	return func(ctx echo.Context) (err error) {
		var newUser datamodel.User
		if err = ctx.Bind(&newUser); err != nil {
			return err
		}
		if err = ctx.Validate(newUser); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		if err := db.AddUser(sqlDb, newUser); err != nil {
			if err == db.ErrorUnique {
				echo.NewHTTPError(http.StatusConflict, err.Error())
			} else {
				echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
		}

		return ctx.NoContent(http.StatusCreated)
	}
}

//Авторизация пользователя (бессрочная)
//POST /login
func postLoginHandler(sqlDb *sql.DB, params params) echo.HandlerFunc {
	return func(ctx echo.Context) (err error) {
		userName := ctx.FormValue("login")
		password := ctx.FormValue("password")

		if userName == "" || password == "" {
			return echo.ErrBadRequest
		}

		id, err := db.CheckUser(sqlDb, userName, password)
		if err != nil {
			return echo.ErrUnauthorized
		}

		token := jwt.New(jwt.SigningMethodHS256)
		claims := token.Claims.(jwt.MapClaims)
		//конвертируем в строку, т.к. иначе оно станет float64 а это не очень хорошо
		claims["userid"] = strconv.Itoa(id)

		t, err := token.SignedString(params.tokenSecret)
		if err != nil {
			return err
		}
		return ctx.JSON(http.StatusOK, map[string]string{"token": t})
	}
}

//Получение списка фильмов
//GET /movies
func getMoviesHandler(sqlDb *sql.DB) echo.HandlerFunc {
	return func(ctx echo.Context) (err error) {
		var params datamodel.GetMoviesFilter
		if err := ctx.Bind(&params); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		totalCount, err := db.GetMoviesTotalCount(sqlDb, params)
		if err != nil {
			return err
		}
		movies, err := db.GetMovies(sqlDb, params)
		if err != nil {
			return err
		}

		ctx.Response().Header().Set("Total-Count", strconv.Itoa(totalCount))
		return ctx.JSON(http.StatusOK, movies)
	}
}

//Получения списка подписок пользователя
//GET /rents
func getRentsHandler(sqlDb *sql.DB) echo.HandlerFunc {
	return func(ctx echo.Context) (err error) {
		userId, err := getAuthLogin(ctx)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		var pagin datamodel.Pagination
		err = ctx.Bind(&pagin)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		totalCount, err := db.GetRentedMoviesCount(sqlDb, userId)
		if err != nil {
			return err
		}

		movies, err := db.GetRentedMovies(sqlDb, userId, pagin)
		if err != nil {
			return err
		}

		ctx.Response().Header().Set("Total-Count", strconv.Itoa(totalCount))
		return ctx.JSON(http.StatusOK, movies)
	}
}

//Добавление фильма movie_id в подписки пользователя
//PUT /rents/{movie_id}
func putRentsHandler(sqlDb *sql.DB) echo.HandlerFunc {
	return func(ctx echo.Context) (err error) {
		userId, err := getAuthLogin(ctx)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		movieId, err := strconv.Atoi(ctx.Param("movieId"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		err = db.PutRent(sqlDb, userId, movieId)
		if err != nil {
			switch err {
			case db.ErrorUnique, db.ErrorForeignKey:
				return echo.NewHTTPError(http.StatusConflict, err.Error())
			default:
				return err
			}
		}

		return ctx.NoContent(http.StatusOK)
	}
}

//Удаление фильма movie_id из подписки пользователя
//DELETE /rents/{movie_id}
func delRentsHandler(sqlDb *sql.DB) echo.HandlerFunc {
	return func(ctx echo.Context) (err error) {
		userId, err := getAuthLogin(ctx)
		if err != nil {
			return err
		}

		movieId, err := strconv.Atoi(ctx.Param("movieId"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		err = db.DelRent(sqlDb, userId, movieId)
		if err != nil {
			switch err {
			case db.ErrorDoNotExist:
				return echo.NewHTTPError(http.StatusNotFound, err.Error())
			default:
				return err
			}
		}

		return ctx.NoContent(http.StatusOK)
	}
}

//Получение списка жанров
//GET /genres
func getGenresHandler(sqlDb *sql.DB) echo.HandlerFunc {
	return func(ctx echo.Context) (err error) {
		genres, err := db.GetGenres(sqlDb)
		if err != nil {
			return err
		}

		return ctx.JSON(http.StatusOK, genres)
	}
}

//getAuthLogin извлекает логин из токена авторизации
//должна вызываться в контексте с middleware.JWT
//при ошибке -- возвращает "getAuthLoginError"
func getAuthLogin(ctx echo.Context) (int, error) {
	err := errors.New("getAuthLogin called without auth token parsed")
	t, ok := ctx.Get("user").(*jwt.Token)
	if !ok {
		return 0, err
	}
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return 0, err
	}
	userIdStr, ok := claims["userid"].(string)
	if !ok {
		return 0, err
	}
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return 0, err
	}
	return userId, nil
}
