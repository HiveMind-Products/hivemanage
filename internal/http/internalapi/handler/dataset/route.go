package dataset

import (
	"github.com/fivemanage/lite/internal/http/middleware"
	"github.com/fivemanage/lite/internal/service/auth"
	"github.com/fivemanage/lite/internal/service/dataset"
	"github.com/labstack/echo/v4"
)

type handler struct{ datasetService *dataset.Service }

func RegisterRoutes(group *echo.Group, datasetService *dataset.Service, authService *auth.Service) {
	h := handler{datasetService: datasetService}
	group.POST("/:organizationId/dataset", h.createDatasetHandler, middleware.OrganizationAdmin(authService))
	group.GET("/:organizationId/dataset", h.listDatasetsHandler)
	group.GET("/:organizationId/dataset/:datasetId/fields", h.listDatasetFieldsHandler)
	group.POST("/:organizationId/dataset/:datasetId/logs", h.queryDatasetLogsHandler)
	group.GET("/:organizationId/dataset/:datasetId/logs/:logId", h.getDatasetLogHandler)
}
