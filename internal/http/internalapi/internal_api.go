package internalapi

import (
	authhander "github.com/fivemanage/lite/internal/http/internalapi/handler/auth"
	datasethandler "github.com/fivemanage/lite/internal/http/internalapi/handler/dataset"
	memberhandler "github.com/fivemanage/lite/internal/http/internalapi/handler/member"
	organizationhandler "github.com/fivemanage/lite/internal/http/internalapi/handler/organization"
	tokenhandler "github.com/fivemanage/lite/internal/http/internalapi/handler/token"
	"github.com/fivemanage/lite/internal/http/middleware"
	"github.com/fivemanage/lite/internal/service/auth"
	"github.com/fivemanage/lite/internal/service/dataset"
	"github.com/fivemanage/lite/internal/service/file"
	"github.com/fivemanage/lite/internal/service/invite"
	"github.com/fivemanage/lite/internal/service/member"
	"github.com/fivemanage/lite/internal/service/organization"
	"github.com/fivemanage/lite/internal/service/system"
	"github.com/fivemanage/lite/internal/service/token"
	"github.com/labstack/echo/v4"
)

func Add(
	group *echo.Group,
	authService *auth.Service,
	tokenService *token.Service,
	organizationService *organization.Service,
	memberService *member.Service,
	inviteService *invite.Service,
	fileService *file.Service,
	datasetService *dataset.Service,
	systemService *system.Service,
) {
	group.Use(middleware.Session(authService))
	group.Use(middleware.CSRF())

	authhander.RegisterRoutes(group, authService, inviteService)
	tokenhandler.RegisterRoutes(group, tokenService, authService)
	organizationhandler.RegisterRoutes(group, organizationService, authService)
	memberhandler.RegisterRoutes(group, memberService, authService, inviteService)
	datasethandler.RegisterRoutes(group, datasetService, authService)
	registerStorageApi(group, fileService, authService)
	registerSystemApi(group, systemService)
}
