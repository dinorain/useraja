//go:generate mockgen -source pg_repository.go -destination mock/pg_repository.go -package mock
package user

import (
	"context"

	"github.com/google/uuid"

	"github.com/dinorain/useraja/internal/models"
	"github.com/dinorain/useraja/pkg/utils"
)

// User pg repository
type UserPGRepository interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
	FindAll(ctx context.Context, pagination *utils.Pagination) ([]models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindById(ctx context.Context, userID uuid.UUID) (*models.User, error)
	UpdateById(ctx context.Context, user *models.User) (*models.User, error)
	DeleteById(ctx context.Context, userID uuid.UUID) error
}
