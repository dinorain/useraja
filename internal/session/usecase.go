//go:generate mockgen -source usecase.go -destination mock/usecase.go -package mock
package session

import (
	"context"

	"github.com/dinorain/useraja/internal/models"
)

// Session UseCase
type SessUseCase interface {
	CreateSession(ctx context.Context, session *models.Session, expire int) (string, error)
	GetSessionByID(ctx context.Context, sessionID string) (*models.Session, error)
	DeleteByID(ctx context.Context, sessionID string) error
}
