package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"

	"github.com/go-playground/validator"

	"github.com/dinorain/useraja/api_gateway_service/internal/middlewares"
	"github.com/dinorain/useraja/config"
	"github.com/dinorain/useraja/internal/interceptors"
	"github.com/dinorain/useraja/pkg/logger"
)

type server struct {
	log  logger.Logger
	cfg  *config.Config
	v    *validator.Validate
	mw   middlewares.MiddlewareManager
	im   interceptors.InterceptorManager
	echo *echo.Echo
}

func NewServer(log logger.Logger, cfg *config.Config) *server {
	return &server{log: log, cfg: cfg, echo: echo.New(), v: validator.New()}
}

func (s *server) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	s.mw = middlewares.NewMiddlewareManager(s.log, s.cfg)

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
