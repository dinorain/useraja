package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/dinorain/useraja/internal/models"
	"github.com/dinorain/useraja/internal/user"
	"github.com/dinorain/useraja/pkg/grpc_errors"
	"github.com/dinorain/useraja/pkg/logger"
)

// Auth redis repository
type userRedisRepo struct {
	redisClient *redis.Client
	basePrefix  string
	logger      logger.Logger
}

var _ user.UserRedisRepository = (*userRedisRepo)(nil)

// Auth redis repository constructor
func NewUserRedisRepo(redisClient *redis.Client, logger logger.Logger) *userRedisRepo {
	return &userRedisRepo{redisClient: redisClient, basePrefix: "user:", logger: logger}
}

// Get user by id
func (r *userRedisRepo) GetByIdCtx(ctx context.Context, key string) (*models.User, error) {
	userBytes, err := r.redisClient.Get(ctx, r.createKey(key)).Bytes()
	if err != nil {
		if err != redis.Nil {
			return nil, grpc_errors.ErrNotFound
		}
		return nil, err
	}
	user := &models.User{}
	if err = json.Unmarshal(userBytes, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Cache user with duration in seconds
func (r *userRedisRepo) SetUserCtx(ctx context.Context, key string, seconds int, user *models.User) error {
	userBytes, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return r.redisClient.Set(ctx, r.createKey(key), userBytes, time.Second*time.Duration(seconds)).Err()
}

// Delete user by key
func (r *userRedisRepo) DeleteUserCtx(ctx context.Context, key string) error {
	return r.redisClient.Del(ctx, r.createKey(key)).Err()
}

func (r *userRedisRepo) createKey(value string) string {
	return fmt.Sprintf("%s: %s", r.basePrefix, value)
}
