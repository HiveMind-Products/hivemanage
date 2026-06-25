package member

import (
	"github.com/fivemanage/lite/internal/http/middleware"
	"github.com/fivemanage/lite/internal/service/auth"
	"github.com/fivemanage/lite/internal/service/member"
	"github.com/labstack/echo/v4"
)

type handler struct{ memberService *member.Service }

func RegisterRoutes(group *echo.Group, memberService *member.Service, authService *auth.Service) {
	h := handler{memberService: memberService}
	adminOnly := middleware.OrganizationAdmin(authService)
	group.GET("/organization/:organizationId/member", h.listMembersHandler, adminOnly)
	group.POST("/organization/:organizationId/member", h.addMemberHandler, adminOnly)
	group.DELETE("/organization/:organizationId/member/:memberId", h.removeMemberHandler, adminOnly)
}
