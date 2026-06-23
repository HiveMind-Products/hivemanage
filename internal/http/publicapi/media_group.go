package publicapi

import (
	"net/http"
	"os"
	"strings"

	"github.com/fivemanage/lite/internal/auth"
	"github.com/fivemanage/lite/internal/http/httputil"
	"github.com/fivemanage/lite/internal/http/middleware"
	"github.com/fivemanage/lite/internal/service/file"
	"github.com/fivemanage/lite/internal/service/token"
	"github.com/fivemanage/lite/pkg/cache"
	"github.com/labstack/echo/v4"
)

func registerMediaApi(group *echo.Group, fileService *file.Service, tokenService *token.Service, cache *cache.Cache) {
	handle := func(fileType string) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			orgId, err := auth.CurrentOrgId(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{"error": "Unauthorized: " + err.Error()})
			}

			f, header, err := httputil.File(c.Request(), fileType)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
			}

			key, err := fileService.CreateFile(ctx, orgId, f, header)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
			}

			bucket := strings.TrimRight(os.Getenv("BUCKET_DOMAIN"), "/")
			return c.JSON(http.StatusOK, echo.Map{"url": bucket + "/" + key})
		}
	}

	group.POST("/image", handle("image"),
		middleware.TokenAuth(tokenService, cache),
		middleware.ValidateMime("image", middleware.WhitelistedImageMIME))

	group.POST("/video", handle("video"),
		middleware.TokenAuth(tokenService, cache),
		middleware.ValidateMime("video", middleware.WhitelistedVideoMIME))

	group.POST("/audio", handle("audio"),
		middleware.TokenAuth(tokenService, cache),
		middleware.ValidateMime("audio", middleware.WhitelistedAudioMIME))

	group.POST("/file", handle("file"),
		middleware.TokenAuth(tokenService, cache),
		middleware.ValidateMime("file", nil))
}
