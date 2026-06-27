package publicapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	internalauth "github.com/fivemanage/lite/internal/auth"
	"github.com/fivemanage/lite/internal/service/file"
	"github.com/labstack/echo/v4"
)

func TestV3RoutesRegisterWithoutConflict(t *testing.T) {
	// Echo panics on conflicting route shapes; ensure the static (/v3/file/base64,
	// /v3/file/presigned-url) and wildcard (/v3/file/*) routes coexist.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("route registration panicked: %v", r)
		}
	}()
	e := echo.New()
	g := e.Group("/api")
	registerV3PresignedUpload(g, &file.Service{})
	registerV3FileApi(g, &file.Service{})
	registerV3PresignedGenerate(g, &file.Service{})
}

func TestRootIsJSONArray(t *testing.T) {
	cases := []struct {
		body string
		want bool
	}{
		{`[{"level":"info"}]`, true},
		{"   \n\t [", true},
		{`{"level":"info"}`, false},
		{"   {", false},
		{"", false},
		{"null", false},
	}
	for _, tc := range cases {
		if got := rootIsJSONArray([]byte(tc.body)); got != tc.want {
			t.Fatalf("rootIsJSONArray(%q)=%v, want %v", tc.body, got, tc.want)
		}
	}
}

func newCtx(method, target, body string, orgID string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if orgID != "" {
		c.Set(internalauth.OrgIDContextKey, orgID)
	}
	return c, rec
}

func TestV3LogsRejectsSingleObject(t *testing.T) {
	h := &v3LogsHandler{} // logService not reached: rejected before ingest
	c, rec := newCtx(http.MethodPost, "/api/v3/logs", `{"level":"info","message":"x"}`, "org123")
	if err := h.submit(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for single object, got %d (%s)", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "error") {
		t.Fatalf("expected error envelope, got %s", rec.Body.String())
	}
}

func TestV3LogsAcceptsArrayShape(t *testing.T) {
	// A valid array passes the array gate (rootIsJSONArray). We assert the gate
	// here; full ingestion is covered by integration tests with ClickHouse.
	if !rootIsJSONArray([]byte(`[{"level":"info","message":"x"}]`)) {
		t.Fatal("valid array should pass the root-array check")
	}
}

func TestV3LogsRequiresAuth(t *testing.T) {
	h := &v3LogsHandler{}
	c, rec := newCtx(http.MethodPost, "/api/v3/logs", `[{"level":"info","message":"x"}]`, "")
	if err := h.submit(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without org context, got %d", rec.Code)
	}
}

func TestV3FileUploadRequiresAuth(t *testing.T) {
	h := &v3FileHandler{} // fileService not reached: auth fails first
	c, rec := newCtx(http.MethodPost, "/api/v3/file", "", "")
	if err := h.upload(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without org context, got %d", rec.Code)
	}
}

func TestV3PresignedUploadRejectsBadToken(t *testing.T) {
	h := &v3PresignedHandler{fileService: &file.Service{}}
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v3/file/presigned-url/not-a-valid-token", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("token")
	c.SetParamValues("not-a-valid-token")

	if err := h.upload(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid presigned token, got %d (%s)", rec.Code, rec.Body.String())
	}
}

func TestV3PresignedGenerateReturnsURL(t *testing.T) {
	h := &v3PresignedHandler{fileService: &file.Service{}}
	c, rec := newCtx(http.MethodGet, "/api/v3/file/presigned-url", "", "org123")
	if err := h.generate(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "presignedUrl") || !strings.Contains(body, "/api/v3/file/presigned-url/") {
		t.Fatalf("expected a presigned upload URL, got %s", body)
	}
}
