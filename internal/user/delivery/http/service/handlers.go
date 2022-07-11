package http

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/dinorain/useraja/config"
	"github.com/dinorain/useraja/internal/middlewares"
	"github.com/dinorain/useraja/internal/models"
	"github.com/dinorain/useraja/internal/session"
	"github.com/dinorain/useraja/internal/user"
	"github.com/dinorain/useraja/internal/user/delivery/http/dto"
	"github.com/dinorain/useraja/pkg/constants"
	httpErrors "github.com/dinorain/useraja/pkg/http_errors"
	"github.com/dinorain/useraja/pkg/logger"
	"github.com/dinorain/useraja/pkg/utils"
)

type userHandlersHTTP struct {
	group  *echo.Group
	logger logger.Logger
	cfg    *config.Config
	mw     middlewares.MiddlewareManager
	v      *validator.Validate
	userUC user.UserUseCase
	sessUC session.SessUseCase
}

func NewUserHandlersHTTP(
	group *echo.Group,
	logger logger.Logger,
	cfg *config.Config,
	mw middlewares.MiddlewareManager,
	v *validator.Validate,
	userUC user.UserUseCase,
	sessUC session.SessUseCase,
) *userHandlersHTTP {
	return &userHandlersHTTP{group: group, logger: logger, cfg: cfg, mw: mw, v: v, userUC: userUC, sessUC: sessUC}
}

