package repository

import (
	"database/sql"
	"errors"
	"time"

	"avito-pvz-service/internal/database"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        string
	Email     string
	Password  string
	Role      string
	CreatedAt time.Time
}

func CreateUser(email, password, role string) (*User, error) {
	var exists bool
	err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", email).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("user with this email already exists")
	}

	// хэширую пароль с использованием bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	id := uuid.New().String()
	createdAt := time.Now()
	_, err = database.DB.Exec(
		"INSERT INTO users (id, email, password, role, created_at) VALUES ($1, $2, $3, $4, $5)",
		id, email, string(hashedPassword), role, createdAt,
	)
	if err != nil {
		return nil, err
	}

	return &User{ID: id, Email: email, Password: string(hashedPassword), Role: role, CreatedAt: createdAt}, nil
}

func GetUserByEmail(email string) (*User, error) {
	var user User
	err := database.DB.QueryRow("SELECT id, email, password, role, created_at FROM users WHERE email=$1", email).
		Scan(&user.ID, &user.Email, &user.Password, &user.Role, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}
