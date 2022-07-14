//go:generate mockgen -source redis_repository.go -destination mock/redis_repository.go -package mock
package session

import (
	"context"

	"github.com/dinorain/useraja/internal/models"
)

// Session repository
type SessRepository interface {
	CreateSession(ctx context.Context, session *models.Session, expire int) (string, error)
	GetSessionById(ctx context.Context, sessionID string) (*models.Session, error)
	DeleteById(ctx context.Context, sessionID string) error
}
