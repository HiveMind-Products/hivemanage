package publicapi

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/fivemanage/lite/api"
	"github.com/fivemanage/lite/internal/auth"
	"github.com/fivemanage/lite/internal/clickhouse"
	"github.com/fivemanage/lite/internal/http/middleware"
	"github.com/fivemanage/lite/internal/service/log"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

type v3LogsHandler struct{ logService *log.Service }

func registerV3LogsApi(group *echo.Group, logService *log.Service) {
	h := &v3LogsHandler{logService: logService}
	logsWrite := middleware.RequireTokenScope(api.PermissionModuleLogs, api.PermissionActionWrite)
	group.POST("/v3/logs", h.submit, echoMiddleware.BodyLimit("2M"), logsWrite)
}

// submit ingests a batch of logs. The root body MUST be a JSON array; a single
// object at the root is rejected with 400.
func (h *v3LogsHandler) submit(c echo.Context) error {
	ctx := c.Request().Context()
	orgID, err := auth.CurrentOrgId(c)
	if err != nil {
		return v3Error(c, http.StatusUnauthorized, "Unauthorized: "+err.Error())
	}

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return v3Error(c, http.StatusBadRequest, "failed to read request body")
	}
	if !rootIsJSONArray(body) {
		return v3Error(c, http.StatusBadRequest, "request body must be a JSON array of log objects")
	}

	var logs []api.LogV3
	if err := json.Unmarshal(body, &logs); err != nil {
		return v3Error(c, http.StatusBadRequest, "invalid logs payload: "+err.Error())
	}

	dataset := c.Request().Header.Get("X-Fivemanage-Dataset")
	if err := h.logService.SubmitLogs(ctx, orgID, dataset, toServiceLogs(logs)); err != nil {
		switch {
		case errors.Is(err, clickhouse.ErrUnavailable):
			return v3Error(c, http.StatusServiceUnavailable, "logging is unavailable")
		case errors.Is(err, clickhouse.ErrWriteFailed):
			slog.Error("failed to store logs", "err", err)
			return v3Error(c, http.StatusInternalServerError, "failed to store logs")
		default:
			// Remaining errors are client-side validation problems (empty batch,
			// invalid level/message/metadata) and are safe to surface as 400.
			return v3Error(c, http.StatusBadRequest, err.Error())
		}
	}
	return c.JSON(http.StatusOK, echo.Map{"status": "ok"})
}

// rootIsJSONArray reports whether the first non-whitespace byte of the body is a
// JSON array opener. Used to reject a single object at the root.
func rootIsJSONArray(body []byte) bool {
	for _, b := range body {
		switch b {
		case ' ', '\t', '\n', '\r':
			continue
		case '[':
			return true
		default:
			return false
		}
	}
	return false
}

// toServiceLogs maps the V3 log request shape onto the internal log type. The V3
// timestamp field is intentionally dropped (see api.LogV3.Timestamp).
func toServiceLogs(in []api.LogV3) []api.Log {
	out := make([]api.Log, len(in))
	for i, l := range in {
		out[i] = api.Log{
			Level:    l.Level,
			Message:  l.Message,
			Resource: l.Resource,
			Metadata: l.Metadata,
		}
	}
	return out
}
