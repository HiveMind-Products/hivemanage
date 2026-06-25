package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	internalauth "github.com/fivemanage/lite/internal/auth"
	"github.com/labstack/echo/v4"
)

func TestCSRFRejectsMissingTokenForMutations(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/dash/example", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := CSRF()(func(c echo.Context) error { return c.NoContent(http.StatusOK) })
	if err := h(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestCSRFAcceptsMatchingCookieAndHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/dash/example", nil)
	req.AddCookie(&http.Cookie{Name: internalauth.CSRFCookieName, Value: "token"})
	req.Header.Set(internalauth.CSRFHeaderName, "token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := CSRF()(func(c echo.Context) error { return c.NoContent(http.StatusOK) })
	if err := h(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
