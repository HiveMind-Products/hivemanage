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
	"github.com/fivemanage/lite/internal/http/httputil"
	"github.com/fivemanage/lite/internal/http/middleware"
	"github.com/fivemanage/lite/internal/service/log"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

// registerLogsApi registers the legacy V2 log ingest route.
//
// Deprecated: superseded by POST /api/v3/logs (which requires a root array).
// Retained for backward compatibility; unlike V3 it also accepts a single log
// object at the root and wraps it into a one-element batch.
func registerLogsApi(group *echo.Group, logService *log.Service) {
	h := &logsHandler{logService: logService}
	logsWrite := middleware.RequireTokenScope(api.PermissionModuleLogs, api.PermissionActionWrite)
	group.POST("/logs", h.submitLogs, echoMiddleware.BodyLimit("2M"), logsWrite)
}

type logsHandler struct{ logService *log.Service }

func (h *logsHandler) submitLogs(c echo.Context) error {
	c.Response().Header().Set("Deprecation", "true")
	ctx := c.Request().Context()
	dataset := c.Request().Header.Get("X-Fivemanage-Dataset")
	orgID, err := auth.CurrentOrgId(c)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, httputil.ErrorResponse("failed to read request body"))
	}

	// V2 compatibility: accept either an array or a single log object at the root.
	var logs []api.Log
	if rootIsJSONArray(body) {
		if err := json.Unmarshal(body, &logs); err != nil {
			return c.JSON(http.StatusBadRequest, httputil.ErrorResponse(err.Error()))
		}
	} else {
		var single api.Log
		if err := json.Unmarshal(body, &single); err != nil {
			return c.JSON(http.StatusBadRequest, httputil.ErrorResponse(err.Error()))
		}
		logs = []api.Log{single}
	}

	if err := h.logService.SubmitLogs(ctx, orgID, dataset, logs); err != nil {
		switch {
		case errors.Is(err, clickhouse.ErrUnavailable):
			return c.JSON(http.StatusServiceUnavailable, httputil.ErrorResponse("logging is unavailable"))
		case errors.Is(err, clickhouse.ErrWriteFailed):
			slog.Error("failed to store logs", "err", err)
			return c.JSON(http.StatusInternalServerError, httputil.ErrorResponse("failed to store logs"))
		default:
			// Remaining errors are client-side validation problems and are safe
			// to surface as 400.
			return c.JSON(http.StatusBadRequest, httputil.ErrorResponse(err.Error()))
		}
	}
	return c.JSON(200, httputil.Response("ok"))
}
