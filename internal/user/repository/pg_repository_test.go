package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/dinorain/useraja/internal/models"
	"github.com/dinorain/useraja/pkg/utils"
)

func TestUserRepository_Create(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer sqlxDB.Close()

	userPGRepository := NewUserPGRepository(sqlxDB)

	columns := []string{"user_id", "first_name", "last_name", "email", "password", "avatar", "role", "created_at", "updated_at"}
	userUUID := uuid.New()
	mockUser := &models.User{
		UserID:    userUUID,
		Email:     "email@gmail.com",
		FirstName: "FirstName",
		LastName:  "LastName",
		Role:      "admin",
		Avatar:    nil,
		Password:  "123456",
	}

	rows := sqlmock.NewRows(columns).AddRow(
		userUUID,
		mockUser.FirstName,
		mockUser.LastName,
		mockUser.Email,
		mockUser.Password,
		mockUser.Avatar,
		mockUser.Role,
		time.Now(),
		time.Now(),
	)

	mock.ExpectQuery(createUserQuery).WithArgs(
		mockUser.FirstName,
		mockUser.LastName,
		mockUser.Email,
		mockUser.Password,
		mockUser.Role,
		mockUser.Avatar,
	).WillReturnRows(rows)

	createdUser, err := userPGRepository.Create(context.Background(), mockUser)
	require.NoError(t, err)
	require.NotNil(t, createdUser)
}

func TestUserRepository_FindByEmail(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer sqlxDB.Close()

	userPGRepository := NewUserPGRepository(sqlxDB)

	columns := []string{"user_id", "first_name", "last_name", "email", "password", "avatar", "role", "created_at", "updated_at"}
	userUUID := uuid.New()
	mockUser := &models.User{
		UserID:    userUUID,
		Email:     "email@gmail.com",
		FirstName: "FirstName",
		LastName:  "LastName",
		Role:      "admin",
		Avatar:    nil,
		Password:  "123456",
	}

	rows := sqlmock.NewRows(columns).AddRow(
		userUUID,
		mockUser.FirstName,
		mockUser.LastName,
		mockUser.Email,
		mockUser.Password,
		mockUser.Avatar,
		mockUser.Role,
		time.Now(),
		time.Now(),
	)

	mock.ExpectQuery(findByEmailQuery).WithArgs(mockUser.Email).WillReturnRows(rows)

	foundUser, err := userPGRepository.FindByEmail(context.Background(), mockUser.Email)
	require.NoError(t, err)
	require.NotNil(t, foundUser)
	require.Equal(t, foundUser.Email, mockUser.Email)
}

func TestUserRepository_FindAll(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer sqlxDB.Close()

	userPGRepository := NewUserPGRepository(sqlxDB)

	columns := []string{"user_id", "first_name", "last_name", "email", "password", "avatar", "role", "created_at", "updated_at"}
	userUUID := uuid.New()
	mockUser := &models.User{
		UserID:    userUUID,
		Email:     "email@gmail.com",
		FirstName: "FirstName",
		LastName:  "LastName",
		Role:      "admin",
		Avatar:    nil,
		Password:  "123456",
	}

	rows := sqlmock.NewRows(columns).AddRow(
		userUUID,
		mockUser.FirstName,
		mockUser.LastName,
		mockUser.Email,
		mockUser.Password,
		mockUser.Avatar,
		mockUser.Role,
		time.Now(),
		time.Now(),
	)

	size := 10
	mock.ExpectQuery(findAllQuery).WithArgs(size, 0).WillReturnRows(rows)
	foundUsers, err := userPGRepository.FindAll(context.Background(), utils.NewPaginationQuery(size, 1))
	require.NoError(t, err)
	require.NotNil(t, foundUsers)
	require.Equal(t, len(foundUsers), 1)

	mock.ExpectQuery(findAllQuery).WithArgs(size, 10).WillReturnRows(rows)
	foundUsers, err = userPGRepository.FindAll(context.Background(), utils.NewPaginationQuery(size, 2))
	require.NoError(t, err)
	require.Nil(t, foundUsers)
}

