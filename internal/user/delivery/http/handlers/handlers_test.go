package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/validator"
	"github.com/golang-jwt/jwt"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/require"

	"github.com/dinorain/useraja/config"
	"github.com/dinorain/useraja/internal/middlewares"
	"github.com/dinorain/useraja/internal/models"
	mockSessUC "github.com/dinorain/useraja/internal/session/mock"
	"github.com/dinorain/useraja/internal/user/delivery/http/dto"
	"github.com/dinorain/useraja/internal/user/mock"
	"github.com/dinorain/useraja/pkg/converter"
	"github.com/dinorain/useraja/pkg/logger"
)

func TestUsersService_Register(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mock.NewMockUserUseCase(ctrl)
	sessUC := mockSessUC.NewMockSessUseCase(ctrl)

	appLogger := logger.NewAppLogger(nil)
	mw := middlewares.NewMiddlewareManager(appLogger, nil)

	e := echo.New()
	v := validator.New()
	cfg := &config.Config{Session: config.Session{Expire: 1234}}
	handlers := NewUserHandlersHTTP(e.Group("user"), appLogger, cfg, mw, v, userUC, sessUC)

	reqDto := &dto.UserRegisterRequestDto{
		Email:     "email@gmail.com",
		FirstName: "FirstName",
		LastName:  "LastName",
		Password:  "123456",
		Role:      "user",
	}

	buf := &bytes.Buffer{}
	_ = json.NewEncoder(buf).Encode(reqDto)

	req := httptest.NewRequest(http.MethodPost, "/user", buf)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res := httptest.NewRecorder()
	ctx := e.NewContext(req, res)

	resDto := &dto.UserRegisterResponseDto{
		UserID: uuid.Nil,
	}

	buf, _ = converter.AnyToBytesBuffer(resDto)

	userUC.EXPECT().Register(gomock.Any(), gomock.Any()).AnyTimes().Return(&models.User{}, nil)
	require.NoError(t, handlers.Register()(ctx))
	require.Equal(t, http.StatusCreated, res.Code)
	require.Equal(t, buf.String(), res.Body.String())
}

func TestUsersService_Login(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mock.NewMockUserUseCase(ctrl)
	sessUC := mockSessUC.NewMockSessUseCase(ctrl)

	appLogger := logger.NewAppLogger(nil)
	mw := middlewares.NewMiddlewareManager(appLogger, nil)

	e := echo.New()
	v := validator.New()
	cfg := &config.Config{Session: config.Session{Expire: 1234}}
	handlers := NewUserHandlersHTTP(e.Group("user"), appLogger, cfg, mw, v, userUC, sessUC)

	reqDto := &dto.UserLoginRequestDto{
		Email:    "email@gmail.com",
		Password: "123456",
	}

	var buf bytes.Buffer
	_ = json.NewEncoder(&buf).Encode(reqDto)

	req := httptest.NewRequest(http.MethodPost, "/user/login", &buf)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res := httptest.NewRecorder()
	ctx := e.NewContext(req, res)

	mockUser := &models.User{
		UserID:    uuid.New(),
		Email:     "email@gmail.com",
		FirstName: "FirstName",
		LastName:  "LastName",
		Password:  "123456",
		Role:      "user",
	}

	userUC.EXPECT().Login(gomock.Any(), reqDto.Email, reqDto.Password).AnyTimes().Return(mockUser, nil)
	sessUC.EXPECT().CreateSession(gomock.Any(), &models.Session{UserID: mockUser.UserID}, cfg.Session.Expire).AnyTimes().Return("s", nil)
	userUC.EXPECT().GenerateTokenPair(gomock.Any(), gomock.Any()).AnyTimes().Return("rt", "at", nil)
	require.NoError(t, handlers.Login()(ctx))
	require.Equal(t, http.StatusCreated, res.Code)
}

