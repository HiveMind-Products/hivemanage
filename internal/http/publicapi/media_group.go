package publicapi

import (
	"net/http"
	"os"

	"github.com/fivemanage/lite/internal/http/httputil"
	"github.com/fivemanage/lite/internal/http/middleware"
	"github.com/fivemanage/lite/internal/service/file"
	"github.com/fivemanage/lite/internal/service/token"
	"github.com/fivemanage/lite/pkg/cache"
	"github.com/labstack/echo/v4"
)

// DRY they said
func registerMediaApi(group *echo.Group, fileService *file.Service, tokenService *token.Service, cache *cache.Cache) {
	group.POST("/image", func(c echo.Context) error {
		var err error
		ctx := c.Request().Context()

		file, header, err := httputil.File(c.Request(), "image")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error": err.Error(),
			})
		}

		key, err := fileService.CreateFile(ctx, "", file, header)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}

		bucket := os.Getenv("BUCKET_DOMAIN")
		url := bucket + "/" + key

		return c.JSON(200, echo.Map{"url": url})
	}, middleware.TokenAuth(tokenService, cache))

	group.POST("/video", func(c echo.Context) error {
		var err error
		ctx := c.Request().Context()

		file, header, err := httputil.File(c.Request(), "video")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error": err.Error(),
			})
		}

		key, err := fileService.CreateFile(ctx, "", file, header)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}

		bucket := os.Getenv("BUCKET_DOMAIN")
		url := bucket + "/" + key

		return c.JSON(200, echo.Map{"url": url})
	})

	group.POST("/audio", func(c echo.Context) error {
		var err error
		ctx := c.Request().Context()

		file, header, err := httputil.File(c.Request(), "audio")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error": err.Error(),
			})
		}

		key, err := fileService.CreateFile(ctx, "", file, header)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err)
		}

		bucket := os.Getenv("BUCKET_DOMAIN")
		url := bucket + "/" + key

		return c.JSON(200, echo.Map{"url": url})
	})
}
