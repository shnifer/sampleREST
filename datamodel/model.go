//Пакет datamodel содержит типы данных, используемых в API
//В случае боевого применения, они вероятно могли бы использоваться различными сервисами
//Поэтому вынесены в отдельный пакет
package datamodel

import (
	"encoding/json"
	"fmt"
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
	Genre string `json:"genre"`
}

//Параметры запроса списка фильмов.
//Содержат необязательные фильтры и данные пагинации
type GetMoviesFilter struct {
	Pagination
	//Строка жанров передаётся в формате json.
	//Предполагается, что после формирования GetMoviesFilter она не изменяется
	GenresStr string `query:"genres"`
	//Фильтры года используются включительно
	MinYear int `query:"min_year"`
	MaxYear int `query:"max_year"`

	//единократно заполняется и возвращается методом Genres()
	//содержит список жанров в формате SQL, например:
	//'comedy','horror'
	genres string
}

//Данные пагинации, лимит и смещение демонстрируемых записей.
type Pagination struct {
	PageLimit  int `query:"page_limit"`
	PageOffset int `query:"page_offset"`
}

//Идемпотентно разбирает параметр GenresStr (genres запроса)
//в строку формата sql, т.е.
//["horror","comedy"] в 'horror','comedy'
func (p *GetMoviesFilter) Genres() (string, error) {
	if p.GenresStr == "" {
		return "", nil
	}
	if p.genres != "" {
		return p.genres, nil
	}

	var genres []string
	if err := json.Unmarshal([]byte(p.GenresStr), &genres); err != nil {
		return "", err
	}
	for i := range genres {
		genres[i] = "'" + genres[i] + "'"
	}
	p.genres = strings.Join(genres, ",")
	return p.genres, nil
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
