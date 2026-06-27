package publicapi

import (
	"github.com/fivemanage/lite/internal/http/middleware"
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

	// Registered BEFORE TokenAuth: the presigned upload route authenticates via
	// its signed token in the path, not via an API key.
	registerV3PresignedUpload(group, fileService)

	group.Use(middleware.TokenAuth(tokenService))

	registerMediaApi(group, fileService)
	registerLogsApi(group, logService)
	registerV3FileApi(group, fileService)
	registerV3PresignedGenerate(group, fileService)
	registerV3LogsApi(group, logService)
}
