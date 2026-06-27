package publicapi

import (
	"errors"
	"net/http"

	"github.com/fivemanage/lite/api"
	"github.com/fivemanage/lite/internal/service/file"
	"github.com/labstack/echo/v4"
)

// v3Error writes the V3 error envelope: { "error": "..." }.
func v3Error(c echo.Context, code int, message string) error {
	return c.JSON(code, echo.Map{"error": message})
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
