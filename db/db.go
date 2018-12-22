//Пакет DB реализует функции непосредственного доступ к базе данных
package db

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/shnifer/sampleREST/datamodel"
	"strings"
)

var db *sql.DB

//Ошибка, соответствующая Postgres ошибке c кодом 23505
var ErrorUnique = errors.New("unique violation")

//Ошибка, соответствующая Postgres ошибке c кодом 23503
var ErrorForeignKey = errors.New("foreign key violation")

//Ошибка отсутствия удаляемого элемента
var ErrorDoNotExist = errors.New("do not exist")

//Open создаёт соединение с базой данных и сохраняет его в глобальной переменной пакета
func Open(driver, source string) (err error) {
	if db != nil {
		if err := db.Close(); err != nil {
			return err
		}
	}

	db, err = sql.Open(driver, source)
	return err
}

//AddUser добавляет пользователя newUser в db
func AddUser(newUser datamodel.User) error {
	//Если возраст не указан будет записано 0.
	//Можно сделать NULL, в зависимости от соглашений
	_, err := db.Exec(
		`INSERT INTO users(login,pass,name,age,contact_tel) 
VALUES ($1, $2, $3, $4, $5)
		`, newUser.Login, newUser.Password, newUser.Name, newUser.Age, newUser.ContactTel)
	return parsePGError(err)
}

//CheckUser возвращает true если в базе есть пользователь user с паролем password
func CheckUser(user, password string) bool {
	row := db.QueryRow(`SELECT FROM users WHERE login=$1 AND pass=$2`, user, password)
	err := row.Scan()
	//считаем, что если что-то пошло не так на стороне БД,
	//то это тоже ошибка авторизации.
	//Отдельно ошибки не обрабатываем
	return err == nil
}

//GetMovies возвращает список фильмов из DB по заданным в param фильтрам
func GetMovies(params datamodel.GetMoviesFilter) (res []datamodel.Movie, err error) {
	res = make([]datamodel.Movie, 0)

	rows, err := db.Query(getMoviesQuery(params))
	if err != nil {
		return nil, err
	}

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
func GetMoviesTotalCount(params datamodel.GetMoviesFilter) (count int, err error) {
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
	genres, _ := params.Genres()
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
func GetRentedMovies(login string, pagin datamodel.Pagination) (res []datamodel.Movie, err error) {
	res = make([]datamodel.Movie, 0)

	rows, err := db.Query(`SELECT movies.id, movies.title, movies.year, movies.genre 
                   FROM movies JOIN rents ON (movies.id=rents.movie_id)
                   WHERE rents.user_login=$1
                   ORDER BY movies.id
                   `+pagin.LimitOffsetQ(), login)
	if err != nil {
		return nil, err
	}

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

//GetRentedMoviesCount возвращает общее количество подписок у пользователя login
func GetRentedMoviesCount(login string) (count int, err error) {
	row := db.QueryRow("SELECT COUNT(*) FROM rents WHERE user_login=$1", login)
	err = row.Scan(&count)
	return count, err
}

//PutRent добавляет подписку пользователя login на фильм id
func PutRent(login string, id int) error {
	_, err := db.Exec(`INSERT INTO rents (user_login, movie_id) VALUES ($1,$2)`, login, id)
	return parsePGError(err)
}

//PutRent удаляет подписку пользователя login на фильм id
func DelRent(login string, id int) error {
	result, err := db.Exec(`DELETE FROM rents WHERE user_login=$1 AND movie_id=$2`, login, id)
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

func GetGenres() (res []string, err error) {
	res = make([]string, 0)

	rows, err := db.Query(`SELECT name FROM genres ORDER BY name`)
	if err != nil {
		return nil, parsePGError(err)
	}
	var name string
	for rows.Next() {
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		res = append(res, name)
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
