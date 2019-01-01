package db

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/shnifer/sampleREST/datamodel"
	"strings"
)

//Ошибка нарушения уникальности, соответствующая Postgres ошибке c кодом 23505
var ErrorUnique = errors.New("unique violation")

//Ошибка нарушения внешнего ключа, соответствующая Postgres ошибке c кодом 23503
var ErrorForeignKey = errors.New("foreign key violation")

//Ошибка отсутствия элемента
var ErrorDoNotExist = errors.New("do not exist")

//AddUser добавляет пользователя newUser в db
func AddUser(db *sql.DB, newUser datamodel.User) error {
	//Если возраст не указан будет записано 0.
	//Можно сделать NULL, в зависимости от соглашений
	_, err := db.Exec(
		`INSERT INTO users(login,password,name,age,contact_tel) 
VALUES ($1, $2, $3, $4, $5)
		`, newUser.Login, newUser.Password, newUser.Name, newUser.Age, newUser.ContactTel)
	return parsePGError(err)
}

//CheckUser возвращает userId, если в базе есть пользователь user с паролем password
//если нет -- возвращает ошибку
func CheckUser(db *sql.DB, user, password string) (userId int, err error) {
	row := db.QueryRow(`SELECT id FROM users WHERE login=$1 AND password=$2`, user, password)
	err = row.Scan(&userId)
	return userId, err
}

//GetMovies возвращает список фильмов из DB по заданным в param фильтрам
func GetMovies(db *sql.DB, params datamodel.GetMoviesFilter) (res []datamodel.Movie, err error) {
	res = make([]datamodel.Movie, 0)

	rows, err := db.Query(getMoviesQuery(params))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movie datamodel.Movie
	for rows.Next() {
		err := rows.Scan(&movie.Id, &movie.Title, &movie.Year, &movie.Genre)
		if err != nil {
			return nil, err
		}
		res = append(res, movie)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return res, nil
}

//getMoviesQuery возвращает запрос к таблице DB.movies по заданным параметрам
func getMoviesQuery(params datamodel.GetMoviesFilter) string {
	//находим части запроса WHERE и LIMIT-OFFSET
	whereQ := getMoviesWhereQ(params)
	limitOffset := params.LimitOffsetQ()
	//и собираем всё воедино
	return "SELECT id, title, year, genre FROM movies " +
		whereQ +
		"ORDER BY ID " +
		limitOffset
}

//GetMoviesTotalCount возвращает количество записей в таблице DB.movies по заданным фильтрам
func GetMoviesTotalCount(db *sql.DB, params datamodel.GetMoviesFilter) (count int, err error) {
	whereQ := getMoviesWhereQ(params)
	row := db.QueryRow("SELECT COUNT(*) FROM movies " + whereQ)
	err = row.Scan(&count)
	return count, err
}

//getMoviesWhereQ возвращает строку "WHERE .... " запроса по указанным в param фильтрам
//или пустую строку, если фильтров нет
func getMoviesWhereQ(params datamodel.GetMoviesFilter) string {
	//массив отдельных условий
	wheres := make([]string, 0)
	add := func(filter string, args ...interface{}) {
		wheres = append(wheres, fmt.Sprintf(filter, args...))
	}

	//Предполагаем, что валидация уже проведена
	genres := params.Genres
	if genres != "" {
		add("genre in(%v)", genres)
	}
	if params.MinYear > 0 {
		add("year>=%v", params.MinYear)
	}
	if params.MaxYear > 0 {
		add("year<=%v", params.MaxYear)
	}

	//собираем в одну SQL конструкцию "WHERE ... AND ..."
	if len(wheres) == 0 {
		return ""
	} else {
		return "WHERE " + strings.Join(wheres, " AND ") + " "
	}
}

//GetRentedMovies возвращает список фильмов,
//на которые у пользователя login оформлена подписка,
//с учётом пагинации
func GetRentedMovies(db *sql.DB, userId int, pagin datamodel.Pagination) (res []datamodel.Movie, err error) {
	res = make([]datamodel.Movie, 0)

	rows, err := db.Query(`SELECT movies.id, movies.title, movies.year, movies.genre 
                   FROM movies JOIN rents ON (movies.id=rents.movie_id)
                   WHERE rents.user_id=$1
                   ORDER BY movies.id
                   `+pagin.LimitOffsetQ(), userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movie datamodel.Movie
	for rows.Next() {
		err := rows.Scan(&movie.Id, &movie.Title, &movie.Title, &movie.Genre)
		if err != nil {
			return nil, err
		}
		res = append(res, movie)
	}

	if rows.Err() != nil {
		return nil, err
	}
	return res, nil
}

//GetRentedMoviesCount возвращает общее количество подписок у пользователя c userId
func GetRentedMoviesCount(db *sql.DB, userId int) (count int, err error) {
	row := db.QueryRow("SELECT COUNT(*) FROM rents WHERE user_id=$1", userId)
	err = row.Scan(&count)
	return count, err
}

//PutRent добавляет подписку пользователя c userId на фильм movieId
func PutRent(db *sql.DB, userId int, movieId int) error {
	_, err := db.Exec(`INSERT INTO rents (user_id, movie_id) VALUES ($1,$2)`, userId, movieId)
	return parsePGError(err)
}

//DelRent удаляет подписку пользователя c userId на фильм movieId
func DelRent(db *sql.DB, userId int, movieId int) error {
	result, err := db.Exec(`DELETE FROM rents WHERE user_id=$1 AND movie_id=$2`, userId, movieId)
	if err != nil {
		return parsePGError(err)
	}
	//Проверяем сколько записей было удалено, если 0 - то это ошибочная попытка
	rowsCount, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsCount == 0 {
		return ErrorDoNotExist
	}
	return nil
}

func GetGenres(db *sql.DB) (res []datamodel.Genre, err error) {
	res = make([]datamodel.Genre, 0)

	rows, err := db.Query(`SELECT id,name FROM genres ORDER BY id`)
	if err != nil {
		return nil, parsePGError(err)
	}
	defer rows.Close()

	var genre datamodel.Genre
	for rows.Next() {
		if err := rows.Scan(&genre.Id, &genre.Name); err != nil {
			return nil, err
		}
		res = append(res, genre)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return res, nil
}

//parsePGError проверяет ошибку, и если это ошибка Postgres
//ErrorUnique или ErrorForeignKey возвращает соответствующую ошибку модуля
//иначе возвращает значение неизменным
func parsePGError(err error) error {
	if err == nil {
		return nil
	}
	pgErr, ok := err.(*pq.Error)
	if !ok {
		return err
	}
	switch pgErr.Code {
	case "23505":
		return ErrorUnique
	case "23503":
		return ErrorForeignKey
	default:
		return err
	}
}
