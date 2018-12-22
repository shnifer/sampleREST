package main

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/shnifer/sampleREST/datamodel"
	"github.com/shnifer/sampleREST/db"
	"log"
	"net/http"
	"strconv"
)

//POST /user
//Добавление нового пользователя
func postUsersHandler(ctx echo.Context) (err error) {
	var newUser datamodel.User
	if err = ctx.Bind(&newUser); err != nil {
		return err
	}
	if err = ctx.Validate(newUser); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := db.AddUser(newUser); err != nil {
		if err == db.ErrorUnique {
			echo.NewHTTPError(http.StatusConflict, err.Error())
		} else {
			echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}

	return ctx.NoContent(http.StatusCreated)
}

//Авторизация пользователя (бессрочная)
//POST /login
func postLoginHandler(ctx echo.Context) (err error) {
	userName := ctx.FormValue("login")
	password := ctx.FormValue("password")

	if userName == "" || password == "" {
		return echo.ErrBadRequest
	}
	if !db.CheckUser(userName, password) {
		return echo.ErrUnauthorized
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["login"] = userName

	t, err := token.SignedString(Params.tokenSecret)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, map[string]string{"token": t})
}

//Получение списка фильмов
//GET /movies
func getMoviesHandler(ctx echo.Context) (err error) {
	var params datamodel.GetMoviesFilter
	if err := ctx.Bind(&params); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if _, err := params.Genres(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	totalCount, err := db.GetMoviesTotalCount(params)
	if err != nil {
		return err
	}
	movies, err := db.GetMovies(params)
	if err != nil {
		return err
	}

	ctx.Response().Header().Set("Total-Count", strconv.Itoa(totalCount))
	return ctx.JSON(http.StatusOK, movies)
}

//Получения списка подписок пользователя
//GET /rents
func getRentsHandler(ctx echo.Context) (err error) {
	login, err := getAuthLogin(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var pagin datamodel.Pagination
	err = ctx.Bind(&pagin)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	totalCount, err := db.GetRentedMoviesCount(login)
	if err != nil {
		return err
	}

	movies, err := db.GetRentedMovies(login, pagin)
	if err != nil {
		return err
	}

	ctx.Response().Header().Set("Total-Count", strconv.Itoa(totalCount))
	return ctx.JSON(http.StatusOK, movies)
}

//Добавление фильма id в подписки пользователя
//PUT /rents/{id}
func putRentsHandler(ctx echo.Context) error {
	login, err := getAuthLogin(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err = db.PutRent(login, id)
	if err != nil {
		switch err {
		case db.ErrorUnique, db.ErrorForeignKey:
			log.Println(err)
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		default:
			return err
		}
	}

	return ctx.NoContent(http.StatusOK)
}

//Удаление фильма id из подписки пользователя
//DELETE /rents/{id}
func delRentsHandler(ctx echo.Context) error {
	login, err := getAuthLogin(ctx)
	if err != nil {
		return err
	}

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err = db.DelRent(login, id)
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

//Получение списка жанров
//GET /genres
func getGenresHandler(ctx echo.Context) error {
	genres, err := db.GetGenres()
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, genres)
}

//getAuthLogin извлекает логин из токена авторизации
//должна вызываться в контексте с middleware.JWT
//при ошибке -- возвращает "getAuthLoginError"
func getAuthLogin(ctx echo.Context) (string, error) {
	err := errors.New("getAuthLogin called without auth token parsed")
	t, ok := ctx.Get("user").(*jwt.Token)
	if !ok {
		return "", err
	}
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return "", err
	}
	login, ok := claims["login"].(string)
	if !ok {
		return "", err
	}
	return login, nil
}
