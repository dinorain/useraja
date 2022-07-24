package service

import (
	"context"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/require"

	"github.com/dinorain/useraja/config"
	"github.com/dinorain/useraja/internal/models"
	mockSessUC "github.com/dinorain/useraja/internal/session/mock"
	"github.com/dinorain/useraja/internal/user/mock"
	"github.com/dinorain/useraja/pkg/logger"
	userService "github.com/dinorain/useraja/proto"
)

func TestUsersService_Register(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userUC := mock.NewMockUserUseCase(ctrl)
	sessUC := mockSessUC.NewMockSessUseCase(ctrl)
	apiLogger := logger.NewAppLogger(nil)
	authServerGRPC := NewAuthServerGRPC(apiLogger, nil, userUC, sessUC)

	reqValue := &userService.RegisterRequest{
		Email:     "email@gmail.com",
		FirstName: "FirstName",
		LastName:  "LastName",
		Password:  "Password",
		Role:      "user",
		Avatar:    "",
	}

	t.Run("Register", func(t *testing.T) {
		t.Parallel()
		userID := uuid.New()
		user := &models.User{
			UserID:    userID,
			Email:     reqValue.Email,
			FirstName: reqValue.FirstName,
			LastName:  reqValue.LastName,
			Password:  reqValue.Password,
			Role:      reqValue.Role,
			Avatar:    nil,
		}

		userUC.EXPECT().Register(gomock.Any(), gomock.Any()).Return(user, nil)

		response, err := authServerGRPC.Register(context.Background(), reqValue)
		require.NoError(t, err)
		require.NotNil(t, response)
		require.Equal(t, reqValue.Email, response.User.Email)
	})
}

func TestUsersService_Login(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userUC := mock.NewMockUserUseCase(ctrl)
	sessUC := mockSessUC.NewMockSessUseCase(ctrl)
	apiLogger := logger.NewAppLogger(nil)
	cfg := &config.Config{Session: config.Session{
		Expire: 10,
	}}
	authServerGRPC := NewAuthServerGRPC(apiLogger, cfg, userUC, sessUC)

	reqValue := &userService.LoginRequest{
		Email:    "email@gmail.com",
		Password: "Password",
	}

	t.Run("Login", func(t *testing.T) {
		t.Parallel()
		userID := uuid.New()
		session := "session"
		user := &models.User{
			UserID:    userID,
			Email:     "email@gmail.com",
			FirstName: "FirstName",
			LastName:  "LastName",
			Password:  "Password",
			Role:      "user",
			Avatar:    nil,
		}

		userUC.EXPECT().Login(gomock.Any(), reqValue.Email, reqValue.Password).Return(user, nil)
		sessUC.EXPECT().CreateSession(gomock.Any(), &models.Session{
			UserID: user.UserID,
		}, cfg.Session.Expire).Return(session, nil)

		response, err := authServerGRPC.Login(context.Background(), reqValue)
		require.NoError(t, err)
		require.NotNil(t, response)
		require.Equal(t, reqValue.Email, response.User.Email)
	})
}

func TestUsersService_FindByID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userUC := mock.NewMockUserUseCase(ctrl)
	sessUC := mockSessUC.NewMockSessUseCase(ctrl)
	apiLogger := logger.NewAppLogger(nil)
	cfg := &config.Config{Session: config.Session{
		Expire: 10,
	}}
	authServerGRPC := NewAuthServerGRPC(apiLogger, cfg, userUC, sessUC)

	userUUID := uuid.New()
	reqValue := &userService.FindByIdRequest{
		Uuid: userUUID.String(),
	}

	t.Run("FindByID", func(t *testing.T) {
		t.Parallel()
		user := &models.User{
			UserID:    userUUID,
			Email:     "email@gmail.com",
			FirstName: "FirstName",
			LastName:  "LastName",
			Password:  "Password",
			Role:      "user",
			Avatar:    nil,
		}

		userUC.EXPECT().CachedFindById(gomock.Any(), user.UserID).Return(user, nil)

		response, err := authServerGRPC.FindById(context.Background(), reqValue)
		require.NoError(t, err)
		require.NotNil(t, response)
		require.Equal(t, reqValue.Uuid, response.User.Uuid)
	})
}

