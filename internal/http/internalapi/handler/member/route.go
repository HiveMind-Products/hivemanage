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
	group.GET("/organization/:organizationId/member", h.listMembersHandler, middleware.OrganizationAdmin(authService))
}
