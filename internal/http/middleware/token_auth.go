package middleware

import (
	"net/http"
	"strings"

	internalauth "github.com/fivemanage/lite/internal/auth"
	"github.com/fivemanage/lite/internal/service/token"
	"github.com/fivemanage/lite/pkg/cache"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func TokenAuth(tokenService *token.Service, _ *cache.Cache) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
			apiToken := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			if apiToken == "" {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized: No token provided"})
			}

			tokenData, err := tokenService.GetToken(ctx, apiToken)
			if err != nil {
				logrus.WithField("error", err).Error("failed to get token from service")
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized: Invalid token"})
			}

			c.Set(internalauth.OrgIDContextKey, tokenData.OrganizationID)
			c.Set("token_data", tokenData)
			return next(c)
		}
	}
}
