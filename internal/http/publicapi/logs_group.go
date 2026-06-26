package publicapi

import (
	"errors"
	"net/http"

	"github.com/fivemanage/lite/api"
	"github.com/fivemanage/lite/internal/auth"
	"github.com/fivemanage/lite/internal/clickhouse"
	"github.com/fivemanage/lite/internal/http/httputil"
	"github.com/fivemanage/lite/internal/http/middleware"
	"github.com/fivemanage/lite/internal/service/log"
	"github.com/fivemanage/lite/internal/service/token"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

func registerLogsApi(group *echo.Group, logService *log.Service, tokenService *token.Service) {
	h := &logsHandler{logService: logService}
	group.POST("/logs", h.submitLogs, echoMiddleware.BodyLimit("2M"), middleware.TokenAuth(tokenService))
}

type logsHandler struct{ logService *log.Service }

func (h *logsHandler) submitLogs(c echo.Context) error {
	ctx := c.Request().Context()
	dataset := c.Request().Header.Get("X-Fivemanage-Dataset")
	orgID, err := auth.CurrentOrgId(c)
	if err != nil {
		return err
	}
	var logs []api.Log
	if err := c.Bind(&logs); err != nil {
		return err
	}
	if err := h.logService.SubmitLogs(ctx, orgID, dataset, logs); err != nil {
		if errors.Is(err, clickhouse.ErrUnavailable) {
			return c.JSON(http.StatusServiceUnavailable, httputil.ErrorResponse("logging is unavailable"))
		}
		return c.JSON(http.StatusBadRequest, httputil.ErrorResponse(err.Error()))
	}
	return c.JSON(200, httputil.Response("ok"))
}