func TestUsersService_FindByEmail(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userUC := mock.NewMockUserUseCase(ctrl)
	sessUC := mockSessUC.NewMockSessUseCase(ctrl)
	apiLogger := logger.NewAppLogger(nil)
	cfg := &config.Config{Session: config.Session{
		Expire: 10,
	}}
	authServerGRPC := NewAuthServerGRPC(apiLogger, cfg, userUC, sessUC)

	reqValue := &userService.FindByEmailRequest{
		Email: "email@gmail.com",
	}

	t.Run("FindByEmail", func(t *testing.T) {
		t.Parallel()
		userID := uuid.New()
		user := &models.User{
			UserID:    userID,
			Email:     "email@gmail.com",
			FirstName: "FirstName",
			LastName:  "LastName",
			Password:  "Password",
			Role:      "user",
			Avatar:    nil,
		}

		userUC.EXPECT().FindByEmail(gomock.Any(), reqValue.Email).Return(user, nil)

		response, err := authServerGRPC.FindByEmail(context.Background(), reqValue)
		require.NoError(t, err)
		require.NotNil(t, response)
		require.Equal(t, reqValue.Email, response.User.Email)
	})
}

func TestUsersService_GetMe(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userUC := mock.NewMockUserUseCase(ctrl)
	sessUC := mockSessUC.NewMockSessUseCase(ctrl)
	apiLogger := logger.NewAppLogger(nil)
	cfg := &config.Config{Session: config.Session{
		Expire: 10,
	}}
	authServerGRPC := NewAuthServerGRPC(apiLogger, cfg, userUC, sessUC)

	userUUID := uuid.New()
	sessionUUID := uuid.New().String()
	reqValue := &userService.GetMeRequest{}

	t.Run("GetMe", func(t *testing.T) {
		t.Parallel()
		user := &models.User{
			UserID:    userUUID,
			Email:     "email@gmail.com",
			FirstName: "FirstName",
			LastName:  "LastName",
			Password:  "Password",
			Role:      "user",
			Avatar:    nil,
		}

		sessUC.EXPECT().GetSessionById(gomock.Any(), sessionUUID).Return(&models.Session{SessionID: sessionUUID}, nil)
		userUC.EXPECT().CachedFindById(gomock.Any(), gomock.Any()).Return(user, nil)

		ctx := context.Background()
		expectedRPCName := "/example.Example/Example"
		expectedHTTPPathPattern := "/v1"
		request, _ := http.NewRequest("GET", "http://www.example.com/v1", nil)
		request.Header.Add("Grpc-Metadata-session_id", sessionUUID)
		annotated, err := runtime.AnnotateIncomingContext(ctx, runtime.NewServeMux(), request, expectedRPCName, runtime.WithHTTPPathPattern(expectedHTTPPathPattern))
		if err != nil {
			t.Errorf("runtime.AnnotateIncomingContext(ctx, %#v) failed with %v; want success", request, err)
			return
		}

		response, err := authServerGRPC.GetMe(annotated, reqValue)
		require.NoError(t, err)
		require.NotNil(t, response)
		require.Equal(t, response.User.Uuid, userUUID.String())
	})
}

func TestUsersService_Logout(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userUC := mock.NewMockUserUseCase(ctrl)
	sessUC := mockSessUC.NewMockSessUseCase(ctrl)
	apiLogger := logger.NewAppLogger(nil)
	cfg := &config.Config{Session: config.Session{
		Expire: 10,
	}}
	authServerGRPC := NewAuthServerGRPC(apiLogger, cfg, userUC, sessUC)

	sessionUUID := uuid.New().String()
	reqValue := &userService.LogoutRequest{}

	t.Run("Logout", func(t *testing.T) {
		t.Parallel()

		sessUC.EXPECT().GetSessionById(gomock.Any(), sessionUUID).Return(&models.Session{SessionID: sessionUUID}, nil)
		sessUC.EXPECT().DeleteById(gomock.Any(), sessionUUID).Return(nil)

		ctx := context.Background()
		expectedRPCName := "/example.Example/Example"
		expectedHTTPPathPattern := "/v1"
		request, _ := http.NewRequest("GET", "http://www.example.com/v1", nil)
		request.Header.Add("Grpc-Metadata-session_id", sessionUUID)
		annotated, err := runtime.AnnotateIncomingContext(ctx, runtime.NewServeMux(), request, expectedRPCName, runtime.WithHTTPPathPattern(expectedHTTPPathPattern))
		if err != nil {
			t.Errorf("runtime.AnnotateIncomingContext(ctx, %#v) failed with %v; want success", request, err)
			return
		}

		response, err := authServerGRPC.Logout(annotated, reqValue)
		require.NoError(t, err)
		require.NotNil(t, response)
	})
}
