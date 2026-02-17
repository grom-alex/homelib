package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grom-alex/homelib/backend/internal/models"
)

func newTestUser() *models.User {
	return &models.User{
		Email:        "test@example.com",
		Username:     "testuser",
		DisplayName:  "Test User",
		PasswordHash: "$2a$10$hashedpassword",
		Role:         "user",
		IsActive:     true,
	}
}

func TestRegisterUser_FirstUserBecomesAdmin(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewUserRepo(mock)
	user := newTestUser()

	mock.ExpectBegin()
	mock.ExpectExec("LOCK TABLE users IN EXCLUSIVE MODE").
		WillReturnResult(pgxmock.NewResult("LOCK TABLE", 0))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("INSERT INTO users").
		WithArgs(user.Email, user.Username, user.DisplayName, user.PasswordHash, "admin").
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("uuid-1", time.Now(), time.Now()))
	mock.ExpectCommit()

	err = repo.RegisterUser(context.Background(), user, true)
	require.NoError(t, err)
	assert.Equal(t, "admin", user.Role)
	assert.Equal(t, "uuid-1", user.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRegisterUser_SecondUserGetsUserRole(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewUserRepo(mock)
	user := newTestUser()

	mock.ExpectBegin()
	mock.ExpectExec("LOCK TABLE users IN EXCLUSIVE MODE").
		WillReturnResult(pgxmock.NewResult("LOCK TABLE", 0))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("INSERT INTO users").
		WithArgs(user.Email, user.Username, user.DisplayName, user.PasswordHash, "user").
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("uuid-2", time.Now(), time.Now()))
	mock.ExpectCommit()

	err = repo.RegisterUser(context.Background(), user, true)
	require.NoError(t, err)
	assert.Equal(t, "user", user.Role)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRegisterUser_DuplicateEmailReturnsError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewUserRepo(mock)
	user := newTestUser()

	mock.ExpectBegin()
	mock.ExpectExec("LOCK TABLE users IN EXCLUSIVE MODE").
		WillReturnResult(pgxmock.NewResult("LOCK TABLE", 0))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("INSERT INTO users").
		WithArgs(user.Email, user.Username, user.DisplayName, user.PasswordHash, "user").
		WillReturnError(fmt.Errorf("duplicate key value violates unique constraint"))
	mock.ExpectRollback()

	err = repo.RegisterUser(context.Background(), user, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRegisterUser_RegistrationDisabledBlocksNonFirst(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewUserRepo(mock)
	user := newTestUser()

	mock.ExpectBegin()
	mock.ExpectExec("LOCK TABLE users IN EXCLUSIVE MODE").
		WillReturnResult(pgxmock.NewResult("LOCK TABLE", 0))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectRollback()

	err = repo.RegisterUser(context.Background(), user, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "registration is disabled")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRegisterUser_RegistrationDisabledAllowsFirst(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewUserRepo(mock)
	user := newTestUser()

	mock.ExpectBegin()
	mock.ExpectExec("LOCK TABLE users IN EXCLUSIVE MODE").
		WillReturnResult(pgxmock.NewResult("LOCK TABLE", 0))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("INSERT INTO users").
		WithArgs(user.Email, user.Username, user.DisplayName, user.PasswordHash, "admin").
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("uuid-first", time.Now(), time.Now()))
	mock.ExpectCommit()

	err = repo.RegisterUser(context.Background(), user, false)
	require.NoError(t, err)
	assert.Equal(t, "admin", user.Role)
	assert.NoError(t, mock.ExpectationsWereMet())
}
