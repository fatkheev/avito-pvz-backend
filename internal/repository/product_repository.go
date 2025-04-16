package repository

import (
	"avito-pvz-service/internal/database"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          string    `json:"id"`
	DateTime    time.Time `json:"date_time"`
	Type        string    `json:"type"`
	ReceptionId string    `json:"reception_id"`
	PVZId       string    `json:"pvz_id"`
}

var allowedProductTypes = map[string]bool{
	"электроника": true,
	"одежда":      true,
	"обувь":       true,
}

func AddProduct(pvzId, productType string) (*Product, error) {
	if !allowedProductTypes[productType] {
		return nil, errors.New("Invalid product type")
	}

	var receptionId, status string
	err := database.DB.QueryRow(`
	    SELECT id, status 
	    FROM receptions 
	    WHERE pvz_id = $1 
	    ORDER BY date_time DESC 
	    LIMIT 1`, pvzId).Scan(&receptionId, &status)
	if err != nil {
		return nil, errors.New("Нет активной приемки")
	}
	if status != "in_progress" {
		return nil, errors.New("Нет активной приемки")
	}

	id := uuid.New().String()
	dateTime := time.Now()

	// Вставляем запись с указанием reception_id и pvz_id.
	_, err = database.DB.Exec(`
	    INSERT INTO products (id, date_time, type, reception_id, pvz_id)
	    VALUES ($1, $2, $3, $4, $5)`,
		id, dateTime, productType, receptionId, pvzId)
	if err != nil {
		return nil, err
	}

	return &Product{
		ID:          id,
		DateTime:    dateTime,
		Type:        productType,
		ReceptionId: receptionId,
		PVZId:       pvzId,
	}, nil
}

func DeleteLastProduct(pvzId string) error {
    // Сначала находим последнюю приёмку для данного PVZ.
    var receptionId, status string
    err := database.DB.QueryRow(`
        SELECT id, status 
        FROM receptions 
        WHERE pvz_id = $1 
        ORDER BY date_time DESC 
        LIMIT 1`, pvzId).Scan(&receptionId, &status)
    if err != nil {
        return errors.New("Нет активной приемки")
    }
    if status != "in_progress" {
        return errors.New("Приемка уже закрыта")
    }
    // Находим последний добавленный товар в этой приёмке (сортируем по времени добавления)
    var productId string
    err = database.DB.QueryRow(`
        SELECT id FROM products 
        WHERE reception_id = $1 
        ORDER BY date_time DESC 
        LIMIT 1`, receptionId).Scan(&productId)
    if err != nil {
        return errors.New("Нет товаров для удаления")
    }
    // Удаляем найденный товар
    _, err = database.DB.Exec("DELETE FROM products WHERE id = $1", productId)
    if err != nil {
        return err
    }
    return nil
}