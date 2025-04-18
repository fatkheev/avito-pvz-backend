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

func TestAddProduct_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	database.DB = db
	pvzID := "1234-pvz"
	receptionID := "5678-reception"

	// mock получения последней открытой приёмки
	mock.ExpectQuery(`SELECT id, status FROM receptions`).
		WithArgs(pvzID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow(receptionID, "in_progress"))

	// mock вставки продукта
	mock.ExpectExec(`INSERT INTO products`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "электроника", receptionID, pvzID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	product, err := AddProduct(pvzID, "электроника")
	require.NoError(t, err)
	assert.Equal(t, "электроника", product.Type)
	assert.Equal(t, receptionID, product.ReceptionId)
	assert.Equal(t, pvzID, product.PVZId)
	assert.WithinDuration(t, time.Now(), product.DateTime, time.Second)
}

func TestAddProduct_InvalidType(t *testing.T) {
	product, err := AddProduct("any", "мебель")
	assert.Nil(t, product)
	assert.EqualError(t, err, "Invalid product type")
}

func TestAddProduct_NoReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	database.DB = db

	mock.ExpectQuery(`SELECT id, status FROM receptions`).
		WithArgs("pvz-1").
		WillReturnError(errors.New("no rows"))

	product, err := AddProduct("pvz-1", "одежда")
	assert.Nil(t, product)
	assert.EqualError(t, err, "Нет активной приемки")
}

func TestAddProduct_ReceptionClosed(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	database.DB = db

	mock.ExpectQuery(`SELECT id, status FROM receptions`).
		WithArgs("pvz-2").
		WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow("abc", "close"))

	product, err := AddProduct("pvz-2", "обувь")
	assert.Nil(t, product)
	assert.EqualError(t, err, "Нет активной приемки")
}

func TestAddProduct_InsertError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	database.DB = db

	mock.ExpectQuery(`SELECT id, status FROM receptions`).
		WithArgs("pvz-3").
		WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow("r1", "in_progress"))

	mock.ExpectExec(`INSERT INTO products`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "обувь", "r1", "pvz-3").
		WillReturnError(errors.New("insert error"))

	product, err := AddProduct("pvz-3", "обувь")
	assert.Nil(t, product)
	assert.EqualError(t, err, "insert error")
}

func TestDeleteLastProduct_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	database.DB = db
	pvzID := "pvz-1"
	receptionID := "reception-1"
	productID := "product-1"

	// Получить последнюю приёмку
	mock.ExpectQuery(`SELECT id, status FROM receptions`).
		WithArgs(pvzID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow(receptionID, "in_progress"))

	// Найти последний товар
	mock.ExpectQuery(`SELECT id FROM products`).
		WithArgs(receptionID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(productID))

	// Удалить товар
	mock.ExpectExec(`DELETE FROM products`).
		WithArgs(productID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = DeleteLastProduct(pvzID)
	assert.NoError(t, err)
}

func TestDeleteLastProduct_NoReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	database.DB = db

	mock.ExpectQuery(`SELECT id, status FROM receptions`).
		WithArgs("pvz-x").
		WillReturnError(errors.New("sql: no rows in result set"))

	err = DeleteLastProduct("pvz-x")
	assert.EqualError(t, err, "Нет активной приемки")
}

func TestDeleteLastProduct_ReceptionClosed(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	database.DB = db

	mock.ExpectQuery(`SELECT id, status FROM receptions`).
		WithArgs("pvz-y").
		WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow("reception-closed", "close"))

	err = DeleteLastProduct("pvz-y")
	assert.EqualError(t, err, "Приемка уже закрыта")
}

func TestDeleteLastProduct_NoProducts(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	database.DB = db
	pvzID := "pvz-2"
	receptionID := "reception-2"

	mock.ExpectQuery(`SELECT id, status FROM receptions`).
		WithArgs(pvzID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow(receptionID, "in_progress"))

	mock.ExpectQuery(`SELECT id FROM products`).
		WithArgs(receptionID).
		WillReturnError(errors.New("no products"))

	err = DeleteLastProduct(pvzID)
	assert.EqualError(t, err, "Нет товаров для удаления")
}

func TestDeleteLastProduct_DeleteFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	database.DB = db
	pvzID := "pvz-3"
	receptionID := "reception-3"
	productID := "product-3"

	mock.ExpectQuery(`SELECT id, status FROM receptions`).
		WithArgs(pvzID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).AddRow(receptionID, "in_progress"))

	mock.ExpectQuery(`SELECT id FROM products`).
		WithArgs(receptionID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(productID))

	mock.ExpectExec(`DELETE FROM products`).
		WithArgs(productID).
		WillReturnError(errors.New("delete failed"))

	err = DeleteLastProduct(pvzID)
	assert.EqualError(t, err, "delete failed")
}
