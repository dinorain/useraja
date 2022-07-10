package server

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	"github.com/dinorain/useraja/config"
	"github.com/dinorain/useraja/internal/interceptors"
	sessRepository "github.com/dinorain/useraja/internal/session/repository"
	sessUseCase "github.com/dinorain/useraja/internal/session/usecase"
	authServerGRPC "github.com/dinorain/useraja/internal/user/delivery/grpc/service"
	userRepository "github.com/dinorain/useraja/internal/user/repository"
	userUseCase "github.com/dinorain/useraja/internal/user/usecase"
	"github.com/dinorain/useraja/pkg/logger"
	userService "github.com/dinorain/useraja/proto"
)

// GRPC Auth Server
type Server struct {
	logger      logger.Logger
	cfg         *config.Config
	db          *sqlx.DB
	redisClient *redis.Client
}

// Server constructor
func NewAuthServer(logger logger.Logger, cfg *config.Config, db *sqlx.DB, redisClient *redis.Client) *Server {
	return &Server{
		logger: logger,
		cfg: cfg,
		db: db,
		redisClient: redisClient,
	}
}

// Run service
func (s *Server) Run() error {
	im := interceptors.NewInterceptorManager(s.logger, s.cfg)
	userRepo := userRepository.NewUserPGRepository(s.db)
	sessRepo := sessRepository.NewSessionRepository(s.redisClient, s.cfg)
	userRedisRepo := userRepository.NewUserRedisRepo(s.redisClient, s.logger)
	userUC := userUseCase.NewUserUseCase(s.logger, userRepo, userRedisRepo)
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
