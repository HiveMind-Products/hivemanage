package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	internalauth "github.com/fivemanage/lite/internal/auth"
	"github.com/fivemanage/lite/internal/database"
	"github.com/fivemanage/lite/internal/permissions"
	"github.com/fivemanage/lite/internal/service/token"
	"github.com/labstack/echo/v4"
)

// tokenDataContextKey is where TokenAuth stashes the resolved token so scope
// checks downstream can read it.
const tokenDataContextKey = "token_data"

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
				if errors.Is(err, token.ErrTokenExpired) {
					return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized: token expired"})
				}
				slog.Error("failed to get token from service", "err", err)
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized: Invalid token"})
			}

			c.Set(internalauth.OrgIDContextKey, tokenData.OrganizationID)
			c.Set(tokenDataContextKey, tokenData)
			return next(c)
		}
	}
}

// RequireTokenScope gates a route on an API token holding the given
// module:action scope. It must run after TokenAuth. A token with no scopes has
// full access (legacy behaviour), so this only restricts scoped tokens.
func RequireTokenScope(module, action string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenData, ok := c.Get(tokenDataContextKey).(*database.Token)
			if !ok || tokenData == nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
			}
			if !permissions.TokenAllows(tokenData.Scopes, module, action) {
				return c.JSON(http.StatusForbidden, echo.Map{
					"error": "token missing required scope: " + permissions.ScopeString(module, action),
				})
			}
			return next(c)
		}
	}
}
