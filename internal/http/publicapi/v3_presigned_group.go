package publicapi

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fivemanage/lite/api"
	"github.com/fivemanage/lite/internal/auth"
	"github.com/fivemanage/lite/internal/http/httputil"
	"github.com/fivemanage/lite/internal/http/middleware"
	"github.com/fivemanage/lite/internal/service/file"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

type v3PresignedHandler struct{ fileService *file.Service }

// registerV3PresignedUpload registers the presigned-upload route. It MUST be
// called before the API-key auth middleware is applied to the group, because it
// authenticates via the signed presigned token rather than an API key.
func registerV3PresignedUpload(group *echo.Group, fileService *file.Service) {
	h := &v3PresignedHandler{fileService: fileService}
	group.POST("/v3/file/presigned-url/:token", h.upload, echoMiddleware.BodyLimit("500M"))
}

// registerV3PresignedGenerate registers the (API-key gated) presigned URL
// generation route.
func registerV3PresignedGenerate(group *echo.Group, fileService *file.Service) {
	h := &v3PresignedHandler{fileService: fileService}
	storageWrite := middleware.RequireTokenScope(api.PermissionModuleStorage, api.PermissionActionWrite)
	group.GET("/v3/file/presigned-url", h.generate, storageWrite)
}

// generate issues a presigned upload URL. Optional query: expiresAt (unix
// seconds; default now + 15 min) and path (folder embedded in the token).
func (h *v3PresignedHandler) generate(c echo.Context) error {
	orgID, err := auth.CurrentOrgId(c)
	if err != nil {
		return v3Error(c, http.StatusUnauthorized, "Unauthorized: "+err.Error())
	}

	expiresAt := time.Now().Add(file.DefaultPresignedExpiry)
	if raw := c.QueryParam("expiresAt"); raw != "" {
		unix, err := httputil.QueryInt(c, "expiresAt", 0)
		if err != nil || unix <= 0 {
			return v3Error(c, http.StatusBadRequest, "invalid expiresAt: must be a unix timestamp")
		}
		expiresAt = time.Unix(int64(unix), 0)
	}

	token := h.fileService.CreatePresignedToken(orgID, expiresAt, c.QueryParam("path"))
	presignedURL := publicBaseURL(c) + "/api/v3/file/presigned-url/" + token
	return c.JSON(http.StatusOK, httputil.Response(echo.Map{"presignedUrl": presignedURL}))
}

// upload accepts the actual file for a previously generated presigned URL. It is
// authenticated by the token in the path, not by an API key.
func (h *v3PresignedHandler) upload(c echo.Context) error {
	ctx := c.Request().Context()

	orgID, tokenPath, err := h.fileService.ResolvePresignedToken(c.Param("token"))
	if err != nil {
		// Invalid or expired token: authenticated by the token, so 401.
		return v3Error(c, http.StatusUnauthorized, err.Error())
	}

	f, header, err := httputil.File(c.Request(), "file")
	if err != nil {
		return v3Error(c, http.StatusBadRequest, "missing file: "+err.Error())
	}
	defer f.Close()

	params, err := v3ParamsFromForm(c)
	if err != nil {
		return v3Error(c, http.StatusBadRequest, "invalid metadata: "+err.Error())
	}
	// If the presigned token pinned a folder, enforce it and ignore any
	// form-supplied path so the uploader cannot widen the token's scope.
	if tokenPath != "" {
		params.Path = tokenPath
	}

	item, err := h.fileService.CreateFileV3(ctx, orgID, f, header, params)
	if err != nil {
		return v3Error(c, uploadStatus(err), err.Error())
	}
	return c.JSON(http.StatusOK, httputil.Response(uploadResult(item)))
}

// publicBaseURL returns the base URL used to build the presigned upload link.
// It prefers the operator-configured PUBLIC_BASE_URL so the returned URL cannot
// be poisoned via a client-controlled Host header; it falls back to the request
// scheme/host when unset.
func publicBaseURL(c echo.Context) string {
	if base := strings.TrimRight(os.Getenv("PUBLIC_BASE_URL"), "/"); base != "" {
		return base
	}
	return c.Scheme() + "://" + c.Request().Host
}
