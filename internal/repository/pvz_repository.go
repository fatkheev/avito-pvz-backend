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
