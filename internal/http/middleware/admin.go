package middleware

import (
	"net/http"

	internalauth "github.com/fivemanage/lite/internal/auth"
	"github.com/fivemanage/lite/internal/http/appctx"
	"github.com/fivemanage/lite/internal/service/auth"
	"github.com/labstack/echo/v4"
)

func OrganizationAdmin(authService *auth.Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := c.(*appctx.Context)
			user := cc.User()
			if user == nil {
				return cc.JSON(http.StatusUnauthorized, echo.Map{"error": "session required"})
			}

			orgID := cc.Param("organizationId")
			if orgID == "" {
				orgID = cc.Param("id")
			}
			if orgID == "" {
				if contextOrgID, ok := cc.Get(internalauth.OrgIDContextKey).(string); ok {
					orgID = contextOrgID
				}
			}
			if orgID == "" {
				return cc.JSON(http.StatusForbidden, echo.Map{"error": "organization context required"})
			}

			isAdmin, err := authService.IsOrganizationAdmin(cc.Request().Context(), user.ID, orgID)
			if err != nil {
				return cc.JSON(http.StatusInternalServerError, echo.Map{"error": "authorization check failed"})
			}
			if !isAdmin {
				return cc.JSON(http.StatusForbidden, echo.Map{"error": "organization admin required"})
			}

			return next(cc)
		}
	}
}
