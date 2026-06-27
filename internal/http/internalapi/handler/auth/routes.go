package auth

import (
	"github.com/fivemanage/lite/internal/service/auth"
	"github.com/labstack/echo/v4"
)

type handler struct {
	authService *auth.Service
}

func RegisterRoutes(group *echo.Group, authService *auth.Service) {
	handler := handler{
		authService: authService,
	}

	group.GET("/auth/session", handler.getSessionHandler)
	group.POST("/auth/login", handler.loginHandler)
	group.POST("/auth/logout", handler.logoutHandler)
	group.GET("/auth/discord", handler.discordLoginHandler)
	group.GET("/auth/discord/callback", handler.discordCallbackHandler)
}
