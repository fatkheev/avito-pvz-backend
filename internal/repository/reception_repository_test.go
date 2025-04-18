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

func TestCreateReception_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	original := database.DB
	database.DB = db
	defer func() { database.DB = original }()

	pvzId := "pvz-123"

	// Нет открытой приёмки
	mock.ExpectQuery("SELECT status FROM receptions").
		WithArgs(pvzId).
		WillReturnRows(sqlmock.NewRows([]string{})) // пусто

	// Успешный INSERT
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

func TestCloseReception_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	database.DB = db

	pvzID := "pvz-1"
	receptionID := "rec-1"
	now := time.Now()

	// Найти последнюю приёмку
	mock.ExpectQuery(`SELECT id, date_time, pvz_id, status FROM receptions`).
		WithArgs(pvzID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
			AddRow(receptionID, now, pvzID, "in_progress"))

	// Обновить статус
	mock.ExpectExec(`UPDATE receptions SET status = 'close' WHERE id =`).
		WithArgs(receptionID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	rec, err := CloseReception(pvzID)
	require.NoError(t, err)
	assert.Equal(t, "close", rec.Status)
	assert.Equal(t, receptionID, rec.ID)
	assert.Equal(t, pvzID, rec.PVZId)
	assert.WithinDuration(t, now, rec.DateTime, time.Second)
}

func TestCloseReception_NoReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	database.DB = db

	mock.ExpectQuery(`SELECT id, date_time, pvz_id, status FROM receptions`).
		WithArgs("pvz-404").
		WillReturnError(errors.New("sql: no rows in result set"))

	rec, err := CloseReception("pvz-404")
	assert.Nil(t, rec)
	assert.EqualError(t, err, "Нет приемки для закрытия")
}

func TestCloseReception_AlreadyClosed(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	database.DB = db

	pvzID := "pvz-2"
	now := time.Now()

	mock.ExpectQuery(`SELECT id, date_time, pvz_id, status FROM receptions`).
		WithArgs(pvzID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
			AddRow("rec-closed", now, pvzID, "close"))

	rec, err := CloseReception(pvzID)
	assert.Nil(t, rec)
	assert.EqualError(t, err, "Приемка уже закрыта")
}

func TestCloseReception_UpdateFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	database.DB = db

	pvzID := "pvz-3"
	recID := "rec-3"
	now := time.Now()

	mock.ExpectQuery(`SELECT id, date_time, pvz_id, status FROM receptions`).
		WithArgs(pvzID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
			AddRow(recID, now, pvzID, "in_progress"))

	mock.ExpectExec(`UPDATE receptions SET status = 'close' WHERE id =`).
		WithArgs(recID).
		WillReturnError(errors.New("update error"))

	rec, err := CloseReception(pvzID)
	assert.Nil(t, rec)
	assert.EqualError(t, err, "update error")
}