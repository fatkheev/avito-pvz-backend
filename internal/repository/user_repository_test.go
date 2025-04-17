package repository

import (
	"avito-pvz-service/internal/database"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUser_Success(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    original := database.DB
    database.DB = db
    defer func() { database.DB = original }()

    email := "test@example.com"
    password := "pass123"
    role := "client"

    mock.ExpectQuery(`SELECT EXISTS\(`).
        WithArgs(email).
        WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

    mock.ExpectExec(`INSERT INTO users`).
        WithArgs(sqlmock.AnyArg(), email, sqlmock.AnyArg(), role, sqlmock.AnyArg()).
        WillReturnResult(sqlmock.NewResult(1, 1))

    user, err := CreateUser(email, password, role)
    require.NoError(t, err)
    assert.Equal(t, email, user.Email)
    assert.Equal(t, role, user.Role)
    assert.WithinDuration(t, time.Now(), user.CreatedAt, time.Second)
}

func TestCreateUser_Duplicate(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    original := database.DB
    database.DB = db
    defer func() { database.DB = original }()

    email := "duplicate@example.com"

    mock.ExpectQuery(`SELECT EXISTS\(`).
        WithArgs(email).
        WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

    user, err := CreateUser(email, "pass", "moderator")
    assert.Nil(t, user)
    assert.EqualError(t, err, "user with this email already exists")
}
