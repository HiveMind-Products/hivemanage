package auth

import (
	"github.com/fivemanage/lite/internal/service/auth"
	"github.com/fivemanage/lite/internal/service/invite"
	"github.com/labstack/echo/v4"
)

type handler struct {
	authService   *auth.Service
	inviteService *invite.Service
}

func RegisterRoutes(group *echo.Group, authService *auth.Service, inviteService *invite.Service) {
	handler := handler{
		authService:   authService,
		inviteService: inviteService,
	}

	group.GET("/auth/session", handler.getSessionHandler)
	group.POST("/auth/login", handler.loginHandler)
	group.POST("/auth/logout", handler.logoutHandler)
	group.GET("/auth/discord", handler.discordLoginHandler)
	group.GET("/auth/discord/link", handler.discordLinkHandler)
	group.GET("/auth/discord/callback", handler.discordCallbackHandler)
}
