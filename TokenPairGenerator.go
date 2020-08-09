package main

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/rs/xid"
)

var jwtKey = []byte("ClientKey")

type Claims struct {
	Token_id xid.ID
	UserGuid string
	jwt.StandardClaims
}

func GetTokenPair(guid string) (map[string]string, error) {
	//Генерируем уникальный id для связи пары токенов
	token_id := xid.New()

	//Генерируем access-токен с временем жизни 15 минут
	at_exp_time := time.Now().Add(15 * time.Minute)
	at_claims := &Claims{
		Token_id: token_id,
		UserGuid: guid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: at_exp_time.Unix(),
		},
	}
	access_token := jwt.NewWithClaims(jwt.SigningMethodHS512, at_claims)
	acc_token, err := access_token.SignedString(jwtKey)
	if err != nil {
		return nil, err
	}

	//Генерируем refresh-токен с времененм жизни 6 часов
	rt_exp_time := time.Now().Add(360 * time.Minute)

	rt_claims := &Claims{
		Token_id: token_id,
		UserGuid: guid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: rt_exp_time.Unix(),
		},
	}

	refresh_token := jwt.NewWithClaims(jwt.SigningMethodHS256, rt_claims)

	ref_token, err := refresh_token.SignedString(jwtKey)
	if err != nil {
		return nil, err
	}

	//Пишем данные в БД
	WriteTokenToDB(token_id, guid, ref_token, rt_exp_time)

	//Возвращаем пару access-refresh токенов
	return map[string]string{
		"access_token":  acc_token,
		"refresh_token": ref_token,
	}, nil

}
