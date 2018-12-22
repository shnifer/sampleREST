/*Тестовое задание. REST API

Параметры сервиса задаются через переменные окружения.

movieAPIServerAddr - порт http сервера. По умолчанию :80

movieAPIDBSource - адрес базы postgres. По умолчанию:
"user=postgres password=mypass dbname=movieAPI sslmode=disable"

movieAPITokenSecret - ключ JWT-токенов. По умолчанию "secret"


============================

POST /user

Добавление нового пользователя

параметры передаются multipart/form-data:
 - login, логин, обязательно
 - password, пароль, обязательно
 - name, читаемое имя, опционально
 - age, возраст, целое число, больше 0, меньше 150, опционально
 - contact_tel, телефон, строка, опционально

возвращает:
 * 201 Created при успехе
 * 400 Bad Request для некорректных данных
 * 409 Conflict для неуникального login
 * 500 Internal Server Error при прочих ошибках SQL


============================

Авторизация пользователя (бессрочная)

POST /login

параметры передаются multipart/form-data:
 - login, логин, опционально
 - password, пароль, опционально

возвращает:
 * 200 OK при успехе
   Body: application/json; charset=UTF-8
   {
   "token": "<token>"
   }

 * 400 Bad Request для незаполненных параметров
 * 401 Unauthorized для неверной пары логин-пароль
 * 500 Internal Server Error при прочих ошибках SQL


============================

Получение списка фильмов

GET /movies

параметры передаются url query:
 Фильтры:
 - genres, фильтр множеством жанров, в json-формате []string.
   Не обязательный, если не указан отбираются все жанры. например:
   genres=["horror", "comedy"]

 - min_year, фильтр года выпуска, включительно.
 - max_year, Не обязательные, можно использовать один или оба.

 Пагинация:
 - page_limit, предел количества записей на страницу.
               Не обязательный, если не указан -- все.

 - page_offset, индекс первой записи на странице.
                Не обязательный, если не указан -- начиная с нулевой.

возвращает:
 * 200 ОК, массив фильмов
   Header:
   Total-Count: <общее количество записей по фильтру>
   Body: application/json; charset=UTF-8
 [
    {
        "id": 1,
        "title": "Home alone",
        "year": 1990,
        "genre": "comedy"
    },
    {
        "id": 2,
        "title": "Mask",
        "year": 1993,
        "genre": "comedy"
    }
]

 * 400 Bad Request для неправильно заполненых параметров
 * 500 Internal Server Error при ошибке SQL запроса


============================

Получения списка подписок пользователя

GET /rents

требует токена авторизации в
Request header:
  Authorization: Bearer <token>

 параметры передаются url query:
 Пагинация:
 - page_limit, предел количества записей на страницу.
               Не обязательный, если не указан -- все.

 - page_offset, индекс первой записи на странице.
                Не обязательный, если не указан -- начиная с нулевой.

возвращает:
 * 200 OK, массив фильмов
   Header:
   Total-Count: <общее количество записей>
   Body: application/json; charset=UTF-8
 [
    {
        "id": 1,
        "title": "Home alone",
        "year": 1990,
        "genre": "comedy"
    },
    {
        "id": 2,
        "title": "Mask",
        "year": 1993,
        "genre": "comedy"
    }
]

 * 400 Bad request при отсутствии токена или некорректных данных пагинации
 * 500 Internal Server Error при ошибке SQL запроса


============================

Добавление фильма id в подписки пользователя

PUT /rents/{id}

требует токена авторизации в
Request header:
  Authorization: Bearer <token>

возвращает:
 * 200 OK при успешном добавлении
 * 400 Bad request при отсутсвии токена или некорректном id
 * 409 Conflict В случае невозможности добавить подписка
       (login, id не существуют в соответствующих таблицах,
       или подписка уже оформлена)
 * 500 Internal Server Error при иных ошибках ошибке SQL запроса


============================

Удаление фильма id из подписки пользователя

DELETE /rents/{id}

требует токена авторизации в
Request header:
  Authorization: Bearer <token>

возвращает:
 * 200 OK при успешном добавлении
 * 400 Bad request при отсутсвии токена или некорректном id
 * 404 Not Found В случае невозможности удалить подписку
       (подписки не существует)
 * 500 Internal Server Error при иных ошибках ошибке SQL запроса


============================

Получение списка жанров

GET /genres

возвращает:
 *200 OK, список жанров
   Body: application/json; charset=UTF-8
   {
    "action",
    "comedy",
    "drama",
    "horror"
   }

 * 500 Internal Server Error при ошибках ошибке SQL запроса
*/
package main
