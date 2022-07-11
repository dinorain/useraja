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
	"github.com/dinorain/useraja/pkg/grpc_errors"
	"github.com/dinorain/useraja/pkg/logger"
	"github.com/dinorain/useraja/pkg/utils"
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

// FindAll find users
func (u *userUseCase) FindAll(ctx context.Context, pagination *utils.Pagination) ([]models.User, error) {
	users, err := u.userPgRepo.FindAll(ctx, pagination)
	if err != nil {
		return nil, errors.Wrap(err, "userPgRepo.FindAll")
	}

	return users, nil
}

// FindByEmail find user by email address
func (u *userUseCase) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	findByEmail, err := u.userPgRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errors.Wrap(err, "userPgRepo.FindByEmail")
	}

	findByEmail.SanitizePassword()

	return findByEmail, nil
}

// FindById find user by uuid
func (u *userUseCase) FindById(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	foundUser, err := u.userPgRepo.FindById(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "userPgRepo.FindById")
	}

	return foundUser, nil
}

// CachedFindById find user by uuid from cache
func (u *userUseCase) CachedFindById(ctx context.Context, userID uuid.UUID) (*models.User, error) {
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

// UpdateById update user by uuid
func (u *userUseCase) UpdateById(ctx context.Context, user *models.User) (*models.User, error) {
	updatedUser, err := u.userPgRepo.UpdateById(ctx, user)
	if err != nil {
		return nil, errors.Wrap(err, "userPgRepo.UpdateById")
	}

	if err := u.redisRepo.SetUserCtx(ctx, updatedUser.UserID.String(), userByIdCacheDuration, updatedUser); err != nil {
		u.logger.Errorf("redisRepo.SetUserCtx", err)
	}

	updatedUser.SanitizePassword()

	return updatedUser, nil
}

// DeleteById delete user by uuid
func (u *userUseCase) DeleteById(ctx context.Context, userID uuid.UUID) error {
	err := u.userPgRepo.DeleteById(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "userPgRepo.DeleteById")
	}

	if err := u.redisRepo.DeleteUserCtx(ctx, userID.String()); err != nil {
		u.logger.Errorf("redisRepo.DeleteUserCtx", err)
	}

	return nil
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

func (u *userUseCase) GenerateTokenPair(user *models.User, sessionID string) (access string, refresh string, err error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["session_id"] = sessionID
	claims["user_id"] = user.UserID
	claims["email"] = user.Email
	claims["role"] = user.Role
	claims["exp"] = time.Now().Add(time.Minute * 15).Unix()

	access, err = token.SignedString([]byte(u.cfg.Server.JwtSecretKey))
	if err != nil {
		return "", "", err
	}

	refreshToken := jwt.New(jwt.SigningMethodHS256)
	rtClaims := refreshToken.Claims.(jwt.MapClaims)
	rtClaims["session_id"] = sessionID
	rtClaims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	refresh, err = refreshToken.SignedString([]byte(u.cfg.Server.JwtSecretKey))
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil
}
