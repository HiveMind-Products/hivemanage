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
	group.POST("/organization", h.createOrganizationHandler)
	group.GET("/organization", h.listOrganizationsHandler)
	group.GET("/organization/:organizationId", h.getOrganizationHandler)
	group.GET("/organization/:organizationId/stats", h.getOrganizationStatsHandler, readOverview)
}
