/*Пакет DB реализует функции непосредственного доступ к базе данных Postgres

=================================

Таблица пользователей
 CREATE TABLE users
 (
    id SERIAL PRIMARY KEY,
    login VARCHAR NOT NULL,
    password VARCHAR NOT NULL,
    name VARCHAR,
    age INTEGER,
    contact_tel VARCHAR
 );

Таблица жанров
 CREATE TABLE genres
 (
     id SERIAL PRIMARY KEY,
     name VARCHAR NOT NULL
 );

Таблица фильмов
 CREATE TABLE movies
 (
    id SERIAL PRIMARY KEY,
    title VARCHAR NOT NULL,
    year INTEGER NOT NULL,
    genre INTEGER NOT NULL REFERENCES genres
 );

Таблица подписок
 CREATE TABLE rents
 (
    user_id INTEGER NOT NULL REFERENCES users,
    movie_id INTEGER NOT NULL REFERENCES movies,
    PRIMARY KEY (user_id, movie_id)
 );

*/
package db
