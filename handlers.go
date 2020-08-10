package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"

	"go.mongodb.org/mongo-driver/bson"
)

type UserParams struct {
	Guid string `json:"userguid"`
}

type TokenToDel struct {
	Token string `json:"token"`
}

/*Маршрут генерации access-refresh токенов
Генерируем пару a/r токенов и пишем их в клиентские куки
Паралельно пишем refresh-токен в виде bcrypt-хэша в БД */
func Signin(w http.ResponseWriter, r *http.Request) {

	var user UserParams
	//Получаем входящие параметры из json. В нашем случае это guid-пользователя
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		//Пришло не то, что ожидалось
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Bad request!")
		return
	}
	//Генерируем пару access-refresh токенов для переданного guid пользователя
	tokens, err := GetTokenPair(user.Guid)
	if err != nil {
		fmt.Fprintf(w, "Token generation error!")
		return
	}

	//Сохраняем access-токен в Cookie клиента
	SetTokenToCookie(w, "token", tokens["access_token"], false)

	//Сохраняем refresh-токен в HttpOnly cookie
	SetTokenToCookie(w, "rtoken", tokens["refresh_token"], true)

}

/*Маршрут для выполнения операции Refresh для access-refresh пары токенов
Исходим из того, что данная операция будет выполняться в случае если access-токен  не валидный/истек срок действия
Проверяем валидность acces-токена и, если он не валидный, проверяем валидность refresh-токена
Если refresh-токен валидный, сверяем его с его хэшем в monogoDB и выдаем новую пару a/r токенов*/
func Refresh(w http.ResponseWriter, r *http.Request) {
	//Проверяем валидность access-токена
	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ac_token_str := c.Value
	ac_claims := &Claims{}

	ac_token, err := jwt.ParseWithClaims(ac_token_str, ac_claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if !ac_token.Valid {
		fmt.Fprintf(w, "Token is invalid!")
		return
	}
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Забираем refresh-токен из клиентких кук и парсим
	rf_cookie, err := r.Cookie("rtoken")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rf_token_str := rf_cookie.Value
	rf_claims := &Claims{}

	rf_token, err := jwt.ParseWithClaims(rf_token_str, rf_claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	//Если Refresh-токен не валидный отправляем соотвествующий статус(или делаем редирект на форму авторизации)
	if !rf_token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Token is invalid!")
		return
	}

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Проверим используется валидная пара токенов(сгенерированы одновременно)
	if rf_claims.Token_id != ac_claims.Token_id {
		w.WriteHeader(http.StatusUnauthorized)
	}

	// Имеем на руках валидный Refresh-токен
	// Проверям acess-токен

	//fmt.Fprintf(w, "Валидация токена выполнена")

	//Подключаемся к БД
	collection, err := GetDatabaseConnection()
	if err != nil {
		log.Fatal(err)
	}

	//Задаем фильтр для запрока с БД
	//Ищем по id созданному в момент выпуска пары токенов
	filter := bson.D{{"operationid", rf_claims.Token_id}}

	var result MonogoWriteModel
	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		//Токена в базе нет, сообщаем об этом клиенту(в идеале отправляем на форму авторизации)
		w.WriteHeader(http.StatusUnauthorized)
	}
	//Сравниваем полученный от клиента Refresh-токен с тем, который лежит в базе
	//По логике, именно он был записан в БД в момент выпуска под используемым в filter id
	err = bcrypt.CompareHashAndPassword([]byte(result.RefreshToken), []byte(rf_token.Raw))

	//
	if err != nil {
		//Токен скомпроментирован, сообщаем об этом клиенту(в идеале отправляем на форму авторизации)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	//Удаляем из базы использованный Refresh-токен
	//Генерируем новую пару токенов для пользовательского guid

	filter_del := bson.D{{"operationid", result.OperationID}}

	_, err = collection.DeleteOne(context.TODO(), filter_del)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	tokens, err := GetTokenPair(result.UserGuid)
	if err != nil {
		fmt.Fprintf(w, "Token generation error!")
		return
	}
	//Сохраняем access-токен в Cookie клиента
	SetTokenToCookie(w, "token", tokens["access_token"], false)

	//Сохраняем refresh-токен в HttpOnly cookie
	SetTokenToCookie(w, "rtoken", tokens["refresh_token"], true)

	fmt.Fprint(w, "Токены успешно обновлены!")
}

//////////////////////////////////
//////////////////////////////////
//////////////////////////////////
//Маршрут для удаленния конкретного Refresh-токена из базы
func DeleteCurrentToken(w http.ResponseWriter, r *http.Request) {

	var tokendel TokenToDel
	//Получаем входящие параметры из json. В нашем случае это guid-пользователя
	err := json.NewDecoder(r.Body).Decode(&tokendel)
	if err != nil {
		//Пришло не то, что ожидалось
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Bad request!")
		return
	}

	current_token_claims := &Claims{}
	jwt.ParseWithClaims(tokendel.Token, current_token_claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	collection, err := GetDatabaseConnection()
	if err != nil {
		log.Fatal(err)
	}

	current_filter := bson.D{{"operationid", current_token_claims.Token_id}}

	_, err = collection.DeleteOne(context.TODO(), current_filter)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	fmt.Fprint(w, "Токен успешно удален!")
}

/*Маршрут для удаления всех токенов конкретного пользователя
Предполагаем, что guid пользователя получаем из POST-запроса*/
func DeleteAllUserTokens(w http.ResponseWriter, r *http.Request) {
	var user UserParams
	//Получаем входящие параметры из json. В нашем случае это guid-пользователя
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		//Пришло не то, что ожидалось
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Bad request!")
		return
	}

	collection, err := GetDatabaseConnection()
	if err != nil {
		log.Fatal(err)
	}
	//В качестве фильтра для удаления задаем guid пользователя
	//И используем операцию DeleteMany
	filter := bson.D{{"userguid", user.Guid}}

	_, err = collection.DeleteMany(context.TODO(), filter)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	fmt.Fprint(w, "Все токены пользователя удалены!")
}

func SetTokenToCookie(Writer http.ResponseWriter, cookieName string, cookieValue string, HttpOnly bool) {
	http.SetCookie(Writer, &http.Cookie{
		Name:     cookieName,
		Value:    cookieValue,
		Path:     "/",
		HttpOnly: HttpOnly,
	})
}

func HomePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Test Application:"+os.Getenv("PORT"))
}
