package member

import (
	"github.com/fivemanage/lite/api"
	"github.com/fivemanage/lite/internal/http/middleware"
	"github.com/fivemanage/lite/internal/service/auth"
	inviteservice "github.com/fivemanage/lite/internal/service/invite"
	"github.com/fivemanage/lite/internal/service/member"
	"github.com/labstack/echo/v4"
)

type handler struct {
	memberService *member.Service
	inviteService *inviteservice.Service
	authService   *auth.Service
}

func RegisterRoutes(group *echo.Group, memberService *member.Service, authService *auth.Service, inviteService *inviteservice.Service) {
	h := handler{memberService: memberService, inviteService: inviteService, authService: authService}
	readTeam := middleware.OrganizationPermission(authService, api.PermissionModuleTeam, api.PermissionActionRead)
	writeTeam := middleware.OrganizationPermission(authService, api.PermissionModuleTeam, api.PermissionActionWrite)
	group.GET("/organization/:organizationId/member", h.listMembersHandler, readTeam)
	group.POST("/organization/:organizationId/member", h.addMemberHandler, writeTeam)
	group.PATCH("/organization/:organizationId/member/:memberId", h.updateMemberHandler, writeTeam)
	group.DELETE("/organization/:organizationId/member/:memberId", h.removeMemberHandler, writeTeam)
	group.POST("/organization/:organizationId/invite", h.createInviteHandler, writeTeam)
	group.GET("/organization/:organizationId/invite", h.listInvitesHandler, readTeam)
	group.DELETE("/organization/:organizationId/invite/:inviteId", h.deleteInviteHandler, writeTeam)
}
