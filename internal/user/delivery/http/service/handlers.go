package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/dinorain/useraja/config"
	"github.com/dinorain/useraja/internal/models"
	"github.com/dinorain/useraja/internal/session"
	"github.com/dinorain/useraja/internal/user"
	"github.com/dinorain/useraja/internal/user/delivery/http/dto"
	"github.com/dinorain/useraja/pkg/grpc_errors"
	httpErrors "github.com/dinorain/useraja/pkg/http_errors"
	"github.com/dinorain/useraja/pkg/logger"
	"github.com/dinorain/useraja/pkg/utils"
)

type userHandlersHTTP struct {
	group  *echo.Group
	logger logger.Logger
	cfg    *config.Config
	userUC user.UserUseCase
	sessUC session.SessUseCase
}

func NewUserHandlersHTTP(
	group *echo.Group,
	logger logger.Logger,
	cfg *config.Config,
	userUC user.UserUseCase,
	sessUC session.SessUseCase,
) *userHandlersHTTP {
	return &userHandlersHTTP{group: group, logger: logger, cfg: cfg, userUC: userUC, sessUC: sessUC}
}

func (h *userHandlersHTTP) Register() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		createDto := &dto.RegisterRequestDto{}
		if err := c.Bind(createDto); err != nil {
			h.logger.WarnMsg("Bind", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		user, err := h.registerReqToUserModel(createDto)
		if err != nil {
			h.logger.Errorf("registerReqToUserModel: %v", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		createdUser, err := h.userUC.Register(ctx, user)
		if err != nil {
			h.logger.Errorf("userUC.Register: %v", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		return c.JSON(http.StatusCreated, dto.RegisterResponseDto{UserID: createdUser.UserID})
	}
}

// Login user with email and password
func (h *userHandlersHTTP) Login() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		loginDto := &dto.LoginRequestDto{}
		if err := c.Bind(loginDto); err != nil {
			h.logger.WarnMsg("Bind", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		email := loginDto.Email
		if !utils.ValidateEmail(email) {
			h.logger.Errorf("ValidateEmail: %v", email)
			return httpErrors.ErrorCtxResponse(c, errors.New("invalid email"), h.cfg.Http.DebugErrorsResponse)
		}

		user, err := h.userUC.Login(ctx, email, loginDto.Password)
		if err != nil {
			h.logger.Errorf("userUC.Login: %v", email)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		session, err := h.sessUC.CreateSession(ctx, &models.Session{
			UserID: user.UserID,
		}, h.cfg.Session.Expire)
		if err != nil {
			h.logger.Errorf("sessUC.CreateSession: %v", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		return c.JSON(http.StatusCreated, dto.LoginResponseDto{UserID: user.UserID, SessionID: session})
	}
}

// FindByID find user by uuid
func (h *userHandlersHTTP) FindByID() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		userUUID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			h.logger.WarnMsg("uuid.FromString", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		user, err := h.userUC.FindById(ctx, userUUID)
		if err != nil {
			h.logger.Errorf("userUC.FindById: %v", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		return c.JSON(http.StatusOK, dto.UserResponseFromModel(user))
	}
}

// GetMe to get session id from, ctx metadata, find user by uuid and returns it
func (h *userHandlersHTTP) GetMe() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		sessID, err := h.getSessionIDFromCtx(ctx)
		if err != nil {
			h.logger.Errorf("getSessionIDFromCtx: %v", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		session, err := h.sessUC.GetSessionByID(ctx, sessID)
		if err != nil {
			h.logger.Errorf("sessUC.GetSessionByID: %v", err)
			if errors.Is(err, redis.Nil) {
				return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
			}
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		user, err := h.userUC.FindById(ctx, session.UserID)
		if err != nil {
			h.logger.Errorf("userUC.FindById: %v", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		return c.JSON(http.StatusOK, dto.UserResponseFromModel(user))
	}
}

// Logout user, delete current session
func (h *userHandlersHTTP) Logout() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		sessID, err := h.getSessionIDFromCtx(ctx)
		if err != nil {
			h.logger.Errorf("getSessionIDFromCtx: %v", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		if err := h.sessUC.DeleteByID(ctx, sessID); err != nil {
			h.logger.Errorf("sessUC.DeleteByID: %v", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		return c.JSON(http.StatusOK, nil)
	}
}

func (h *userHandlersHTTP) getSessionIDFromCtx(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "metadata.FromIncomingContext: %v", grpc_errors.ErrNoCtxMetaData)
	}

	sessionID := md.Get("session_id")
	if sessionID[0] == "" {
		return "", status.Errorf(codes.PermissionDenied, "md.Get sessionId: %v", grpc_errors.ErrInvalidSessionId)
	}

	return sessionID[0], nil
}

func (h *userHandlersHTTP) registerReqToUserModel(r *dto.RegisterRequestDto) (*models.User, error) {
	userCandidate := &models.User{
		Email:     r.Email,
		FirstName: r.FirstName,
		LastName:  r.LastName,
		Role:      r.Role,
		Avatar:    &r.Avatar,
		Password:  r.Password,
	}

	if err := userCandidate.PrepareCreate(); err != nil {
		return nil, err
	}

	return userCandidate, nil
}
