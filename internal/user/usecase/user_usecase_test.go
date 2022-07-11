package usecase

import (
	"context"
	"database/sql"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/dinorain/useraja/config"
	"github.com/dinorain/useraja/internal/models"
	"github.com/dinorain/useraja/internal/user/mock"
	"github.com/dinorain/useraja/pkg/logger"
)

func TestUserUseCase_Register(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userPGRepository := mock.NewMockUserPGRepository(ctrl)
	userRedisRepository := mock.NewMockUserRedisRepository(ctrl)
	apiLogger := logger.NewAppLogger(nil)

	cfg := &config.Config{Server: config.ServerConfig{JwtSecretKey: "secret123"}}
	userUC := NewUserUseCase(cfg, apiLogger, userPGRepository, userRedisRepository)

	userID := uuid.New()
	mockUser := &models.User{
		Email:     "email@gmail.com",
		FirstName: "FirstName",
		LastName:  "LastName",
		Role:      "admin",
		Avatar:    nil,
		Password:  "123456",
	}

	ctx := context.Background()

	userPGRepository.EXPECT().FindByEmail(gomock.Any(), mockUser.Email).Return(nil, sql.ErrNoRows)

	userPGRepository.EXPECT().Create(gomock.Any(), mockUser).Return(&models.User{
		UserID:    userID,
		Email:     "email@gmail.com",
		FirstName: "FirstName",
		LastName:  "LastName",
		Role:      "admin",
		Avatar:    nil,
		Password:  "123456",
	}, nil)

	createdUser, err := userUC.Register(ctx, mockUser)
	require.NoError(t, err)
	require.NotNil(t, createdUser)
	require.Equal(t, createdUser.UserID, userID)
}

func TestUserUseCase_FindByEmail(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userPGRepository := mock.NewMockUserPGRepository(ctrl)
	userRedisRepository := mock.NewMockUserRedisRepository(ctrl)
	apiLogger := logger.NewAppLogger(nil)

	cfg := &config.Config{Server: config.ServerConfig{JwtSecretKey: "secret123"}}
	userUC := NewUserUseCase(cfg, apiLogger, userPGRepository, userRedisRepository)

	userID := uuid.New()
	mockUser := &models.User{
		UserID:    userID,
		Email:     "email@gmail.com",
		FirstName: "FirstName",
		LastName:  "LastName",
		Role:      "admin",
		Avatar:    nil,
		Password:  "123456",
	}

	ctx := context.Background()

	userPGRepository.EXPECT().FindByEmail(gomock.Any(), mockUser.Email).Return(mockUser, nil)

	user, err := userUC.FindByEmail(ctx, mockUser.Email)
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, user.Email, mockUser.Email)
}

func TestUserUseCase_Login(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userPGRepository := mock.NewMockUserPGRepository(ctrl)
	userRedisRepository := mock.NewMockUserRedisRepository(ctrl)
	apiLogger := logger.NewAppLogger(nil)

	cfg := &config.Config{Server: config.ServerConfig{JwtSecretKey: "secret123"}}
	userUC := NewUserUseCase(cfg, apiLogger, userPGRepository, userRedisRepository)

	userID := uuid.New()
	mockUser := &models.User{
		UserID:    userID,
		Email:     "email@gmail.com",
		FirstName: "FirstName",
		LastName:  "LastName",
		Role:      "admin",
		Avatar:    nil,
		Password:  "123456",
	}

	ctx := context.Background()

	userPGRepository.EXPECT().FindByEmail(gomock.Any(), mockUser.Email).Return(mockUser, nil)
	_, err := userUC.Login(ctx, mockUser.Email, mockUser.Password)
	require.NotNil(t, err)
}

func TestUserUseCase_FindById(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userPGRepository := mock.NewMockUserPGRepository(ctrl)
	userRedisRepository := mock.NewMockUserRedisRepository(ctrl)
	apiLogger := logger.NewAppLogger(nil)

	cfg := &config.Config{}
	userUC := NewUserUseCase(cfg, apiLogger, userPGRepository, userRedisRepository)

	userID := uuid.New()
	mockUser := &models.User{
		UserID:    userID,
		Email:     "email@gmail.com",
		FirstName: "FirstName",
		LastName:  "LastName",
		Role:      "admin",
		Avatar:    nil,
		Password:  "123456",
	}

	ctx := context.Background()

	userRedisRepository.EXPECT().GetByIDCtx(gomock.Any(), mockUser.UserID.String()).AnyTimes().Return(nil, redis.Nil)
	userPGRepository.EXPECT().FindById(gomock.Any(), mockUser.UserID).Return(mockUser, nil)

	user, err := userUC.FindById(ctx, mockUser.UserID)
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, user.UserID, mockUser.UserID)

	userRedisRepository.EXPECT().GetByIDCtx(gomock.Any(), mockUser.UserID.String()).AnyTimes().Return(nil, redis.Nil)
}

