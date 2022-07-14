//go:generate mockgen -source redis_repository.go -destination mock/redis_repository.go -package mock
package user

import (
	"context"

	"github.com/dinorain/useraja/internal/models"
)

// User Redis repository interface
type UserRedisRepository interface {
	GetByIdCtx(ctx context.Context, key string) (*models.User, error)
	SetUserCtx(ctx context.Context, key string, seconds int, user *models.User) error
	DeleteUserCtx(ctx context.Context, key string) error
}
