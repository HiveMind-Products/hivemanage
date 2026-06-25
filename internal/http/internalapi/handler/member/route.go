package member

import (
	"github.com/fivemanage/lite/api"
	"github.com/fivemanage/lite/internal/http/middleware"
	"github.com/fivemanage/lite/internal/service/auth"
	"github.com/fivemanage/lite/internal/service/member"
	"github.com/labstack/echo/v4"
)

type handler struct{ memberService *member.Service }

func RegisterRoutes(group *echo.Group, memberService *member.Service, authService *auth.Service) {
	h := handler{memberService: memberService}
	readTeam := middleware.OrganizationPermission(authService, api.PermissionModuleTeam, api.PermissionActionRead)
	writeTeam := middleware.OrganizationPermission(authService, api.PermissionModuleTeam, api.PermissionActionWrite)
	group.GET("/organization/:organizationId/member", h.listMembersHandler, readTeam)
	group.POST("/organization/:organizationId/member", h.addMemberHandler, writeTeam)
	group.PATCH("/organization/:organizationId/member/:memberId", h.updateMemberHandler, writeTeam)
	group.DELETE("/organization/:organizationId/member/:memberId", h.removeMemberHandler, writeTeam)
}
