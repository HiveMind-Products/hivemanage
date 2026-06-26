package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	internalauth "github.com/fivemanage/lite/internal/auth"
	"github.com/fivemanage/lite/internal/service/token"
	"github.com/labstack/echo/v4"
)

func TokenAuth(tokenService *token.Service) echo.MiddlewareFunc {
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
				slog.Error("failed to get token from service", "err", err)
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized: Invalid token"})
			}

			c.Set(internalauth.OrgIDContextKey, tokenData.OrganizationID)
			c.Set("token_data", tokenData)
			return next(c)
		}
	}
}
