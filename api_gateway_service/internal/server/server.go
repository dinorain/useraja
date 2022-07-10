package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

	"github.com/go-playground/validator"

	"github.com/dinorain/useraja/api_gateway_service/internal/middlewares"
	"github.com/dinorain/useraja/config"
	"github.com/dinorain/useraja/internal/interceptors"
	sessRepository "github.com/dinorain/useraja/internal/session/repository"
	sessUseCase "github.com/dinorain/useraja/internal/session/usecase"
	userDeliveryHTTP "github.com/dinorain/useraja/internal/user/delivery/http/service"
	userRepository "github.com/dinorain/useraja/internal/user/repository"
	userUseCase "github.com/dinorain/useraja/internal/user/usecase"
	"github.com/dinorain/useraja/pkg/logger"
)

type server struct {
	log         logger.Logger
	cfg         *config.Config
	v           *validator.Validate
	mw          middlewares.MiddlewareManager
	im          *interceptors.InterceptorManager
	echo        *echo.Echo
	db          *sqlx.DB
	redisClient *redis.Client
}

func NewServer(log logger.Logger, cfg *config.Config) *server {
	return &server{log: log, cfg: cfg, echo: echo.New(), v: validator.New()}
}

func (s *server) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	s.mw = middlewares.NewMiddlewareManager(s.log, s.cfg)
	s.im = interceptors.NewInterceptorManager(s.log, s.cfg)

	userRepo := userRepository.NewUserPGRepository(s.db)
	sessRepo := sessRepository.NewSessionRepository(s.redisClient, s.cfg)
	userRedisRepo := userRepository.NewUserRedisRepo(s.redisClient, s.log)
	userUC := userUseCase.NewUserUseCase(s.log, userRepo, userRedisRepo)
	sessUC := sessUseCase.NewSessionUseCase(sessRepo, s.cfg)

	userHandlers := userDeliveryHTTP.NewUserHandlersHTTP(s.echo.Group("user"), s.log, s.cfg, userUC, sessUC)
	userHandlers.MapRoutes()

	go func() {
		if err := s.runHttpServer(); err != nil {
			s.log.Errorf(" s.runHttpServer: %v", err)
			cancel()
		}
	}()
	s.log.Infof("API Gateway is listening on PORT: %s", s.cfg.Http.Port)

	<-ctx.Done()
	if err := s.echo.Server.Shutdown(ctx); err != nil {
		s.log.WarnMsg("echo.Server.Shutdown", err)
	}

	return nil
}
