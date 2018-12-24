/*Пакет DB реализует функции непосредственного доступ к базе данных

=================================

CREATE TABLE users
(
    id SERIAL PRIMARY KEY,
    login VARCHAR NOT NULL,
    password VARCHAR NOT NULL,
    name VARCHAR,
    age INTEGER,
    contact_tel VARCHAR
);

CREATE TABLE genres
(
    id SERIAL PRIMARY KEY,
    name VARCHAR NOT NULL
);

CREATE TABLE movies
(
    id SERIAL PRIMARY KEY,
    title VARCHAR NOT NULL,
    year INTEGER NOT NULL,
    genre INTEGER NOT NULL REFERENCES genres
);

CREATE TABLE rents
(
    user_id INTEGER NOT NULL REFERENCES users,
    movie_id INTEGER NOT NULL REFERENCES movies,
    PRIMARY KEY (user_id, movie_id)
);

*/
package db
