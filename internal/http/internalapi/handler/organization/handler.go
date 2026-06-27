package organization

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/fivemanage/lite/api"
	"github.com/fivemanage/lite/internal/clickhouse"
	"github.com/fivemanage/lite/internal/http/appctx"
	"github.com/fivemanage/lite/internal/http/httputil"
	"github.com/fivemanage/lite/internal/http/validator"
	"github.com/labstack/echo/v4"
)

// createOrganizationHandler godoc
// @Summary      Create organization
// @Description  Create a new organization
// @Tags         organization
// @Accept       json
// @Produce      json
// @Param        data  body      api.CreateOrganizationRequest  true  "Create Organization Request"
// @Success      200   {object}  httputil.ResponseData{data=api.Organization}
// @Failure      500   {object}  httputil.ErrorResponseData
// @Router       /dash/organization [post]
func (r *handler) createOrganizationHandler(c echo.Context) error {
	cc := c.(*appctx.Context)
	ctx := cc.Request().Context()

	var data api.CreateOrganizationRequest
	if err := validator.BindAndValidate(cc, &data); err != nil {
		slog.Error("failed to bind and validate token", "err", err)
		return cc.JSON(500, httputil.ErrorResponse(err.Error()))
	}

	user := cc.User()

	organization, err := r.organizationService.CreateOrganization(ctx, &data, user.ID)
	if err != nil {
		slog.Error("failed to create organization", "err", err)
		return cc.JSON(500, httputil.ErrorResponse(err.Error()))
	}

	return cc.JSON(200, httputil.Response(organization))
}

// listOrganizationsHandler godoc
// @Summary      List organizations
// @Description  List all organizations
// @Tags         organization
// @Produce      json
// @Success      200  {object}  httputil.ResponseData{data=[]api.Organization}
// @Failure      500  {object}  httputil.ErrorResponseData
// @Router       /dash/organization [get]
func (r *handler) listOrganizationsHandler(c echo.Context) error {
	cc := c.(*appctx.Context)
	ctx := cc.Request().Context()

	user := cc.User()
	organizations, err := r.organizationService.ListOrganizations(ctx, user.ID)
	if err != nil {
		slog.Error("failed to list organizations", "err", err)
		return cc.JSON(500, httputil.ErrorResponse(err.Error()))
	}

	return cc.JSON(200, httputil.Response(organizations))
}

// getOrganizationHandler godoc
// @Summary      Get organization
// @Description  Get organization by ID
// @Tags         organization
// @Produce      json
// @Param        id   path      string  true  "Organization ID"
// @Success      200  {object}  httputil.ResponseData{data=api.Organization}
// @Failure      500  {object}  httputil.ErrorResponseData
// @Router       /dash/organization/{id} [get]
func (r *handler) getOrganizationHandler(c echo.Context) error {
	cc := c.(*appctx.Context)
	ctx := cc.Request().Context()

	id := cc.Param("organizationId")

	organization, err := r.organizationService.FindOrganizationByID(ctx, id)
	if err != nil {
		slog.Error("failed to find organization", "err", err)
		return cc.JSON(500, httputil.ErrorResponse(err.Error()))
	}

	return cc.JSON(200, httputil.Response(organization))
}

// updateOrganizationHandler godoc
// @Summary      Update organization
// @Description  Rename an organization
// @Tags         organization
// @Accept       json
// @Produce      json
// @Param        id    path      string                         true  "Organization ID"
// @Param        data  body      api.UpdateOrganizationRequest  true  "Update Organization Request"
// @Success      200   {object}  httputil.ResponseData{data=api.Organization}
// @Failure      500   {object}  httputil.ErrorResponseData
// @Router       /dash/organization/{id} [patch]
func (r *handler) updateOrganizationHandler(c echo.Context) error {
	cc := c.(*appctx.Context)
	ctx := cc.Request().Context()
	id := cc.Param("organizationId")

	var data api.UpdateOrganizationRequest
	if err := validator.BindAndValidate(cc, &data); err != nil {
		return cc.JSON(http.StatusBadRequest, httputil.ErrorResponse(err.Error()))
	}

	organization, err := r.organizationService.UpdateOrganization(ctx, id, data.Name)
	if err != nil {
		slog.Error("failed to update organization", "err", err)
		return cc.JSON(http.StatusInternalServerError, httputil.ErrorResponse(err.Error()))
	}

	return cc.JSON(http.StatusOK, httputil.Response(organization))
}

// deleteOrganizationHandler godoc
// @Summary      Delete organization
// @Description  Permanently delete an organization and its data
// @Tags         organization
// @Produce      json
// @Param        id   path  string  true  "Organization ID"
// @Success      204
// @Failure      500  {object}  httputil.ErrorResponseData
// @Router       /dash/organization/{id} [delete]
func (r *handler) deleteOrganizationHandler(c echo.Context) error {
	cc := c.(*appctx.Context)
	ctx := cc.Request().Context()
	id := cc.Param("organizationId")

	if err := r.organizationService.DeleteOrganization(ctx, id); err != nil {
		slog.Error("failed to delete organization", "err", err)
		return cc.JSON(http.StatusInternalServerError, httputil.ErrorResponse(err.Error()))
	}

	return cc.NoContent(http.StatusNoContent)
}

// getOrganizationUsageHandler godoc
// @Summary      Get organization usage
// @Description  Get detailed storage and logging usage for an organization
// @Tags         organization
// @Produce      json
// @Param        id   path      string  true  "Organization ID"
// @Success      200  {object}  httputil.ResponseData{data=api.OrganizationUsage}
// @Failure      500  {object}  httputil.ErrorResponseData
// @Router       /dash/organization/{id}/usage [get]
func (r *handler) getOrganizationUsageHandler(c echo.Context) error {
	cc := c.(*appctx.Context)
	ctx := cc.Request().Context()
	id := cc.Param("organizationId")

	usage, err := r.organizationService.GetUsage(ctx, id)
	if err != nil {
		slog.Error("failed to get organization usage", "err", err)
		return cc.JSON(http.StatusInternalServerError, httputil.ErrorResponse(err.Error()))
	}

	return cc.JSON(http.StatusOK, httputil.Response(usage))
}

// getOrganizationStatsHandler godoc
// @Summary      Get organization stats
// @Description  Get statistics for an organization
// @Tags         organization
// @Produce      json
// @Param        id   path      string  true  "Organization ID"
// @Success      200  {object}  httputil.ResponseData{data=api.OrganizationStats}
// @Failure      500  {object}  httputil.ErrorResponseData
// @Router       /dash/organization/{id}/stats [get]
func (r *handler) getOrganizationStatsHandler(c echo.Context) error {
	cc := c.(*appctx.Context)
	ctx := cc.Request().Context()

	id := cc.Param("organizationId")

	stats, err := r.organizationService.GetStats(ctx, id)
	if err != nil {
		slog.Error("failed to get organization stats", "err", err)
		if errors.Is(err, clickhouse.ErrUnavailable) {
			return cc.JSON(http.StatusServiceUnavailable, httputil.ErrorResponse("logging is unavailable"))
		}
		return cc.JSON(500, httputil.ErrorResponse(err.Error()))
	}

	return cc.JSON(200, httputil.Response(stats))
}
