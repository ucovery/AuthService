package main

import (
	"log"
	"net/http"
	"os"
)

func main() {

	port := GetPort()
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
	log.Fatal(http.ListenAndServe(port, nil))
}

func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
		log.Println("[-] No PORT environment variable detected. Setting to ", port)
	}
	return ":" + port
}
