package middleware

import (
	"net/http"

	"github.com/fivemanage/lite/internal/http/appctx"
	"github.com/fivemanage/lite/internal/service/auth"
	"github.com/labstack/echo/v4"
)

func OrganizationPermission(authService *auth.Service, module string, action string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := c.(*appctx.Context)
			user := cc.User()
			if user == nil {
				return cc.JSON(http.StatusUnauthorized, echo.Map{"error": "session required"})
			}

			orgID, err := resolveOrgID(cc)
			if err != nil {
				return cc.JSON(http.StatusForbidden, echo.Map{"error": err.Error()})
			}

			allowed, err := authService.HasOrganizationPermission(cc.Request().Context(), user.ID, orgID, module, action)
			if err != nil {
				return cc.JSON(http.StatusInternalServerError, echo.Map{"error": "authorization check failed"})
			}
			if !allowed {
				return cc.JSON(http.StatusForbidden, echo.Map{"error": "permission denied"})
			}

			return next(cc)
		}
	}
}
