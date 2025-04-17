package repository

import (
	"avito-pvz-service/internal/database"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateReception_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	original := database.DB
	database.DB = db
	defer func() { database.DB = original }()

	pvzId := "pvz-123"

	// 1. Нет открытой приёмки
	mock.ExpectQuery("SELECT status FROM receptions").
		WithArgs(pvzId).
		WillReturnRows(sqlmock.NewRows([]string{})) // пусто

	// 2. Успешный INSERT
	mock.ExpectExec("INSERT INTO receptions").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), pvzId, "in_progress").
		WillReturnResult(sqlmock.NewResult(1, 1))

	reception, err := CreateReception(pvzId)
	require.NoError(t, err)
	assert.Equal(t, "in_progress", reception.Status)
	assert.Equal(t, pvzId, reception.PVZId)
}

func TestCreateReception_AlreadyInProgress(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	original := database.DB
	database.DB = db
	defer func() { database.DB = original }()

	pvzId := "pvz-456"

	// Возвращаем уже открытую приёмку
	mock.ExpectQuery("SELECT status FROM receptions").
		WithArgs(pvzId).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("in_progress"))

	reception, err := CreateReception(pvzId)
	assert.Nil(t, reception)
	assert.EqualError(t, err, "Нельзя создать новую приёмку: предыдущая не закрыта")
}

func TestCreateReception_SelectError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	original := database.DB
	database.DB = db
	defer func() { database.DB = original }()

	mock.ExpectQuery("SELECT status FROM receptions").
		WithArgs("pvz-error").
		WillReturnError(errors.New("db select error"))

	reception, err := CreateReception("pvz-error")
	assert.Nil(t, reception)
	assert.EqualError(t, err, "db select error")
}

func TestCreateReception_InsertError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	original := database.DB
	database.DB = db
	defer func() { database.DB = original }()

	pvzId := "pvz-insert-fail"

	mock.ExpectQuery("SELECT status FROM receptions").
		WithArgs(pvzId).
		WillReturnRows(sqlmock.NewRows([]string{}))

	mock.ExpectExec("INSERT INTO receptions").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), pvzId, "in_progress").
		WillReturnError(errors.New("insert failed"))

	reception, err := CreateReception(pvzId)
	assert.Nil(t, reception)
	assert.EqualError(t, err, "insert failed")
}
