package repository

import (
	"avito-pvz-service/internal/database"
	"errors"
	"time"

	"github.com/google/uuid"
)

type PVZ struct {
	ID               string    `json:"id"`
	RegistrationDate time.Time `json:"registration_date"`
	City             string    `json:"city"`
}

var allowedCities = map[string]bool{
	"Москва":          true,
	"Санкт-Петербург": true,
	"Казань":          true,
}

// PVZRecord объединяет данные ПВЗ и связанные с ним приёмки.
type PVZRecord struct {
	PVZ        PVZ               `json:"pvz"`
	Receptions []ReceptionRecord `json:"receptions"`
}

// ReceptionRecord объединяет данные приемки и список товаров.
type ReceptionRecord struct {
	Reception Reception `json:"reception"`
	Products  []Product `json:"products"`
}

func CreatePVZ(city string) (*PVZ, error) {
	if !allowedCities[city] {
		return nil, errors.New("ПВЗ можно завести только в Москве, Санкт-Петербурге или Казани")
	}

	id := uuid.New().String()
	registrationDate := time.Now()

	query := "INSERT INTO pvz (id, registration_date, city) VALUES ($1, $2, $3)"
	_, err := database.DB.Exec(query, id, registrationDate, city)
	if err != nil {
		return nil, err
	}

	return &PVZ{
		ID:               id,
		RegistrationDate: registrationDate,
		City:             city,
	}, nil
}

func GetPVZRecords(startDate, endDate *time.Time, page, limit int) ([]PVZRecord, error) {
	if startDate == nil || endDate == nil {
		return nil, errors.New("startDate and endDate parameters are required")
	}
	offset := (page - 1) * limit

	// Извлекаем список уникальных ПВЗ, у которых есть приёмки в указанном диапазоне.
	rows, err := database.DB.Query(`
        SELECT DISTINCT p.id, p.registration_date, p.city
        FROM pvz p
        JOIN receptions r ON p.id = r.pvz_id
        WHERE r.date_time BETWEEN $1 AND $2
        ORDER BY p.registration_date DESC
        OFFSET $3 LIMIT $4`,
		*startDate, *endDate, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []PVZRecord
	for rows.Next() {
		var pvz PVZ
		err := rows.Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City)
		if err != nil {
			return nil, err
		}

		// Извлекаем приёмки для данного ПВЗ в указанном диапазоне.
		recRows, err := database.DB.Query(`
            SELECT id, date_time, pvz_id, status
            FROM receptions
            WHERE pvz_id = $1 AND date_time BETWEEN $2 AND $3
            ORDER BY date_time DESC`,
			pvz.ID, *startDate, *endDate)
		if err != nil {
			return nil, err
		}
		var receptions []ReceptionRecord
		for recRows.Next() {
			var rec Reception
			err := recRows.Scan(&rec.ID, &rec.DateTime, &rec.PVZId, &rec.Status)
			if err != nil {
				recRows.Close()
				return nil, err
			}
			// Извлекаем товары для данной приёмки.
			prodRows, err := database.DB.Query(`
                SELECT id, date_time, type, reception_id, pvz_id
                FROM products
                WHERE reception_id = $1
                ORDER BY date_time ASC`, rec.ID)
			if err != nil {
				recRows.Close()
				return nil, err
			}
			var products []Product
			for prodRows.Next() {
				var prod Product
				err := prodRows.Scan(&prod.ID, &prod.DateTime, &prod.Type, &prod.ReceptionId, &prod.PVZId)
				if err != nil {
					prodRows.Close()
					recRows.Close()
					return nil, err
				}
				products = append(products, prod)
			}
			prodRows.Close()
			receptions = append(receptions, ReceptionRecord{
				Reception: rec,
				Products:  products,
			})
		}
		recRows.Close()

		records = append(records, PVZRecord{
			PVZ:        pvz,
			Receptions: receptions,
		})
	}

	return records, nil
}

func GetAllPVZ() ([]PVZ, error) {
    rows, err := database.DB.Query("SELECT id, registration_date, city FROM pvz")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var result []PVZ
    for rows.Next() {
        var p PVZ
        if err := rows.Scan(&p.ID, &p.RegistrationDate, &p.City); err != nil {
            return nil, err
        }
        result = append(result, p)
    }
    return result, nil
}