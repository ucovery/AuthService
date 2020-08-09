package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {

	addr, err := determineListenAddress()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", HomePage)

	/*port := GetPort()*/
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

	log.Printf("Listening on %s...\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}

/*func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
		log.Println("[-] No PORT environment variable detected. Setting to ", port)
	}
	return ":" + port
}*/

func determineListenAddress() (string, error) {
	port := os.Getenv("PORT")
	if port == "" {
		return "", fmt.Errorf("$PORT not set")
	}
	return ":" + port, nil
}
