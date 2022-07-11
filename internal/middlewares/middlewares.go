package middlewares

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/dinorain/useraja/config"
	"github.com/dinorain/useraja/internal/models"
	"github.com/dinorain/useraja/pkg/logger"
)

type MiddlewareManager interface {
	RequestLoggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc
	IsLoggedIn() echo.MiddlewareFunc
	IsAdmin(next echo.HandlerFunc) echo.HandlerFunc
}

type middlewareManager struct {
	logger logger.Logger
	cfg    *config.Config
}

var _ MiddlewareManager = (*middlewareManager)(nil)

func NewMiddlewareManager(logger logger.Logger, cfg *config.Config) *middlewareManager {
	return &middlewareManager{logger: logger, cfg: cfg}
}

func (mw *middlewareManager) IsLoggedIn() echo.MiddlewareFunc {
	return middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(mw.cfg.Server.JwtSecretKey),
	})
}

func (mw *middlewareManager) IsAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := c.Get("user").(*jwt.Token)
		if !ok {
			mw.logger.Warnf("jwt.Token: %+v", c.Get("user"))
			return errors.New("invalid token header")
		}
		claims := user.Claims.(jwt.MapClaims)
		if !ok {
			mw.logger.Warnf("jwt.MapClaims: %+v", c.Get("user"))
			return errors.New("invalid token header")
		}
		isAdmin, ok := claims["admin"].(string)
		if !ok {
			mw.logger.Warnf("isAdmin: %v", claims)
			return errors.New("invalid token header")
		}

		if isAdmin != models.UserRoleAdmin {
			return echo.ErrUnauthorized
		}

		return next(c)
	}
}

func (mw *middlewareManager) RequestLoggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {

		start := time.Now()
		err := next(ctx)

		req := ctx.Request()
		res := ctx.Response()
		status := res.Status
		size := res.Size
		s := time.Since(start)

		if !mw.checkIgnoredURI(ctx.Request().RequestURI, mw.cfg.Http.IgnoreLogUrls) {
			mw.logger.HttpMiddlewareAccessLogger(req.Method, req.URL.String(), status, size, s)
		}

		return err
	}
}

func (mw *middlewareManager) checkIgnoredURI(requestURI string, uriList []string) bool {
	for _, s := range uriList {
		if strings.Contains(requestURI, s) {
			return true
		}
	}
	return false
}
