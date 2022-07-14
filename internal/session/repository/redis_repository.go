package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/dinorain/useraja/config"
	"github.com/dinorain/useraja/internal/models"
	"github.com/dinorain/useraja/internal/session"
)

const (
	basePrefix = "sessions:"
)

// Session repository
type sessionRepo struct {
	redisClient *redis.Client
	basePrefix  string
	cfg         *config.Config
}

var _ session.SessRepository = (*sessionRepo)(nil)

// Session repository constructor
func NewSessionRepository(redisClient *redis.Client, cfg *config.Config) session.SessRepository {
	return &sessionRepo{redisClient: redisClient, basePrefix: basePrefix, cfg: cfg}
}

// Create session in redis
func (s *sessionRepo) CreateSession(ctx context.Context, sess *models.Session, expire int) (string, error) {
	sess.SessionID = uuid.New().String()
	sessionKey := s.generateKey(sess.SessionID)

	sessBytes, err := json.Marshal(&sess)
	if err != nil {
		return "", errors.WithMessage(err, "sessionRepo.CreateSession.json.Marshal")
	}
	if err = s.redisClient.Set(ctx, sessionKey, sessBytes, time.Second*time.Duration(expire)).Err(); err != nil {
		return "", errors.Wrap(err, "sessionRepo.CreateSession.redisClient.Set")
	}
	return sess.SessionID, nil
}

// Get session by id
func (s *sessionRepo) GetSessionById(ctx context.Context, sessionID string) (*models.Session, error) {
	sessBytes, err := s.redisClient.Get(ctx, s.generateKey(sessionID)).Bytes()
	if err != nil {
		return nil, errors.Wrap(err, "sessionRep.GetSessionById.redisClient.Get")
	}

	sess := &models.Session{}
	if err = json.Unmarshal(sessBytes, &sess); err != nil {
		return nil, errors.Wrap(err, "sessionRepo.GetSessionById.json.Unmarshal")
	}
	return sess, nil
}

// Delete session by id
func (s *sessionRepo) DeleteById(ctx context.Context, sessionID string) error {
	if err := s.redisClient.Del(ctx, s.generateKey(sessionID)).Err(); err != nil {
		return errors.Wrap(err, "sessionRepo.DeleteById")
	}
	return nil
}

func (s *sessionRepo) generateKey(sessionID string) string {
	return fmt.Sprintf("%s: %s", s.basePrefix, sessionID)
}
