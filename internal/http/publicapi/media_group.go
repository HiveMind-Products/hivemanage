package publicapi

import (
	"net/http"

	"github.com/fivemanage/lite/internal/auth"
	"github.com/fivemanage/lite/internal/http/httputil"
	"github.com/fivemanage/lite/internal/http/middleware"
	"github.com/fivemanage/lite/internal/service/file"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

// registerMediaApi registers the legacy V2 media upload routes.
//
// Deprecated: these routes are superseded by POST /api/v3/file (field "file").
// They are retained for backward compatibility and reuse the V3 upload logic
// internally, but keep their original flat { "url": ... } response shape.
func registerMediaApi(group *echo.Group, fileService *file.Service) {
	handle := func(fileType string) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Deprecation", "true")
			ctx := c.Request().Context()
			orgId, err := auth.CurrentOrgId(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{"error": "Unauthorized: " + err.Error()})
			}
			f, header, err := httputil.File(c.Request(), fileType)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
			}
			defer f.Close()
			item, err := fileService.CreateFileV3(ctx, orgId, f, header, file.CreateFileV3Params{})
			if err != nil {
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
			}
			return c.JSON(http.StatusOK, echo.Map{"url": item.URL})
		}
	}
	group.POST("/image", handle("image"), echoMiddleware.BodyLimit("500M"), middleware.ValidateMime("image", middleware.WhitelistedImageMIME))
	group.POST("/video", handle("video"), echoMiddleware.BodyLimit("500M"), middleware.ValidateMime("video", middleware.WhitelistedVideoMIME))
	group.POST("/audio", handle("audio"), echoMiddleware.BodyLimit("500M"), middleware.ValidateMime("audio", middleware.WhitelistedAudioMIME))
	group.POST("/file", handle("file"), echoMiddleware.BodyLimit("500M"))
}
