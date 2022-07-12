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

// Register
// @Tags Users
// @Summary To register user
// @Description To create user, admin only
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param payload body dto.UserRegisterRequestDto true "Payload"
// @Success 200 {object} dto.UserRegisterResponseDto
// @Router /user [post]
func (h *userHandlersHTTP) Register() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		createDto := &dto.UserRegisterRequestDto{}
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

		return c.JSON(http.StatusCreated, dto.UserRegisterResponseDto{UserID: createdUser.UserID})
	}
}

// Login
// @Tags Users
// @Summary User login
// @Description User login with email and password
// @Accept json
// @Produce json
// @Param payload body dto.UserLoginRequestDto true "Payload"
// @Success 200 {object} dto.UserLoginResponseDto
// @Router /user/login [post]
func (h *userHandlersHTTP) Login() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		loginDto := &dto.UserLoginRequestDto{}
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

		accessToken, refreshToken, err := h.userUC.GenerateTokenPair(user, session)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, dto.UserLoginResponseDto{UserID: user.UserID, Tokens: &dto.UserRefreshTokenResponseDto{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}})
	}
}

// FindAll
// @Tags Users
// @Summary Find all users
// @Description Find all users, admin only
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param size query string false "pagination size"
// @Param page query string false "pagination page"
// @Success 200 {object} dto.UserFindResponseDto
// @Router /user [get]
func (h *userHandlersHTTP) FindAll() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		pq := utils.NewPaginationFromQueryParams(c.QueryParam(constants.Size), c.QueryParam(constants.Page))
		users, err := h.userUC.FindAll(ctx, pq)
		if err != nil {
			h.logger.Errorf("userUC.FindAll: %v", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		return c.JSON(http.StatusOK, dto.UserFindResponseDto{
			Data: users,
			Meta: utils.PaginationMetaDto{
				Limit:  pq.GetLimit(),
				Offset: pq.GetOffset(),
				Page:   pq.GetPage(),
			},
		})
	}
}

// FindByID
// @Tags Users
// @Summary Find user
// @Description Find existing user by id
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} dto.UserResponseDto
// @Router /user/{id} [get]
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

// UpdateByID
// @Tags Users
// @Summary Update user
// @Description Update existing user
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "User ID"
// @Param payload body dto.UserUpdateRequestDto true "Payload"
// @Success 200 {object} dto.UserResponseDto
// @Router /user/{id} [put]
func (h *userHandlersHTTP) UpdateByID() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		userUUID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			h.logger.WarnMsg("uuid.FromString", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		updateDto := &dto.UserUpdateRequestDto{}
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

// DeleteByID
// @Tags Users
// @Summary Delete user
// @Description Delete existing user
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} nil
// @Param id path string true "User ID"
// @Router /user/{id} [delete]
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

// GetMe
// @Tags Users
// @Summary Find me
// @Description Get session id from token, find user by uuid and returns it
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} dto.UserResponseDto
// @Router /user/me [get]
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

// Logout
// @Tags Users
// @Summary User logout
// @Description Delete current session
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} nil
// @Router /user/logout [post]
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

// RefreshToken
// @Tags Users
// @Summary Refresh access token
// @Description Refresh access token
// @Accept json
// @Produce json
// @Param payload body dto.UserRefreshTokenDto true "Payload"
// @Success 200 {object} dto.UserRefreshTokenResponseDto
// @Router /user/refresh [post]
func (h *userHandlersHTTP) RefreshToken() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		refreshTokenDto := &dto.UserRefreshTokenDto{}
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

		accessToken, refreshToken, err := h.userUC.GenerateTokenPair(user, sessID)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, dto.UserRefreshTokenResponseDto{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		})
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

func (h *userHandlersHTTP) registerReqToUserModel(r *dto.UserRegisterRequestDto) (*models.User, error) {
	userCandidate := &models.User{
		Email:     r.Email,
		FirstName: r.FirstName,
		LastName:  r.LastName,
		Role:      r.Role,
		Avatar:    nil,
		Password:  r.Password,
	}

	if err := userCandidate.PrepareCreate(); err != nil {
		return nil, err
	}

	return userCandidate, nil
}

func (h *userHandlersHTTP) updateReqToUserModel(updateCandidate *models.User, r *dto.UserUpdateRequestDto) (*models.User, error) {

	if r.FirstName != nil {
		updateCandidate.FirstName = strings.TrimSpace(*r.FirstName)
	}
	if r.LastName != nil {
		updateCandidate.LastName = strings.TrimSpace(*r.LastName)
	}
	if r.Avatar != nil {
		avatar := strings.TrimSpace(*r.Avatar)
		updateCandidate.Avatar = &avatar
	}
	if r.Password != nil {
		updateCandidate.Password = *r.Password
		if err := updateCandidate.HashPassword(); err != nil {
			return nil, err
		}
	}

	return updateCandidate, nil
}
