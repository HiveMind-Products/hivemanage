package dataset

import (
	"github.com/fivemanage/lite/api"
	"github.com/fivemanage/lite/internal/http/middleware"
	"github.com/fivemanage/lite/internal/service/auth"
	"github.com/fivemanage/lite/internal/service/dataset"
	"github.com/labstack/echo/v4"
)

type handler struct{ datasetService *dataset.Service }

func RegisterRoutes(group *echo.Group, datasetService *dataset.Service, authService *auth.Service) {
	h := handler{datasetService: datasetService}
	readLogs := middleware.OrganizationPermission(authService, api.PermissionModuleLogs, api.PermissionActionRead)
	writeLogs := middleware.OrganizationPermission(authService, api.PermissionModuleLogs, api.PermissionActionWrite)
	group.POST("/:organizationId/dataset", h.createDatasetHandler, writeLogs)
	group.GET("/:organizationId/dataset", h.listDatasetsHandler, readLogs)
	group.GET("/:organizationId/dataset/:datasetId/fields", h.listDatasetFieldsHandler, readLogs)
	group.POST("/:organizationId/dataset/:datasetId/logs", h.queryDatasetLogsHandler, readLogs)
	group.GET("/:organizationId/dataset/:datasetId/logs/:logId", h.getDatasetLogHandler, readLogs)
}
