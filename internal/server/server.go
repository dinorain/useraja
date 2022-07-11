package server

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"

	"github.com/go-playground/validator"

	"github.com/dinorain/useraja/config"
	"github.com/dinorain/useraja/internal/interceptors"
	"github.com/dinorain/useraja/internal/middlewares"
	sessRepository "github.com/dinorain/useraja/internal/session/repository"
	sessUseCase "github.com/dinorain/useraja/internal/session/usecase"
	authServerGRPC "github.com/dinorain/useraja/internal/user/delivery/grpc/service"
	userDeliveryHTTP "github.com/dinorain/useraja/internal/user/delivery/http/service"
	userRepository "github.com/dinorain/useraja/internal/user/repository"
	userUseCase "github.com/dinorain/useraja/internal/user/usecase"
	"github.com/dinorain/useraja/pkg/logger"
	userService "github.com/dinorain/useraja/proto"
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

	im := interceptors.NewInterceptorManager(s.logger, s.cfg)
	s.mw = middlewares.NewMiddlewareManager(s.logger, s.cfg)
	userRepo := userRepository.NewUserPGRepository(s.db)
	sessRepo := sessRepository.NewSessionRepository(s.redisClient, s.cfg)
	userRedisRepo := userRepository.NewUserRedisRepo(s.redisClient, s.logger)
	userUC := userUseCase.NewUserUseCase(s.cfg, s.logger, userRepo, userRedisRepo)
	sessUC := sessUseCase.NewSessionUseCase(sessRepo, s.cfg)

	l, err := net.Listen("tcp", s.cfg.Http.Port)
	if err != nil {
		return err
	}
	defer l.Close()

	userHandlers := userDeliveryHTTP.NewUserHandlersHTTP(s.echo.Group("user"), s.logger, s.cfg, s.mw, s.v, userUC, sessUC)
	userHandlers.UserMapRoutes()

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

	server := grpc.NewServer(grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle: s.cfg.Server.MaxConnectionIdle * time.Minute,
		Timeout:           s.cfg.Server.Timeout * time.Second,
		MaxConnectionAge:  s.cfg.Server.MaxConnectionAge * time.Minute,
		Time:              s.cfg.Server.Timeout * time.Minute,
	}),
		grpc.UnaryInterceptor(im.Logger),
		grpc.ChainUnaryInterceptor(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpcrecovery.UnaryServerInterceptor(),
		),
	)

	if s.cfg.Server.Mode != "Production" {
		reflection.Register(server)
	}

	authGRPCServer := authServerGRPC.NewAuthServerGRPC(s.logger, s.cfg, userUC, sessUC)
	userService.RegisterUserServiceServer(server, authGRPCServer)

	go func() {
		s.logger.Infof("Server is listening on port: %v", s.cfg.Server.Port)
		if err := server.Serve(l); err != nil {
			s.logger.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	server.GracefulStop()
	s.logger.Info("Server Exited Properly")

	return nil
}

// Run GRPC service
func (s *Server) RunGRPC() error {
	im := interceptors.NewInterceptorManager(s.logger, s.cfg)
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

	server := grpc.NewServer(grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle: s.cfg.Server.MaxConnectionIdle * time.Minute,
		Timeout:           s.cfg.Server.Timeout * time.Second,
		MaxConnectionAge:  s.cfg.Server.MaxConnectionAge * time.Minute,
		Time:              s.cfg.Server.Timeout * time.Minute,
	}),
		grpc.UnaryInterceptor(im.Logger),
		grpc.ChainUnaryInterceptor(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpcrecovery.UnaryServerInterceptor(),
		),
	)

	if s.cfg.Server.Mode != "Production" {
		reflection.Register(server)
	}

	authGRPCServer := authServerGRPC.NewAuthServerGRPC(s.logger, s.cfg, userUC, sessUC)
	userService.RegisterUserServiceServer(server, authGRPCServer)

	go func() {
		s.logger.Infof("Server is listening on port: %v", s.cfg.Server.Port)
		if err := server.Serve(l); err != nil {
			s.logger.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	server.GracefulStop()
	s.logger.Info("Server Exited Properly")

	return nil
}