func TestUserUseCase_CachedFindById(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userPGRepository := mock.NewMockUserPGRepository(ctrl)
	userRedisRepository := mock.NewMockUserRedisRepository(ctrl)
	apiLogger := logger.NewAppLogger(nil)

	cfg := &config.Config{}
	userUC := NewUserUseCase(cfg, apiLogger, userPGRepository, userRedisRepository)

	userID := uuid.New()
	mockUser := &models.User{
		UserID:    userID,
		Email:     "email@gmail.com",
		FirstName: "FirstName",
		LastName:  "LastName",
		Role:      "admin",
		Avatar:    nil,
		Password:  "123456",
	}

	ctx := context.Background()

	userRedisRepository.EXPECT().GetByIDCtx(gomock.Any(), mockUser.UserID.String()).AnyTimes().Return(nil, redis.Nil)
	userPGRepository.EXPECT().FindById(gomock.Any(), mockUser.UserID).Return(mockUser, nil)
	userRedisRepository.EXPECT().SetUserCtx(gomock.Any(), mockUser.UserID.String(), 3600, mockUser).AnyTimes().Return(nil)

	user, err := userUC.CachedFindById(ctx, mockUser.UserID)
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, user.UserID, mockUser.UserID)
}

func TestUserUseCase_UpdateById(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userPGRepository := mock.NewMockUserPGRepository(ctrl)
	userRedisRepository := mock.NewMockUserRedisRepository(ctrl)
	apiLogger := logger.NewAppLogger(nil)

	cfg := &config.Config{}
	userUC := NewUserUseCase(cfg, apiLogger, userPGRepository, userRedisRepository)

	userID := uuid.New()
	mockUser := &models.User{
		UserID:    userID,
		Email:     "email@gmail.com",
		FirstName: "FirstName",
		LastName:  "LastName",
		Role:      "admin",
		Avatar:    nil,
		Password:  "123456",
	}

	ctx := context.Background()

	userPGRepository.EXPECT().UpdateById(gomock.Any(), mockUser).Return(mockUser, nil)
	userRedisRepository.EXPECT().SetUserCtx(gomock.Any(), mockUser.UserID.String(), 3600, mockUser).AnyTimes().Return(nil)

	user, err := userUC.UpdateById(ctx, mockUser)
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, user.UserID, mockUser.UserID)
}

func TestUserUseCase_DeleteById(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userPGRepository := mock.NewMockUserPGRepository(ctrl)
	userRedisRepository := mock.NewMockUserRedisRepository(ctrl)
	apiLogger := logger.NewAppLogger(nil)

	cfg := &config.Config{}
	userUC := NewUserUseCase(cfg, apiLogger, userPGRepository, userRedisRepository)

	userID := uuid.New()
	mockUser := &models.User{
		UserID:    userID,
		Email:     "email@gmail.com",
		FirstName: "FirstName",
		LastName:  "LastName",
		Role:      "admin",
		Avatar:    nil,
		Password:  "123456",
	}

	ctx := context.Background()

	userPGRepository.EXPECT().DeleteById(gomock.Any(), mockUser.UserID).Return(nil)
	userRedisRepository.EXPECT().DeleteUserCtx(gomock.Any(), mockUser.UserID.String()).AnyTimes().Return(nil)

	err := userUC.DeleteById(ctx, mockUser.UserID)
	require.NoError(t, err)

	userPGRepository.EXPECT().FindById(gomock.Any(), mockUser.UserID).AnyTimes().Return(nil, nil)
	userRedisRepository.EXPECT().GetByIDCtx(gomock.Any(), mockUser.UserID.String()).AnyTimes().Return(nil, redis.Nil)
}

func TestUserUseCase_GenerateTokenPair(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userPGRepository := mock.NewMockUserPGRepository(ctrl)
	userRedisRepository := mock.NewMockUserRedisRepository(ctrl)
	apiLogger := logger.NewAppLogger(nil)

	cfg := &config.Config{}
	userUC := NewUserUseCase(cfg, apiLogger, userPGRepository, userRedisRepository)

	userID := uuid.New()
	mockUser := &models.User{
		UserID:    userID,
		Email:     "email@gmail.com",
		FirstName: "FirstName",
		LastName:  "LastName",
		Role:      "admin",
		Avatar:    nil,
		Password:  "123456",
	}

	at, rt, err := userUC.GenerateTokenPair(mockUser, mockUser.UserID.String())
	require.NoError(t, err)
	require.NotEqual(t, at, "")
	require.NotEqual(t, rt, "")
}
