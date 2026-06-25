package token

import (
	"github.com/fivemanage/lite/internal/http/middleware"
	"github.com/fivemanage/lite/internal/service/auth"
	"github.com/fivemanage/lite/internal/service/token"
	"github.com/labstack/echo/v4"
)

type handler struct{ tokenService *token.Service }

func RegisterRoutes(group *echo.Group, tokenService *token.Service, authService *auth.Service) {
	handler := handler{tokenService: tokenService}
	adminOnly := middleware.OrganizationAdmin(authService)
	group.POST("/:organizationId/token", handler.createTokenHandler, adminOnly)
	group.GET("/:organizationId/token", handler.listTokensHandler)
	group.DELETE("/:organizationId/token/:id", handler.deleteTokenHandler, adminOnly)
}
