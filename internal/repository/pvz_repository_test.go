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

	// временно подменим базу
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

func TestGetPVZRecords_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	database.DB = db

	start := time.Date(2025, 4, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 4, 17, 23, 59, 59, 0, time.UTC)

	// ПВЗ
	mock.ExpectQuery(`SELECT DISTINCT p\.id, p\.registration_date, p\.city FROM pvz p JOIN receptions r ON p\.id = r\.pvz_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}).
			AddRow("pvz-1", time.Now(), "Москва"))

	// Приёмки
	mock.ExpectQuery(`SELECT id, date_time, pvz_id, status FROM receptions`).
		WithArgs("pvz-1", start, end).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
			AddRow("reception-1", time.Now(), "pvz-1", "in_progress"))

	// Товары
	mock.ExpectQuery(`SELECT id, date_time, type, reception_id, pvz_id FROM products WHERE reception_id = .* ORDER BY date_time ASC`).
		WithArgs("reception-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "type", "reception_id", "pvz_id"}).
			AddRow("product-1", time.Now(), "одежда", "reception-1", "pvz-1"))

	result, err := GetPVZRecords(&start, &end, 1, 10)
	require.NoError(t, err)
	require.Len(t, result, 1)

	rec := result[0]
	assert.Equal(t, "Москва", rec.PVZ.City)
	require.Len(t, rec.Receptions, 1)
	assert.Equal(t, "reception-1", rec.Receptions[0].Reception.ID)
	assert.Equal(t, "одежда", rec.Receptions[0].Products[0].Type)
}

func TestGetPVZRecords_PVZQueryFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	database.DB = db

	start, end := time.Now(), time.Now()

	mock.ExpectQuery(`SELECT DISTINCT p\.id, p\.registration_date, p\.city FROM pvz p JOIN receptions r ON p\.id = r\.pvz_id`).
		WillReturnError(errors.New("pvz error"))

	result, err := GetPVZRecords(&start, &end, 1, 10)
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pvz error")
}

func TestGetPVZRecords_ReceptionsQueryFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	database.DB = db

	start, end := time.Now(), time.Now()

	mock.ExpectQuery(`SELECT DISTINCT p\.id, p\.registration_date, p\.city FROM pvz p JOIN receptions r ON p\.id = r\.pvz_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}).
			AddRow("pvz-1", time.Now(), "Москва"))

	mock.ExpectQuery(`SELECT id, date_time, pvz_id, status FROM receptions`).
		WithArgs("pvz-1", start, end).
		WillReturnError(errors.New("reception error"))

	result, err := GetPVZRecords(&start, &end, 1, 10)
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reception error")
}

func TestGetPVZRecords_ProductsQueryFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	database.DB = db

	start, end := time.Now(), time.Now()

	mock.ExpectQuery(`SELECT DISTINCT p\.id, p\.registration_date, p\.city FROM pvz p JOIN receptions r ON p\.id = r\.pvz_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}).
			AddRow("pvz-1", time.Now(), "Москва"))

	mock.ExpectQuery(`SELECT id, date_time, pvz_id, status FROM receptions`).
		WithArgs("pvz-1", start, end).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
			AddRow("reception-1", time.Now(), "pvz-1", "in_progress"))

	mock.ExpectQuery(`SELECT id, date_time, type, reception_id, pvz_id FROM products WHERE reception_id = .* ORDER BY date_time ASC`).
		WithArgs("reception-1").
		WillReturnError(errors.New("product error"))

	result, err := GetPVZRecords(&start, &end, 1, 10)
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "product error")
}