func TestUsersService_FindAll(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mock.NewMockUserUseCase(ctrl)
	sessUC := mockSessUC.NewMockSessUseCase(ctrl)

	appLogger := logger.NewAppLogger(nil)
	mw := middlewares.NewMiddlewareManager(appLogger, nil)

	e := echo.New()
	v := validator.New()
	cfg := &config.Config{Session: config.Session{Expire: 1234}}
	handlers := NewUserHandlersHTTP(e.Group("user"), appLogger, cfg, mw, v, userUC, sessUC)

	req := httptest.NewRequest(http.MethodGet, "/user", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res := httptest.NewRecorder()
	ctx := e.NewContext(req, res)

	var users []models.User
	users = append(users, models.User{
		UserID:    uuid.New(),
		Email:     "email@gmail.com",
		FirstName: "FirstName",
		LastName:  "LastName",
		Password:  "123456",
		Role:      "user",
	})

	userUC.EXPECT().FindAll(gomock.Any(), gomock.Any()).AnyTimes().Return(users, nil)
	require.NoError(t, handlers.FindAll()(ctx))
	require.Equal(t, http.StatusOK, res.Code)
}

func TestUsersService_FindById(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mock.NewMockUserUseCase(ctrl)
	sessUC := mockSessUC.NewMockSessUseCase(ctrl)

	cfg := &config.Config{Session: config.Session{Expire: 1234}}
	appLogger := logger.NewAppLogger(cfg)
	mw := middlewares.NewMiddlewareManager(appLogger, nil)

	e := echo.New()
	v := validator.New()
	handlers := NewUserHandlersHTTP(e.Group("user"), appLogger, cfg, mw, v, userUC, sessUC)

	req := httptest.NewRequest(http.MethodGet, "/user/:id", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res := httptest.NewRecorder()
	ctx := e.NewContext(req, res)

	ctx.SetParamNames("id")
	ctx.SetParamValues("2ceba62a-35f4-444b-a358-4b14834837e1")

	userUC.EXPECT().CachedFindById(gomock.Any(), gomock.Any()).AnyTimes().Return(&models.User{}, nil)
	require.NoError(t, handlers.FindById()(ctx))
	require.Equal(t, http.StatusOK, res.Code)
}

func TestUsersService_UpdateById(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mock.NewMockUserUseCase(ctrl)
	sessUC := mockSessUC.NewMockSessUseCase(ctrl)

	cfg := &config.Config{Session: config.Session{Expire: 1234}}
	appLogger := logger.NewAppLogger(cfg)
	mw := middlewares.NewMiddlewareManager(appLogger, cfg)

	e := echo.New()
	e.Use(middleware.JWT([]byte("secret")))
	v := validator.New()
	handlers := NewUserHandlersHTTP(e.Group("user"), appLogger, cfg, mw, v, userUC, sessUC)

	change := "changed"
	reqDto := &dto.UserUpdateRequestDto{
		FirstName: &change,
		LastName:  &change,
		Password:  &change,
		Avatar:    &change,
	}

	buf := &bytes.Buffer{}
	_ = json.NewEncoder(buf).Encode(reqDto)

	userUUID := uuid.New()
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["session_id"] = uuid.New().String()
	claims["user_id"] = userUUID.String()
	claims["role"] = "user"
	claims["exp"] = time.Now().Add(time.Minute * 15).Unix()
	validToken, _ := token.SignedString([]byte("secret"))

	req := httptest.NewRequest(http.MethodPost, "/user/:id", buf)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, fmt.Sprintf("bearer %v", validToken))

	t.Run("Forbidden update by other user", func(t *testing.T) {
		t.Parallel()

		res := httptest.NewRecorder()
		ctx := e.NewContext(req, res)

		handler := handlers.UpdateById()
		h := middleware.JWTWithConfig(middleware.JWTConfig{
			Claims:     claims,
			SigningKey: []byte("secret"),
		})(handler)

		ctx.SetParamNames("id")
		ctx.SetParamValues("2ceba62a-35f4-444b-a358-4b14834837e1")

		userUC.EXPECT().UpdateById(gomock.Any(), gomock.Any()).AnyTimes().Return(&models.User{UserID: userUUID}, nil)

		require.NoError(t, h(ctx))
		require.Equal(t, http.StatusForbidden, res.Code)
	})

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		res := httptest.NewRecorder()
		ctx := e.NewContext(req, res)

		handler := handlers.UpdateById()
		h := middleware.JWTWithConfig(middleware.JWTConfig{
			Claims:     claims,
			SigningKey: []byte("secret"),
		})(handler)

		ctx.SetParamNames("id")
		ctx.SetParamValues(userUUID.String())

		userUC.EXPECT().UpdateById(gomock.Any(), gomock.Any()).AnyTimes().Return(&models.User{UserID: userUUID}, nil)
		userUC.EXPECT().FindById(gomock.Any(), userUUID).AnyTimes().Return(&models.User{UserID: userUUID}, nil)

		require.NoError(t, h(ctx))
		require.Equal(t, http.StatusOK, res.Code)
	})
}

