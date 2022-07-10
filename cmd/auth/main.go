package main

import (
	"log"
	"os"

	"github.com/dinorain/useraja/config"
	"github.com/dinorain/useraja/internal/server"
	"github.com/dinorain/useraja/pkg/logger"
	"github.com/dinorain/useraja/pkg/postgres"
	"github.com/dinorain/useraja/pkg/redis"
	"github.com/dinorain/useraja/pkg/utils"
)

func main() {
	log.Println("Starting auth microservice")

	configPath := utils.GetConfigPath(os.Getenv("CONFIG"))
	cfg, err := config.GetConfig(configPath)
	if err != nil {
		log.Fatalf("Loading config: %v", err)
	}

	appLogger := logger.NewAppLogger(cfg)
	appLogger.InitLogger()
	appLogger.Infof(
		"AppVersion: %s, LogLevel: %s, Mode: %s, SSL: %v",
		cfg.Server.AppVersion,
		cfg.Logger.Level,
		cfg.Server.Mode,
		cfg.Server.SSL,
	)
	appLogger.Infof("Success parsed config: %#v", cfg.Server.AppVersion)

	redisClient := redis.NewRedisClient(cfg)
	defer redisClient.Close()
	appLogger.Info("Redis connected")

	psqlDB, err := postgres.NewPsqlDB(cfg)
	if err != nil {
		appLogger.Fatalf("Postgresql init: %s", err)
	}
	defer psqlDB.Close()

	authServer := server.NewAuthServer(appLogger, cfg, psqlDB, redisClient)
	appLogger.Fatal(authServer.Run())
}
