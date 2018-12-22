package main

import "os"

// В рамках примера используем глобальную переменную для общих настроек.
// По соглашениям проекта брали бы настройки из переменных окружений, флагов и/или файла настройки.
var Params struct {
	serverAddr  string
	tokenSecret []byte
	dbSource    string
}

//читаем общие настройки.
func init() {
	var exist bool
	if Params.serverAddr, exist = os.LookupEnv("movieAPIServerAddr"); !exist {
		Params.serverAddr = ":80"
	}
	if Params.dbSource, exist = os.LookupEnv("movieAPIDBSource"); !exist {
		Params.dbSource = "user=postgres password=mypass dbname=movieAPI sslmode=disable"
	}
	var str string
	if str, exist = os.LookupEnv("movieAPITokenSecret"); !exist {
		//значение ключа по умолчанию. Используется для простоты запуска тестового примера.
		//НЕ надо иметь никаких ключей в боевом коде.
		str = "secret"
	}
	Params.tokenSecret = []byte(str)
}