func TestUsersService_DeleteById(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mock.NewMockUserUseCase(ctrl)
	sessUC := mockSessUC.NewMockSessUseCase(ctrl)

	cfg := &config.Config{Session: config.Session{Expire: 1234}}
	appLogger := logger.NewAppLogger(cfg)
	mw := middlewares.NewMiddlewareManager(appLogger, nil)

	e := echo.New()
	v := validator.New()
	handlers := NewUserHandlersHTTP(e.Group("user"), appLogger, cfg, mw, v, userUC, sessUC)

	req := httptest.NewRequest(http.MethodDelete, "/user/:id", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	res := httptest.NewRecorder()
	ctx := e.NewContext(req, res)

	userUUID := uuid.New()
	ctx.SetParamNames("id")
	ctx.SetParamValues(userUUID.String())

	userUC.EXPECT().DeleteById(gomock.Any(), userUUID).AnyTimes().Return(nil)
	require.NoError(t, handlers.DeleteById()(ctx))
	require.Equal(t, http.StatusOK, res.Code)
}

func TestUsersService_GetMe(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mock.NewMockUserUseCase(ctrl)
	sessUC := mockSessUC.NewMockSessUseCase(ctrl)

	cfg := &config.Config{Session: config.Session{Expire: 1234}}
	appLogger := logger.NewAppLogger(cfg)
	mw := middlewares.NewMiddlewareManager(appLogger, cfg)

	e := echo.New()
	e.Use(middleware.JWT([]byte("secret")))
	v := validator.New()
	handlers := NewUserHandlersHTTP(e.Group("user"), appLogger, cfg, mw, v, userUC, sessUC)

	userUUID := uuid.New()
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["session_id"] = uuid.New().String()
	claims["user_id"] = userUUID.String()
	claims["role"] = "user"
	claims["exp"] = time.Now().Add(time.Minute * 15).Unix()
	validToken, _ := token.SignedString([]byte("secret"))

	req := httptest.NewRequest(http.MethodPost, "/user/logout", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, fmt.Sprintf("bearer %v", validToken))

	res := httptest.NewRecorder()
	ctx := e.NewContext(req, res)

	handler := handlers.GetMe()
	h := middleware.JWTWithConfig(middleware.JWTConfig{
		Claims:     claims,
		SigningKey: []byte("secret"),
	})(handler)

	sessUC.EXPECT().GetSessionById(gomock.Any(), claims["session_id"].(string)).AnyTimes().Return(&models.Session{}, nil)
	userUC.EXPECT().CachedFindById(gomock.Any(), gomock.Any()).AnyTimes().Return(&models.User{}, nil)

	require.NoError(t, h(ctx))
	require.Equal(t, http.StatusOK, res.Code)
}

func TestUsersService_Logout(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mock.NewMockUserUseCase(ctrl)
	sessUC := mockSessUC.NewMockSessUseCase(ctrl)

	cfg := &config.Config{Session: config.Session{Expire: 1234}}
	appLogger := logger.NewAppLogger(cfg)
	mw := middlewares.NewMiddlewareManager(appLogger, cfg)

	e := echo.New()
	e.Use(middleware.JWT([]byte("secret")))
	v := validator.New()
	handlers := NewUserHandlersHTTP(e.Group("user"), appLogger, cfg, mw, v, userUC, sessUC)

	userUUID := uuid.New()
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["session_id"] = uuid.New().String()
	claims["user_id"] = userUUID.String()
	claims["role"] = "user"
	claims["exp"] = time.Now().Add(time.Minute * 15).Unix()
	validToken, _ := token.SignedString([]byte("secret"))

	req := httptest.NewRequest(http.MethodPost, "/user/logout", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, fmt.Sprintf("bearer %v", validToken))

	res := httptest.NewRecorder()
	ctx := e.NewContext(req, res)

	handler := handlers.Logout()
	h := middleware.JWTWithConfig(middleware.JWTConfig{
		Claims:     claims,
		SigningKey: []byte("secret"),
	})(handler)

	sessUC.EXPECT().DeleteById(gomock.Any(), claims["session_id"].(string)).AnyTimes().Return(nil)

	require.NoError(t, h(ctx))
	require.Equal(t, http.StatusOK, res.Code)
}

func TestUsersService_RefreshToken(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userUC := mock.NewMockUserUseCase(ctrl)
	sessUC := mockSessUC.NewMockSessUseCase(ctrl)

	cfg := &config.Config{Session: config.Session{Expire: 1234}, Server: config.ServerConfig{JwtSecretKey: "secret"}}
	appLogger := logger.NewAppLogger(cfg)
	mw := middlewares.NewMiddlewareManager(appLogger, cfg)

	e := echo.New()
	v := validator.New()
	handlers := NewUserHandlersHTTP(e.Group("user"), appLogger, cfg, mw, v, userUC, sessUC)

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["session_id"] = uuid.New().String()
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	validToken, _ := token.SignedString([]byte("secret"))

	reqDto := &dto.UserRefreshTokenDto{
		RefreshToken: validToken,
	}

	buf := &bytes.Buffer{}
	_ = json.NewEncoder(buf).Encode(reqDto)

	req := httptest.NewRequest(http.MethodPost, "/user/refresh", buf)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	res := httptest.NewRecorder()
	ctx := e.NewContext(req, res)

	sessUC.EXPECT().GetSessionById(gomock.Any(), claims["session_id"].(string)).AnyTimes().Return(&models.Session{}, nil)
	userUC.EXPECT().FindById(gomock.Any(), gomock.Any()).AnyTimes().Return(&models.User{}, nil)
	userUC.EXPECT().GenerateTokenPair(gomock.Any(), gomock.Any()).AnyTimes().Return("rt", "at", nil)

	require.NoError(t, handlers.RefreshToken()(ctx))
	require.Equal(t, http.StatusOK, res.Code)
}
