package publicapi

import (
	"github.com/fivemanage/lite/internal/service/file"
	"github.com/fivemanage/lite/internal/service/log"
	"github.com/fivemanage/lite/internal/service/token"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "github.com/fivemanage/lite/docs"
)

func Add(group *echo.Group,
	fileService *file.Service,
	tokenService *token.Service,
	logService *log.Service,
) {
	group.GET("/swagger/*", echoSwagger.WrapHandler)

	registerMediaApi(group, fileService, tokenService)
	registerLogsApi(group, logService, tokenService)
}
