package repository

import (
	"avito-pvz-service/internal/database"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Reception struct {
	ID       string    `json:"id"`
	DateTime time.Time `json:"date_time"`
	PVZId    string    `json:"pvz_id"`
	Status   string    `json:"status"`
}

func CreateReception(pvzId string) (*Reception, error) {
	var status string
	err := database.DB.QueryRow("SELECT status FROM receptions WHERE pvz_id = $1 ORDER BY date_time DESC LIMIT 1", pvzId).Scan(&status)
	if err == nil {
		if status == "in_progress" {
			return nil, errors.New("Нельзя создать новую приёмку: предыдущая не закрыта")
		}
	} else if err.Error() != "sql: no rows in result set" {
		return nil, err
	}

	id := uuid.New().String()
	dateTime := time.Now()

	_, err = database.DB.Exec(
		"INSERT INTO receptions (id, date_time, pvz_id, status) VALUES ($1, $2, $3, $4)",
		id, dateTime, pvzId, "in_progress",
	)
	if err != nil {
		return nil, err
	}

	return &Reception{
		ID:       id,
		DateTime: dateTime,
		PVZId:    pvzId,
		Status:   "in_progress",
	}, nil
}

func CloseReception(pvzId string) (*Reception, error) {
	var reception Reception
	err := database.DB.QueryRow("SELECT id, date_time, pvz_id, status FROM receptions WHERE pvz_id = $1 ORDER BY date_time DESC LIMIT 1", pvzId).
		Scan(&reception.ID, &reception.DateTime, &reception.PVZId, &reception.Status)
	if err != nil {
		return nil, errors.New("Нет приемки для закрытия")
	}
	if reception.Status != "in_progress" {
		return nil, errors.New("Приемка уже закрыта")
	}

	_, err = database.DB.Exec("UPDATE receptions SET status = 'close' WHERE id = $1", reception.ID)
	if err != nil {
		return nil, err
	}

	reception.Status = "close"
	return &reception, nil
}
