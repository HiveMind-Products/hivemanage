package publicapi

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/fivemanage/lite/api"
	"github.com/fivemanage/lite/internal/service/file"
	"github.com/labstack/echo/v4"
)

// v3Error writes the V3 error envelope: { "error": "..." }.
func v3Error(c echo.Context, code int, message string) error {
	return c.JSON(code, echo.Map{"error": message})
}

// v3ServiceError writes a V3 error response for a service error. For client
// errors (status < 500) the underlying message is safe to surface; for server
// errors the full error is logged server-side and a generic message is returned
// to the client so internal details are not leaked.
func v3ServiceError(c echo.Context, status int, err error, genericMsg string) error {
	if status >= http.StatusInternalServerError {
		slog.Error(genericMsg, "method", c.Request().Method, "path", c.Path(), "err", err)
		return v3Error(c, status, genericMsg)
	}
	return v3Error(c, status, err.Error())
}

// v3ListFilesResponse is the envelope for GET /api/v3/file (list).
type v3ListFilesResponse struct {
	Status     string            `json:"status"`
	Data       []*api.FileItemV3 `json:"data"`
	Pagination api.PaginationV3  `json:"pagination"`
}

// uploadStatus maps a file-service error to the documented HTTP status. Upload
// validation problems (bad input, size, metadata, base64) are client errors;
// anything else is treated as a server error.
func uploadStatus(err error) int {
	var ue file.UploadStorageError
	if errors.As(err, &ue) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

// notFoundOrServer maps get/delete errors: missing files are 404, the rest 500.
func notFoundOrServer(err error) int {
	var ge *file.GetFileError
	if errors.As(err, &ge) {
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}
