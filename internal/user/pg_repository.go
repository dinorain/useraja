package user

import (
	"context"

	"github.com/google/uuid"

	"github.com/dinorain/useraja/internal/models"
)

// User pg repository
type UserPGRepository interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindById(ctx context.Context, userID uuid.UUID) (*models.User, error)
	UpdateById(ctx context.Context, user *models.User) (*models.User, error)
	DeleteById(ctx context.Context, userID uuid.UUID) error
}
