package usecase

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/dinorain/useraja/config"
	"github.com/dinorain/useraja/internal/models"
	"github.com/dinorain/useraja/internal/user"
	"github.com/dinorain/useraja/internal/user/delivery/http/dto"
	"github.com/dinorain/useraja/pkg/grpc_errors"
	"github.com/dinorain/useraja/pkg/logger"
)

const (
	userByIdCacheDuration = 3600
)

// User UseCase
type userUseCase struct {
	cfg        *config.Config
	logger     logger.Logger
	userPgRepo user.UserPGRepository
	redisRepo  user.UserRedisRepository
}

var _ user.UserUseCase = (*userUseCase)(nil)

// New User UseCase
func NewUserUseCase(cfg *config.Config, logger logger.Logger, userRepo user.UserPGRepository, redisRepo user.UserRedisRepository) *userUseCase {
	return &userUseCase{cfg: cfg, logger: logger, userPgRepo: userRepo, redisRepo: redisRepo}
}

// Register new user
func (u *userUseCase) Register(ctx context.Context, user *models.User) (*models.User, error) {
	existsUser, err := u.userPgRepo.FindByEmail(ctx, user.Email)
	if existsUser != nil || err == nil {
		return nil, grpc_errors.ErrEmailExists
	}

	return u.userPgRepo.Create(ctx, user)
}

// Find use by email address
func (u *userUseCase) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	findByEmail, err := u.userPgRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errors.Wrap(err, "userPgRepo.FindByEmail")
	}

	findByEmail.SanitizePassword()

	return findByEmail, nil
}

// Find use by uuid
func (u *userUseCase) FindById(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	cachedUser, err := u.redisRepo.GetByIDCtx(ctx, userID.String())
	if err != nil && !errors.Is(err, redis.Nil) {
		u.logger.Errorf("redisRepo.GetByIDCtx", err)
	}
	if cachedUser != nil {
		return cachedUser, nil
	}

	foundUser, err := u.userPgRepo.FindById(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "userPgRepo.FindById")
	}

	if err := u.redisRepo.SetUserCtx(ctx, foundUser.UserID.String(), userByIdCacheDuration, foundUser); err != nil {
		u.logger.Errorf("redisRepo.SetUserCtx", err)
	}

	return foundUser, nil
}

// Login user with email and password
func (u *userUseCase) Login(ctx context.Context, email string, password string) (*models.User, error) {
	foundUser, err := u.userPgRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errors.Wrap(err, "userPgRepo.FindByEmail")
	}

	if err := foundUser.ComparePasswords(password); err != nil {
		return nil, errors.Wrap(err, "user.ComparePasswords")
	}

	return foundUser, err
}

func (u *userUseCase) GenerateTokenPair(user *models.User, sessionID string) (*dto.RefreshTokenResponseDto, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["session_id"] = sessionID
	claims["user_id"] = user.UserID
	claims["email"] = user.Email
	claims["role"] = user.Role
	claims["exp"] = time.Now().Add(time.Minute * 15).Unix()

	t, err := token.SignedString([]byte(u.cfg.Server.JwtSecretKey))
	if err != nil {
		return nil, err
	}

	refreshToken := jwt.New(jwt.SigningMethodHS256)
	rtClaims := refreshToken.Claims.(jwt.MapClaims)
	rtClaims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	rt, err := refreshToken.SignedString([]byte(u.cfg.Server.JwtSecretKey))
	if err != nil {
		return nil, err
	}

	return &dto.RefreshTokenResponseDto{
		AccessToken:  t,
		RefreshToken: rt,
	}, nil
}
