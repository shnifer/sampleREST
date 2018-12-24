//Пакет datamodel содержит типы данных, используемых в API
//В случае боевого применения, они вероятно могли бы использоваться различными сервисами
//Поэтому вынесены в отдельный пакет
package datamodel

import (
	"fmt"
	"strconv"
	"strings"
)

//Описание пользователя. Соответствует записи db.users
type User struct {
	Login      string `form:"login" validate:"required"`
	Password   string `form:"password" validate:"required"`
	Name       string `form:"name"`
	Age        byte   `form:"age" validate:"omitempty,gt=0,lt=150"`
	ContactTel string `form:"contact_tel"`
}

//Описание фильми. Соответствует записи db.movies
type Movie struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
	Year  int    `json:"year"`
	Genre int    `json:"genre_id"`
}

//Описание фильми. Соответствует записи db.movies
type Genre struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type genresString string

//Разбираем параметр списка жанров из строки через запятую в список значений
//проверяем на то, выражают ли они целые числа.
func (gs *genresString) UnmarshalParam(src string) error {
	if src == "" {
		*gs = ""
		return nil
	}
	parts := strings.Split(src, ",")
	for _, part := range parts {
		_, err := strconv.Atoi(part)
		if err != nil {
			return err
		}
	}
	*gs += genresString(src)
	return nil
}

//Параметры запроса списка фильмов.
//Содержат необязательные фильтры и данные пагинации
type GetMoviesFilter struct {
	Pagination
	//Строка жанров, передаётся как список id через запятую
	//кастомный genresString.UnmarshalParam проверяет строку при Bind'е
	Genres genresString `query:"genres"`
	//Фильтры года используются включительно
	MinYear int `query:"min_year"`
	MaxYear int `query:"max_year"`
}

//Параметры пагинации, лимит и смещение демонстрируемых записей.
//Предполагается, что пагинация опция, параметры которой должны заявляться клиентом явно.
//Можно сделать пагинацию по умолчанию, и параметр для явной её отмены.
type Pagination struct {
	PageLimit  int `query:"page_limit"`
	PageOffset int `query:"page_offset"`
}

//Возвращает часть SQL SELECT запроса по данным параметров пагинации
//OFFSET {p.PageOffset} LIMIT {p.PageLimit} для ненулевых параметров
func (p Pagination) LimitOffsetQ() (res string) {
	if p.PageOffset > 0 {
		res += fmt.Sprintf("OFFSET %v ", p.PageOffset)
	}
	if p.PageLimit > 0 {
		res += fmt.Sprintf("LIMIT %v ", p.PageLimit)
	}
	return res
}
