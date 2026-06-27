package middleware

import (
	"log/slog"
	"net/http"

	internalauth "github.com/fivemanage/lite/internal/auth"
	"github.com/fivemanage/lite/internal/http/appctx"
	"github.com/fivemanage/lite/internal/service/auth"
	"github.com/labstack/echo/v4"
)

func Session(authService *auth.Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := c.(*appctx.Context)
			ctx := cc.Request().Context()
			orgID := cc.Param("organizationId")

			path := cc.Request().URL.Path
			switch path {
			// Public auth routes: these establish a session, so they must not
			// require one. The Discord OAuth start/callback are part of login.
			case "/api/dash/auth/login",
				"/api/dash/auth/register",
				"/api/dash/auth/discord",
				"/api/dash/auth/discord/callback":
				return next(cc)
			}

			sessionCookie, err := cc.Cookie(internalauth.SessionCookieName)
			if err != nil {
				return cc.JSON(http.StatusUnauthorized, map[string]string{"error": "session required"})
			}

			user, err := authService.UserBySession(ctx, sessionCookie.Value)
			if err != nil {
				return cc.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid session"})
			}

			if len(orgID) > 0 {
				isMember, err := authService.IsOrganizationMember(ctx, int64(user.ID), orgID)
				if err != nil {
					slog.Error("failed to check organization membership", "error", err, "userID", user.ID, "orgID", orgID)
					// we need a better error here
					return cc.JSON(http.StatusInternalServerError, map[string]string{"error": "authorization check failed"})
				}

				if !isMember {
					// and here
					return cc.JSON(http.StatusForbidden, map[string]string{"error": "access denied to organization"})
				}
			}

			cc.Set(internalauth.UserContextKey, user)
			cc.Set(internalauth.OrgIDContextKey, orgID)
			return next(cc)
		}
	}
}
