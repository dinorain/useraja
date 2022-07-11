package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/dinorain/useraja/internal/models"
	"github.com/dinorain/useraja/internal/user"
	"github.com/dinorain/useraja/pkg/utils"
)

// User repository
type UserRepository struct {
	db *sqlx.DB
}

var _ user.UserPGRepository = (*UserRepository)(nil)

// User repository constructor
func NewUserPGRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	createdUser := &models.User{}
	if err := r.db.QueryRowxContext(
		ctx,
		createUserQuery,
		user.FirstName,
		user.LastName,
		user.Email,
		user.Password,
		user.Role,
		user.Avatar,
	).StructScan(createdUser); err != nil {
		return nil, errors.Wrap(err, "Create.QueryRowxContext")
	}

	return createdUser, nil
}

// UpdateById update existing user
func (r *UserRepository) UpdateById(ctx context.Context, user *models.User) (*models.User, error) {
	if res, err := r.db.ExecContext(
		ctx,
		updateByIDQuery,
		user.UserID,
		user.FirstName,
		user.LastName,
		user.Email,
		user.Password,
		user.Role,
		user.Avatar,
	); err != nil {
		return nil, errors.Wrap(err, "Update.ExecContext")
	} else {
		_, err := res.RowsAffected()
		if err != nil {
			return nil, errors.Wrap(err, "Update.RowsAffected")
		}
	}

	return user, nil
}

// FindAll Find users
func (r *UserRepository) FindAll(ctx context.Context, pagination *utils.Pagination) ([]models.User, error) {
	var users []models.User
	if err := r.db.SelectContext(ctx, &users, findAllQuery, pagination.GetLimit(), pagination.GetOffset()); err != nil {
		return nil, errors.Wrap(err, "FindById.SelectContext")
	}

	return users, nil
}

// FindByEmail Find by user email address
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	if err := r.db.GetContext(ctx, user, findByEmailQuery, email); err != nil {
		return nil, errors.Wrap(err, "FindByEmail.GetContext")
	}

	return user, nil
}

// FindById Find user by uuid
func (r *UserRepository) FindById(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user := &models.User{}
	if err := r.db.GetContext(ctx, user, findByIDQuery, userID); err != nil {
		return nil, errors.Wrap(err, "FindById.GetContext")
	}

	return user, nil
}

// DeleteById Find user by uuid
func (r *UserRepository) DeleteById(ctx context.Context, userID uuid.UUID) error {
	if res, err := r.db.ExecContext(ctx, deleteByIDQuery, userID); err != nil {
		return errors.Wrap(err, "DeleteById.ExecContext")
	} else {
		cnt, err := res.RowsAffected()
		if err != nil {
			return errors.Wrap(err, "DeleteById.RowsAffected")
		} else if cnt == 0 {
			return sql.ErrNoRows
		}
	}

	return nil
}
