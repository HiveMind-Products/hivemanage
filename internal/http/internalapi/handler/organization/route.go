package organization

import (
	"github.com/fivemanage/lite/api"
	"github.com/fivemanage/lite/internal/http/middleware"
	"github.com/fivemanage/lite/internal/service/auth"
	"github.com/fivemanage/lite/internal/service/organization"
	"github.com/labstack/echo/v4"
)

type handler struct{ organizationService *organization.Service }

func RegisterRoutes(group *echo.Group, organizationService *organization.Service, authService *auth.Service) {
	h := handler{organizationService: organizationService}
	readOverview := middleware.OrganizationPermission(authService, api.PermissionModuleOverview, api.PermissionActionRead)
	adminOnly := middleware.OrganizationAdmin(authService)
	group.POST("/organization", h.createOrganizationHandler)
	group.GET("/organization", h.listOrganizationsHandler)
	// Renaming and permanently deleting an organization are owner-level,
	// destructive actions — gate them on the admin role, not the delegable
	// team:write permission.
	group.GET("/organization/:organizationId", h.getOrganizationHandler, readOverview)
	group.PATCH("/organization/:organizationId", h.updateOrganizationHandler, adminOnly)
	group.DELETE("/organization/:organizationId", h.deleteOrganizationHandler, adminOnly)
	group.GET("/organization/:organizationId/stats", h.getOrganizationStatsHandler, readOverview)
	group.GET("/organization/:organizationId/usage", h.getOrganizationUsageHandler, readOverview)
}
