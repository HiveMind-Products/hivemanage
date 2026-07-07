package token

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/fivemanage/lite/api"
	"github.com/fivemanage/lite/internal/http/appctx"
	"github.com/fivemanage/lite/internal/http/httputil"
	"github.com/fivemanage/lite/internal/http/validator"
	"github.com/fivemanage/lite/internal/permissions"
	tokenservice "github.com/fivemanage/lite/internal/service/token"
	"github.com/labstack/echo/v4"
)

// createTokenHandler godoc
// @Summary      Create token
// @Description  Create a new API token for the current organization
// @Tags         token
// @Accept       json
// @Produce      json
// @Param        organizationId  path      string  true  "Organization ID"
// @Param        data            body      api.CreateTokenRequest  true  "Create Token Request"
// @Success      200             {object}  httputil.ResponseData{data=api.CreateTokenResponse}
// @Failure      500             {object}  httputil.ErrorResponseData
// @Router       /dash/{organizationId}/token [post]
func (r *handler) createTokenHandler(c echo.Context) error {
	cc := c.(*appctx.Context)
	ctx, cancel := context.WithTimeout(cc.Request().Context(), 5*time.Second)
	defer cancel()

	organizationID := cc.Param("organizationId")

	var data api.CreateTokenRequest
	if err := validator.BindAndValidate(cc, &data); err != nil {
		return err
	}

	data.OrganizationID = organizationID
	user := cc.User()
	apiToken, err := r.tokenService.CreateToken(ctx, &data, user.ID)
	if err != nil {
		var invalidScope *permissions.InvalidScopeError
		if errors.As(err, &invalidScope) || errors.Is(err, tokenservice.ErrExpiryInPast) {
			return cc.JSON(400, httputil.ErrorResponse(err.Error()))
		}
		slog.Error("failed to create token", "organization_id", organizationID, "err", err)
		return cc.JSON(500, httputil.ErrorResponse("Failed to create token"))
	}

	// we return the apiToken as a one-time thing
	tokenResponse := &api.CreateTokenResponse{
		Token: apiToken,
	}

	return cc.JSON(200, httputil.Response(tokenResponse))
}

// listTokensHandler godoc
// @Summary      List tokens
// @Description  List all API tokens for the current organization
// @Tags         token
// @Produce      json
// @Param        organizationId  path      string  true  "Organization ID"
// @Success      200             {object}  httputil.ResponseData{data=[]api.ListTokensResponse}
// @Failure      500             {object}  httputil.ErrorResponseData
// @Router       /dash/{organizationId}/token [get]
func (r *handler) listTokensHandler(c echo.Context) error {
	cc := c.(*appctx.Context)
	ctx := cc.Request().Context()

	organizationID := cc.Param("organizationId")

	tokens, err := r.tokenService.ListTokens(ctx, organizationID)
	if err != nil {
		slog.Error("failed to list tokens", "organization_id", organizationID, "err", err)
		return cc.JSON(500, httputil.ErrorResponse("Failed to list tokens"))
	}

	return cc.JSON(200, httputil.Response(tokens))
}

// deleteTokenHandler godoc
// @Summary      Delete token
// @Description  Delete an API token by ID
// @Tags         token
// @Produce      json
// @Param        organizationId  path      string  true  "Organization ID"
// @Param        id              path      string  true  "Token ID"
// @Success      200             {object}  httputil.ResponseData{data=nil}
// @Failure      500             {object}  httputil.ErrorResponseData
// @Router       /dash/{organizationId}/token/{id} [delete]
func (r *handler) deleteTokenHandler(c echo.Context) error {
	cc := c.(*appctx.Context)
	ctx := cc.Request().Context()

	organizationID := cc.Param("organizationId")
	tokenID := cc.Param("id")

	err := r.tokenService.DeleteToken(ctx, organizationID, tokenID)
	if err != nil {
		slog.Error("failed to delete token", "organization_id", organizationID, "token_id", tokenID, "err", err)
		return cc.JSON(500, httputil.ErrorResponse("Failed to delete token"))
	}

	return cc.JSON(200, httputil.Response(nil))
}
