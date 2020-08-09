package main

import (
	"log"
	"net/http"
)

func main() {

	//Определяем маршруты
	//Маршрут выдает пару access-refresh токенов
	http.HandleFunc("/signin", Signin)

	//Маршрут обновляет пару access-refresh токенов
	http.HandleFunc("/refresh", Refresh)

	//Маршрут удаляет конкретный токен
	http.HandleFunc("/deleteone", DeleteCurrentToken)

	//Маршрут удаляет все токены выданные конекретному guid
	http.HandleFunc("/deleteall", DeleteAllUserTokens)

	//Запускаем сервер на порте 8000
	log.Fatal(http.ListenAndServe(":8000", nil))
}
