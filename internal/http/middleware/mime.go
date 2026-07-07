package middleware

import (
	"fmt"
	"net/http"

	"github.com/fivemanage/lite/internal/http/httputil"
	"github.com/gabriel-vasile/mimetype"
	"github.com/labstack/echo/v4"
)

var WhitelistedImageMIME = []string{
	"image/png",
	"image/apng",
	"image/webp",
	"image/jpeg",
	"image/gif",
}

// Note: application/octet-stream is intentionally NOT whitelisted. mimetype
// returns it for any content it cannot classify, so allowing it would let the
// whitelist be bypassed by arbitrary/unrecognized binaries.
var WhitelistedVideoMIME = []string{
	"video/ogg",
	"video/mp4",
	"video/mpeg",
	"video/webm",
	"video/quicktime",
}

var WhitelistedAudioMIME = []string{
	"audio/mpeg",
	"audio/mp3",
	"audio/ogg",
	"audio/webm",
	"audio/wav",
	"video/webm",
}

func ValidateMime(fileKey string, whitelistedTypes []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			file, _, err := httputil.File(ctx.Request(), fileKey)
			if err != nil {
				return ctx.JSON(http.StatusBadRequest, echo.Map{
					"error": err.Error(),
				})
			}

			mime, err := httputil.DetectMime(file)
			if err != nil {
				return ctx.JSON(http.StatusBadRequest, echo.Map{
					"error": err.Error(),
				})
			}

			if len(whitelistedTypes) == 0 {
				return next(ctx)
			}

			ok := mimetype.EqualsAny(mime.String(), whitelistedTypes...)
			if !ok {
				return ctx.JSON(http.StatusForbidden, echo.Map{
					"error": fmt.Sprintf("Content-Type %s is not allowed", mime.String()),
				})
			}

			return next(ctx)
		}
	}
}
