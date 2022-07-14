package repository

import (
	"context"
	"log"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/dinorain/useraja/internal/models"
)

func SetupRedis() *userRedisRepo {
	mr, err := miniredis.Run()
	if err != nil {
		log.Fatal(err)
	}
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	userRedisRepository := NewUserRedisRepo(client, nil)
	return userRedisRepository
}

func TestUserRedisRepo_SetUserCtx(t *testing.T) {
	t.Parallel()

	redisRepo := SetupRedis()

	t.Run("SetUserCtx", func(t *testing.T) {
		user := &models.User{
			UserID: uuid.New(),
		}

		err := redisRepo.SetUserCtx(context.Background(), redisRepo.createKey(user.UserID.String()), 10, user)
		require.NoError(t, err)
	})
}

func TestUserRedisRepo_GetByIdCtx(t *testing.T) {
	t.Parallel()

	redisRepo := SetupRedis()

	t.Run("GetByIdCtx", func(t *testing.T) {
		user := &models.User{
			UserID: uuid.New(),
		}

		err := redisRepo.SetUserCtx(context.Background(), redisRepo.createKey(user.UserID.String()), 10, user)
		require.NoError(t, err)

		user, err = redisRepo.GetByIdCtx(context.Background(), redisRepo.createKey(user.UserID.String()))
		require.NoError(t, err)
		require.NotNil(t, user)
	})
}

func TestUserRedisRepo_DeleteUserCtx(t *testing.T) {
	t.Parallel()

	redisRepo := SetupRedis()

	t.Run("DeleteUserCtx", func(t *testing.T) {
		user := &models.User{
			UserID: uuid.New(),
		}

		err := redisRepo.DeleteUserCtx(context.Background(), redisRepo.createKey(user.UserID.String()))
		require.NoError(t, err)
	})
}
