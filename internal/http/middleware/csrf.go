package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"

	internalauth "github.com/fivemanage/lite/internal/auth"
	"github.com/fivemanage/lite/internal/crypt"
	"github.com/labstack/echo/v4"
)

func CSRF() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			method := c.Request().Method
			if method != http.MethodPost && method != http.MethodPut && method != http.MethodPatch && method != http.MethodDelete {
				return next(c)
			}

			path := c.Request().URL.Path
			if path == "/api/dash/auth/login" || path == "/api/dash/auth/register" {
				return next(c)
			}

			cookie, err := c.Cookie(internalauth.CSRFCookieName)
			if err != nil || cookie.Value == "" {
				return c.JSON(http.StatusForbidden, echo.Map{"error": "csrf token required"})
			}

			header := c.Request().Header.Get(internalauth.CSRFHeaderName)
			if header == "" || subtle.ConstantTimeCompare([]byte(header), []byte(cookie.Value)) != 1 {
				return c.JSON(http.StatusForbidden, echo.Map{"error": "invalid csrf token"})
			}

			// Signed double-submit: when a session is present the CSRF token must be
			// the HMAC of the session ID, so a planted or forged CSRF cookie cannot
			// be made valid without knowledge of the HttpOnly session ID.
			if sessionCookie, err := c.Cookie(internalauth.SessionCookieName); err == nil && sessionCookie.Value != "" {
				expected := crypt.ComputeHMAC(os.Getenv("API_TOKEN_HMAC_SECRET"), "csrf:"+sessionCookie.Value)
				if subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(expected)) != 1 {
					return c.JSON(http.StatusForbidden, echo.Map{"error": "invalid csrf token"})
				}
			}

			return next(c)
		}
	}
}
