package server

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

	"github.com/go-playground/validator"

	"github.com/dinorain/useraja/config"
	"github.com/dinorain/useraja/internal/middlewares"
	sessRepository "github.com/dinorain/useraja/internal/session/repository"
	sessUseCase "github.com/dinorain/useraja/internal/session/usecase"
	userDeliveryHTTP "github.com/dinorain/useraja/internal/user/delivery/http/service"
	userRepository "github.com/dinorain/useraja/internal/user/repository"
	userUseCase "github.com/dinorain/useraja/internal/user/usecase"
	"github.com/dinorain/useraja/pkg/logger"
)

type Server struct {
	logger      logger.Logger
	cfg         *config.Config
	v           *validator.Validate
	echo        *echo.Echo
	mw          middlewares.MiddlewareManager
	db          *sqlx.DB
	redisClient *redis.Client
}

// Server constructor
func NewAuthServer(logger logger.Logger, cfg *config.Config, db *sqlx.DB, redisClient *redis.Client) *Server {
	return &Server{
		logger:      logger,
		cfg:         cfg,
		v:           validator.New(),
		echo:        echo.New(),
		db:          db,
		redisClient: redisClient,
	}
}

// Run service
func (s *Server) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	s.mw = middlewares.NewMiddlewareManager(s.logger, s.cfg)
	userRepo := userRepository.NewUserPGRepository(s.db)
	sessRepo := sessRepository.NewSessionRepository(s.redisClient, s.cfg)
	userRedisRepo := userRepository.NewUserRedisRepo(s.redisClient, s.logger)
	userUC := userUseCase.NewUserUseCase(s.cfg, s.logger, userRepo, userRedisRepo)
	sessUC := sessUseCase.NewSessionUseCase(sessRepo, s.cfg)

	l, err := net.Listen("tcp", s.cfg.Server.Port)
	if err != nil {
		return err
	}
	defer l.Close()

	userHandlers := userDeliveryHTTP.NewUserHandlersHTTP(s.echo.Group("user"), s.logger, s.cfg, s.mw, s.v, userUC, sessUC)
	userHandlers.UserMapRoutes()

	adminHandlers := userDeliveryHTTP.NewUserHandlersHTTP(s.echo.Group("admin/user"), s.logger, s.cfg, s.mw, s.v, userUC, sessUC)
	adminHandlers.AdminMapRoutes()

	go func() {
		if err := s.runHttpServer(); err != nil {
			s.logger.Errorf(" s.runHttpServer: %v", err)
			cancel()
		}
	}()
	s.logger.Infof("API Gateway is listening on PORT: %s", s.cfg.Http.Port)

	<-ctx.Done()
	if err := s.echo.Server.Shutdown(ctx); err != nil {
		s.logger.WarnMsg("echo.Server.Shutdown", err)
	}

	return nil
}
