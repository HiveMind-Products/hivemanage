package publicapi

import (
	"os"
	"strings"

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
	// The Swagger UI/spec exposes the full API surface; only serve it in dev so it
	// is not world-readable in production. Set SWAGGER_ENABLED=true to force-enable.
	if strings.EqualFold(os.Getenv("ENV"), "dev") || strings.EqualFold(os.Getenv("SWAGGER_ENABLED"), "true") {
		group.GET("/swagger/*", echoSwagger.WrapHandler)
	}

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
