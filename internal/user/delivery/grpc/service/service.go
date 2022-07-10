package service

import (
	"github.com/dinorain/useraja/config"
	"github.com/dinorain/useraja/internal/session"
	"github.com/dinorain/useraja/internal/user"
	"github.com/dinorain/useraja/pkg/logger"
)

type usersService struct {
	logger logger.Logger
	cfg    *config.Config
	userUC user.UserUseCase
	sessUC session.SessUseCase
}

// Auth service constructor
func NewAuthServerGRPC(logger logger.Logger, cfg *config.Config, userUC user.UserUseCase, sessUC session.SessUseCase) *usersService {
	return &usersService{logger: logger, cfg: cfg, userUC: userUC, sessUC: sessUC}
}
