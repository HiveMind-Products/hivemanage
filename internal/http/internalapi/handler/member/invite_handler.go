package member

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/fivemanage/lite/api"
	internalauth "github.com/fivemanage/lite/internal/auth"
	"github.com/fivemanage/lite/internal/database"
	"github.com/fivemanage/lite/internal/http/appctx"
	"github.com/fivemanage/lite/internal/http/httputil"
	"github.com/fivemanage/lite/internal/http/validator"
	inviteservice "github.com/fivemanage/lite/internal/service/invite"
	"github.com/labstack/echo/v4"
)

func (h *handler) createInviteHandler(c echo.Context) error {
	cc := c.(*appctx.Context)
	ctx := cc.Request().Context()
	organizationID := cc.Param("organizationId")

	var req api.CreateInviteRequest
	if err := validator.BindAndValidate(cc, &req); err != nil {
		return cc.JSON(http.StatusBadRequest, httputil.ErrorResponse(err.Error()))
	}

	user, err := internalauth.CurrentUser(cc)
	if err != nil {
		return cc.JSON(http.StatusUnauthorized, httputil.ErrorResponse("Not authenticated"))
	}

	invite, err := h.inviteService.Create(ctx, inviteservice.CreateParams{
		OrganizationID:  organizationID,
		Role:            req.Role,
		Permissions:     req.Permissions,
		DiscordUsername: req.DiscordUsername,
		Email:           req.Email,
		CreatedBy:       user.ID,
	})
	if err != nil {
		slog.Error("failed to create invite", "err", err)
		return cc.JSON(http.StatusInternalServerError, httputil.ErrorResponse("Failed to create invite"))
	}

	return cc.JSON(http.StatusOK, httputil.Response(inviteResponse(cc, invite)))
}

func (h *handler) listInvitesHandler(c echo.Context) error {
	cc := c.(*appctx.Context)
	ctx := cc.Request().Context()
	organizationID := cc.Param("organizationId")

	invites, err := h.inviteService.ListPending(ctx, organizationID)
	if err != nil {
		slog.Error("failed to list invites", "err", err)
		return cc.JSON(http.StatusInternalServerError, httputil.ErrorResponse("Failed to list invites"))
	}

	out := make([]api.InviteResponse, 0, len(invites))
	for _, invite := range invites {
		out = append(out, inviteResponse(cc, invite))
	}
	return cc.JSON(http.StatusOK, httputil.Response(out))
}

func (h *handler) deleteInviteHandler(c echo.Context) error {
	cc := c.(*appctx.Context)
	ctx := cc.Request().Context()
	organizationID := cc.Param("organizationId")
	inviteID := cc.Param("inviteId")

	err := h.inviteService.Delete(ctx, organizationID, inviteID)
	if err != nil {
		if errors.Is(err, inviteservice.ErrInviteInvalid) {
			return cc.JSON(http.StatusNotFound, httputil.ErrorResponse("Invite not found"))
		}
		slog.Error("failed to delete invite", "err", err)
		return cc.JSON(http.StatusInternalServerError, httputil.ErrorResponse("Failed to delete invite"))
	}

	return cc.NoContent(http.StatusNoContent)
}

func inviteResponse(cc *appctx.Context, invite *database.Invite) api.InviteResponse {
	url := cc.Scheme() + "://" + cc.Request().Host + "/invite/" + invite.ID
	return api.InviteResponse{
		ID:              invite.ID,
		URL:             url,
		OrganizationID:  invite.OrganizationID,
		Role:            invite.Role,
		DiscordUsername: invite.DiscordUsername,
		Email:           invite.Email,
		ExpiresAt:       invite.ExpiresAt,
		AcceptedAt:      invite.AcceptedAt,
		CreatedAt:       invite.CreatedAt,
	}
}