func (h *userHandlersHTTP) Register() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		createDto := &dto.RegisterRequestDto{}
		if err := c.Bind(createDto); err != nil {
			h.logger.WarnMsg("bind", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		if err := h.v.StructCtx(ctx, createDto); err != nil {
			h.logger.WarnMsg("validate", err)
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
			h.logger.WarnMsg("bind", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		if err := h.v.StructCtx(ctx, loginDto); err != nil {
			h.logger.WarnMsg("validate", err)
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

		tokens, err := h.userUC.GenerateTokenPair(user, session)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, dto.LoginResponseDto{UserID: user.UserID, Tokens: tokens})
	}
}

// FindAll find users
func (h *userHandlersHTTP) FindAll() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		pq := utils.NewPaginationFromQueryParams(c.QueryParam(constants.Size), c.QueryParam(constants.Page))
		users, err := h.userUC.FindAll(ctx, pq)
		if err != nil {
			h.logger.Errorf("userUC.FindAll: %v", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		return c.JSON(http.StatusOK, users)
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

		user, err := h.userUC.CachedFindById(ctx, userUUID)
		if err != nil {
			h.logger.Errorf("userUC.CachedFindById: %v", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		return c.JSON(http.StatusOK, dto.UserResponseFromModel(user))
	}
}

// UpdateByID update user by uuid
func (h *userHandlersHTTP) UpdateByID() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		userUUID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			h.logger.WarnMsg("uuid.FromString", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		updateDto := &dto.UpdateRequestDto{}
		if err := c.Bind(updateDto); err != nil {
			h.logger.WarnMsg("bind", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		if err := h.v.StructCtx(ctx, updateDto); err != nil {
			h.logger.WarnMsg("validate", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		user, err := h.userUC.FindById(ctx, userUUID)
		if err != nil {
			h.logger.Errorf("userUC.FindById: %v", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		user, err = h.updateReqToUserModel(user, updateDto)
		if err != nil {
			h.logger.Errorf("updateReqToUserModel: %v", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		user, err = h.userUC.UpdateById(ctx, user)
		if err != nil {
			h.logger.Errorf("userUC.UpdateById: %v", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		return c.JSON(http.StatusOK, dto.UserResponseFromModel(user))
	}
}

// DeleteByID delete user by uuid
func (h *userHandlersHTTP) DeleteByID() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		userUUID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			h.logger.WarnMsg("uuid.FromString", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		if err := h.userUC.DeleteById(ctx, userUUID); err != nil {
			h.logger.Errorf("userUC.DeleteById: %v", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		return c.JSON(http.StatusOK, nil)
	}
}

// GetMe to get session id from, ctx metadata, find user by uuid and returns it
func (h *userHandlersHTTP) GetMe() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		sessID, err := h.getSessionIDFromCtx(c)
		if err != nil {
			h.logger.Errorf("getSessionIDFromCtx: %v", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		session, err := h.sessUC.GetSessionByID(ctx, sessID)
		if err != nil {
			h.logger.Errorf("sessUC.GetSessionByID: %v", err)
			if errors.Is(err, redis.Nil) {
				return echo.ErrUnauthorized
			}
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		user, err := h.userUC.CachedFindById(ctx, session.UserID)
		if err != nil {
			h.logger.Errorf("userUC.CachedFindById: %v", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		return c.JSON(http.StatusOK, dto.UserResponseFromModel(user))
	}
}

// Logout user, delete current session
func (h *userHandlersHTTP) Logout() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		sessID, err := h.getSessionIDFromCtx(c)
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

// RefreshToken to refresh tokens
func (h *userHandlersHTTP) RefreshToken() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		refreshTokenDto := &dto.RefreshTokenDto{}
		if err := c.Bind(refreshTokenDto); err != nil {
			h.logger.WarnMsg("bind", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		token, err := jwt.Parse(refreshTokenDto.RefreshToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				h.logger.Errorf("jwt.SigningMethodHMAC: %v", token.Header["alg"])
				return nil, fmt.Errorf("jwt.SigningMethodHMAC: %v", token.Header["alg"])
			}

			return []byte(h.cfg.Server.JwtSecretKey), nil
		})

		if err != nil {
			h.logger.Warnf("jwt.Parse")
			return httpErrors.ErrorCtxResponse(c, errors.New("invalid refresh token"), h.cfg.Http.DebugErrorsResponse)
		}

		if !token.Valid {
			h.logger.Warnf("token.Valid")
			return httpErrors.ErrorCtxResponse(c, errors.New("invalid refresh token"), h.cfg.Http.DebugErrorsResponse)
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			h.logger.Warnf("jwt.MapClaims: %+v", token.Claims)
			return httpErrors.ErrorCtxResponse(c, errors.New("invalid refresh token"), h.cfg.Http.DebugErrorsResponse)
		}

		sessID, ok := claims["session_id"].(string)
		if !ok {
			h.logger.Warnf("session_id: %+v", claims)
			return httpErrors.ErrorCtxResponse(c, errors.New("invalid refresh token"), h.cfg.Http.DebugErrorsResponse)
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

		newTokenPair, err := h.userUC.GenerateTokenPair(user, sessID)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, newTokenPair)
	}
}

func (h *userHandlersHTTP) getSessionIDFromCtx(c echo.Context) (string, error) {
	user, ok := c.Get("user").(*jwt.Token)
	if !ok {
		h.logger.Warnf("jwt.Token: %+v", c.Get("user"))
		return "", errors.New("invalid token header")
	}

	claims, ok := user.Claims.(jwt.MapClaims)
	if !ok {
		h.logger.Warnf("jwt.MapClaims: %+v", c.Get("user"))
		return "", errors.New("invalid token header")
	}

	sessionID, ok := claims["session_id"].(string)
	if !ok {
		h.logger.Warnf("session_id: %+v", claims)
		return "", errors.New("invalid token header")
	}

	return sessionID, nil
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

func (h *userHandlersHTTP) updateReqToUserModel(updateCandidate *models.User, r *dto.UpdateRequestDto) (*models.User, error) {

	if r.FirstName != nil {
		updateCandidate.FirstName = strings.TrimSpace(*r.FirstName)
	}
	if r.LastName != nil {
		updateCandidate.LastName = strings.TrimSpace(*r.LastName)
	}
	if r.Password != nil {
		updateCandidate.Password = *r.Password
		if err := updateCandidate.HashPassword(); err != nil {
			return nil, err
		}
	}

	return updateCandidate, nil
}
