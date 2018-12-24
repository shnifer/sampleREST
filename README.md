# sampleREST
test work

## To install DB

1) You need to have Postgres Server installed from https://www.postgresql.org

2) Download DB backup from https://github.com/shnifer/sampleREST/raw/master/db/movie.dump

3) restore DB with postgres/bin:

```
createdb --username=postgres movieAPI
pg_restore --username=postgres --dbname=movieAPI "<downloadPath>movie.dump"
(you will need to enter your password twice)
```

## To build the service

```
go get github.com/shnifer/sampleREST/...
go build github.com/shnifer/sampleREST/cmd/movieserice
```

## run 

```
SET movieAPIDBSource="<postgres data source>"
  (e.g. default "user=postgres password=postgres dbname=movieAPI sslmode=disable")
SET movieAPITokenSecret="yourTokenKey"
movieserice
```
