package publicapi

import (
	"net/http"
	"strings"

	"github.com/fivemanage/lite/api"
	"github.com/fivemanage/lite/internal/auth"
	"github.com/fivemanage/lite/internal/http/httputil"
	"github.com/fivemanage/lite/internal/service/file"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

type v3FileHandler struct{ fileService *file.Service }

// registerV3FileApi registers the token-gated V3 file routes. The presigned-url
// upload route is registered separately (see registerV3PresignedUpload) because
// it must bypass API-key auth.
func registerV3FileApi(group *echo.Group, fileService *file.Service) {
	h := &v3FileHandler{fileService: fileService}
	group.POST("/v3/file", h.upload, echoMiddleware.BodyLimit("500M"))
	// base64 bodies inflate ~33%, so allow extra headroom over the 500M file cap.
	group.POST("/v3/file/base64", h.uploadBase64, echoMiddleware.BodyLimit("700M"))
	group.GET("/v3/file", h.list)
	group.GET("/v3/file/*", h.get)
	group.DELETE("/v3/file/*", h.delete)
}

func (h *v3FileHandler) upload(c echo.Context) error {
	ctx := c.Request().Context()
	orgID, err := auth.CurrentOrgId(c)
	if err != nil {
		return v3Error(c, http.StatusUnauthorized, "Unauthorized: "+err.Error())
	}
	// Prefer the documented "file" field, but also accept media-typed field
	// names (image/video/audio) so FiveM clients that post under those names
	// (e.g. screenshot-basic, which defaults to "image") work unchanged.
	f, header, err := httputil.FileAny(c.Request(), "image", "video", "audio")
	if err != nil {
		return v3Error(c, http.StatusBadRequest, "missing file: "+err.Error())
	}
	defer f.Close()

	params, err := v3ParamsFromForm(c)
	if err != nil {
		return v3Error(c, http.StatusBadRequest, "invalid metadata: "+err.Error())
	}

	item, err := h.fileService.CreateFileV3(ctx, orgID, f, header, params)
	if err != nil {
		return v3ServiceError(c, uploadStatus(err), err, "failed to upload file")
	}
	return c.JSON(http.StatusOK, httputil.Response(uploadResult(item)))
}

func (h *v3FileHandler) uploadBase64(c echo.Context) error {
	ctx := c.Request().Context()
	orgID, err := auth.CurrentOrgId(c)
	if err != nil {
		return v3Error(c, http.StatusUnauthorized, "Unauthorized: "+err.Error())
	}
	var req api.UploadFileBase64Request
	if err := c.Bind(&req); err != nil {
		return v3Error(c, http.StatusBadRequest, "invalid request body: "+err.Error())
	}
	item, err := h.fileService.CreateFileBase64V3(ctx, orgID, req)
	if err != nil {
		return v3ServiceError(c, uploadStatus(err), err, "failed to upload file")
	}
	return c.JSON(http.StatusOK, httputil.Response(uploadResult(item)))
}

func (h *v3FileHandler) list(c echo.Context) error {
	ctx := c.Request().Context()
	orgID, err := auth.CurrentOrgId(c)
	if err != nil {
		return v3Error(c, http.StatusUnauthorized, "Unauthorized: "+err.Error())
	}
	page, _ := httputil.QueryInt(c, "page", 1)
	limit, _ := httputil.QueryInt(c, "limit", file.DefaultListLimit)
	fileType := c.QueryParam("type")
	folder := c.QueryParam("path")

	items, total, err := h.fileService.ListFilesV3(ctx, orgID, fileType, folder, page, limit)
	if err != nil {
		return v3ServiceError(c, http.StatusInternalServerError, err, "failed to list files")
	}
	normPage, normLimit, _ := file.NormalizeListParams(page, limit)
	return c.JSON(http.StatusOK, v3ListFilesResponse{
		Status:     "ok",
		Data:       items,
		Pagination: api.PaginationV3{Page: normPage, Limit: normLimit, Total: total},
	})
}

func (h *v3FileHandler) get(c echo.Context) error {
	ctx := c.Request().Context()
	orgID, err := auth.CurrentOrgId(c)
	if err != nil {
		return v3Error(c, http.StatusUnauthorized, "Unauthorized: "+err.Error())
	}
	idOrKey := strings.TrimPrefix(c.Param("*"), "/")
	if idOrKey == "" {
		return v3Error(c, http.StatusBadRequest, "file id or key is required")
	}
	item, err := h.fileService.GetFileV3(ctx, orgID, idOrKey)
	if err != nil {
		return v3ServiceError(c, notFoundOrServer(err), err, "failed to get file")
	}
	return c.JSON(http.StatusOK, httputil.Response(item))
}

func (h *v3FileHandler) delete(c echo.Context) error {
	ctx := c.Request().Context()
	orgID, err := auth.CurrentOrgId(c)
	if err != nil {
		return v3Error(c, http.StatusUnauthorized, "Unauthorized: "+err.Error())
	}
	idOrKey := strings.TrimPrefix(c.Param("*"), "/")
	if idOrKey == "" {
		return v3Error(c, http.StatusBadRequest, "file id or key is required")
	}
	if err := h.fileService.DeleteFileV3(ctx, orgID, idOrKey); err != nil {
		return v3ServiceError(c, notFoundOrServer(err), err, "failed to delete file")
	}
	return c.JSON(http.StatusOK, echo.Map{"status": "ok"})
}

func uploadResult(item *api.FileItemV3) api.UploadFileV3Response {
	return api.UploadFileV3Response{ID: item.ID, URL: item.URL, OriginalURL: item.OriginalURL}
}

// v3ParamsFromForm extracts the optional V3 upload fields from a multipart form.
// retentionExempt is accepted under both camelCase and snake_case (the docs use
// snake_case on the presigned-upload route).
func v3ParamsFromForm(c echo.Context) (file.CreateFileV3Params, error) {
	metadata, err := file.ParseMetadataJSON(c.FormValue("metadata"))
	if err != nil {
		return file.CreateFileV3Params{}, err
	}
	return file.CreateFileV3Params{
		Filename:        c.FormValue("filename"),
		Path:            c.FormValue("path"),
		Metadata:        metadata,
		RetentionExempt: parseFormBool(c.FormValue("retentionExempt")) || parseFormBool(c.FormValue("retention_exempt")),
	}, nil
}

func parseFormBool(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "true", "1", "yes", "on":
		return true
	default:
		return false
	}
}
