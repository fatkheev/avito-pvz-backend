package repository

import (
	"avito-pvz-service/internal/database"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePVZ_AllowedCity(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// временно подменим базу в package database
	original := database.DB
	database.DB = db
	defer func() { database.DB = original }()

	city := "Москва"

	mock.ExpectExec("INSERT INTO pvz").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), city).
		WillReturnResult(sqlmock.NewResult(1, 1))

	pvz, err := CreatePVZ(city)
	require.NoError(t, err)
	assert.Equal(t, city, pvz.City)
	assert.WithinDuration(t, time.Now(), pvz.RegistrationDate, time.Second)
}

func TestCreatePVZ_DisallowedCity(t *testing.T) {
	// этот тест не использует базу, можно без моков
	pvz, err := CreatePVZ("Новосибирск")
	assert.Nil(t, pvz)
	assert.EqualError(t, err, "ПВЗ можно завести только в Москве, Санкт-Петербурге или Казани")
}

func TestCreatePVZ_SQLFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	original := database.DB
	database.DB = db
	defer func() { database.DB = original }()

	mock.ExpectExec("INSERT INTO pvz").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "Казань").
		WillReturnError(errors.New("db insert failed"))

	pvz, err := CreatePVZ("Казань")
	assert.Nil(t, pvz)
	assert.EqualError(t, err, "db insert failed")
}
