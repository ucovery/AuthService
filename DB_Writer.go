package main

import (
	"log"
	"time"

	"context"

	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
)

// Модель данных для записи в БД
type MonogoWriteModel struct {
	OperationID  xid.ID
	UserGuid     string
	RefreshToken string
	TokenExpTime time.Time
}

//Пишем refresh-токен в collection базы данных
func WriteTokenToDB(tokenID xid.ID, userGuid string, RefreshToken string, TokenExpTime time.Time) {
	//Получаем коннект от БД
	collection, err := GetDatabaseConnection()
	if err != nil {
		log.Fatal(err)
	}

	encripted_token, err := bcrypt.GenerateFromPassword([]byte(RefreshToken), 5)
	if err != nil {
		log.Fatal(err)
	}

	data_to_write := MonogoWriteModel{tokenID, userGuid, string(encripted_token), TokenExpTime}

	_, err = collection.InsertOne(context.TODO(), data_to_write)
	if err != nil {
		log.Fatal(err)
	}
}