func TestUserRepository_FindById(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer sqlxDB.Close()

	userPGRepository := NewUserPGRepository(sqlxDB)

	columns := []string{"user_id", "first_name", "last_name", "email", "password", "avatar", "role", "created_at", "updated_at"}
	userUUID := uuid.New()
	mockUser := &models.User{
		UserID:    userUUID,
		Email:     "email@gmail.com",
		FirstName: "FirstName",
		LastName:  "LastName",
		Role:      "admin",
		Avatar:    nil,
		Password:  "123456",
	}

	rows := sqlmock.NewRows(columns).AddRow(
		userUUID,
		mockUser.FirstName,
		mockUser.LastName,
		mockUser.Email,
		mockUser.Password,
		mockUser.Avatar,
		mockUser.Role,
		time.Now(),
		time.Now(),
	)

	mock.ExpectQuery(findByIDQuery).WithArgs(mockUser.UserID).WillReturnRows(rows)

	foundUser, err := userPGRepository.FindById(context.Background(), mockUser.UserID)
	require.NoError(t, err)
	require.NotNil(t, foundUser)
	require.Equal(t, foundUser.UserID, mockUser.UserID)
}

func TestUserRepository_UpdateById(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer sqlxDB.Close()

	userPGRepository := NewUserPGRepository(sqlxDB)

	columns := []string{"user_id", "first_name", "last_name", "email", "password", "avatar", "role", "created_at", "updated_at"}
	userUUID := uuid.New()
	mockUser := &models.User{
		UserID:    userUUID,
		Email:     "email@gmail.com",
		FirstName: "FirstName",
		LastName:  "LastName",
		Role:      "admin",
		Avatar:    nil,
		Password:  "123456",
	}

	_ = sqlmock.NewRows(columns).AddRow(
		userUUID,
		mockUser.FirstName,
		mockUser.LastName,
		mockUser.Email,
		mockUser.Password,
		mockUser.Avatar,
		mockUser.Role,
		time.Now(),
		time.Now(),
	)

	mockUser.FirstName = "FirstNameChanged"
	mock.ExpectExec(updateByIDQuery).WithArgs(
		mockUser.UserID,
		mockUser.FirstName,
		mockUser.LastName,
		mockUser.Email,
		mockUser.Password,
		mockUser.Role,
		mockUser.Avatar).WillReturnResult(sqlmock.NewResult(0, 1))

	updatedUser, err := userPGRepository.UpdateById(context.Background(), mockUser)
	require.NoError(t, err)
	require.NotNil(t, mockUser)
	require.Equal(t, updatedUser.FirstName, mockUser.FirstName)
	require.Equal(t, updatedUser.UserID, mockUser.UserID)
}

func TestUserRepository_DeleteById(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer sqlxDB.Close()

	userPGRepository := NewUserPGRepository(sqlxDB)

	columns := []string{"user_id", "first_name", "last_name", "email", "password", "avatar", "role", "created_at", "updated_at"}
	userUUID := uuid.New()
	mockUser := &models.User{
		UserID:    userUUID,
		Email:     "email@gmail.com",
		FirstName: "FirstName",
		LastName:  "LastName",
		Role:      "admin",
		Avatar:    nil,
		Password:  "123456",
	}

	_ = sqlmock.NewRows(columns).AddRow(
		userUUID,
		mockUser.FirstName,
		mockUser.LastName,
		mockUser.Email,
		mockUser.Password,
		mockUser.Avatar,
		mockUser.Role,
		time.Now(),
		time.Now(),
	)

	mock.ExpectExec(deleteByIDQuery).WithArgs(mockUser.UserID).WillReturnResult(sqlmock.NewResult(0, 1))

	err = userPGRepository.DeleteById(context.Background(), mockUser.UserID)
	require.NoError(t, err)
	require.NotNil(t, mockUser)
}
