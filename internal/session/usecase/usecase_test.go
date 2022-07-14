package usecase

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/dinorain/useraja/internal/models"
	"github.com/dinorain/useraja/internal/session/mock"
)

func TestSessionUC_CreateSession(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSessRepo := mock.NewMockSessRepository(ctrl)
	sessUC := NewSessionUseCase(mockSessRepo, nil)

	ctx := context.Background()
	sess := &models.Session{}
	sid := "session id"

	mockSessRepo.EXPECT().CreateSession(gomock.Any(), gomock.Eq(sess), 10).Return(sid, nil)

	createdSess, err := sessUC.CreateSession(ctx, sess, 10)
	require.NoError(t, err)
	require.Nil(t, err)
	require.NotEqual(t, createdSess, "")
}

func TestSessionUC_GetSessionById(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSessRepo := mock.NewMockSessRepository(ctrl)
	sessUC := NewSessionUseCase(mockSessRepo, nil)

	ctx := context.Background()
	sess := &models.Session{}
	sid := "session id"

	mockSessRepo.EXPECT().GetSessionById(gomock.Any(), gomock.Eq(sid)).Return(sess, nil)

	session, err := sessUC.GetSessionById(ctx, sid)
	require.NoError(t, err)
	require.Nil(t, err)
	require.NotNil(t, session)
}

func TestSessionUC_DeleteById(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSessRepo := mock.NewMockSessRepository(ctrl)
	sessUC := NewSessionUseCase(mockSessRepo, nil)

	ctx := context.Background()
	sid := "session id"

	mockSessRepo.EXPECT().DeleteById(gomock.Any(), gomock.Eq(sid)).Return(nil)

	err := sessUC.DeleteById(ctx, sid)
	require.NoError(t, err)
	require.Nil(t, err)
}
