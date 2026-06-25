package member

import (
	"errors"
	"strconv"

	"github.com/fivemanage/lite/api"
	"github.com/fivemanage/lite/internal/http/appctx"
	"github.com/fivemanage/lite/internal/http/httputil"
	"github.com/fivemanage/lite/internal/http/validator"
	memberservice "github.com/fivemanage/lite/internal/service/member"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// listMembersHandler godoc
// @Summary      List members
// @Description  List members of an organization
// @Tags         member
// @Produce      json
// @Param        organizationId   path      string  true  "Organization ID"
// @Success      200  {object}  httputil.ResponseData{data=[]api.OrganizationMember}
// @Failure      500  {object}  httputil.ErrorResponseData
// @Router       /dash/organization/{organizationId}/member [get]
func (r *handler) listMembersHandler(c echo.Context) error {
	cc := c.(*appctx.Context)
	ctx := cc.Request().Context()

	id := cc.Param("organizationId")

	members, err := r.memberService.ListMembers(ctx, id)
	if err != nil {
		logrus.WithError(err).Error("failed to list members")
		return cc.JSON(500, httputil.ErrorResponse(err.Error()))
	}

	return cc.JSON(200, httputil.Response(members))
}

// addMemberHandler godoc
// @Summary      Add member
// @Description  Create a new user and add them to the organization
// @Tags         member
// @Accept       json
// @Produce      json
// @Param        organizationId   path      string                    true  "Organization ID"
// @Param        data             body      api.CreateMemberRequest   true  "Create Member Request"
// @Success      200  {object}  httputil.ResponseData{data=api.CreateMemberResponse}
// @Failure      500  {object}  httputil.ErrorResponseData
// @Router       /dash/organization/{organizationId}/member [post]
func (r *handler) addMemberHandler(c echo.Context) error {
	cc := c.(*appctx.Context)
	ctx := cc.Request().Context()

	organizationID := cc.Param("organizationId")

	var data api.CreateMemberRequest
	if err := validator.BindAndValidate(cc, &data); err != nil {
		logrus.WithError(err).Error("failed to bind and validate member request")
		return cc.JSON(400, httputil.ErrorResponse(err.Error()))
	}

	resp, err := r.memberService.AddMember(ctx, organizationID, &data)
	if err != nil {
		logrus.WithError(err).Error("failed to add member")
		return cc.JSON(500, httputil.ErrorResponse(err.Error()))
	}

	return cc.JSON(200, httputil.Response(resp))
}

// updateMemberHandler godoc
// @Summary      Update member
// @Description  Update a member role and permissions
// @Tags         member
// @Accept       json
// @Produce      json
// @Param        organizationId   path      string                   true  "Organization ID"
// @Param        memberId         path      string                   true  "Member ID"
// @Param        data             body      api.UpdateMemberRequest  true  "Update Member Request"
// @Success      200  {object}  httputil.ResponseData{data=api.OrganizationMember}
// @Failure      500  {object}  httputil.ErrorResponseData
// @Router       /dash/organization/{organizationId}/member/{memberId} [patch]
func (r *handler) updateMemberHandler(c echo.Context) error {
	cc := c.(*appctx.Context)
	ctx := cc.Request().Context()

	organizationID := cc.Param("organizationId")
	memberIDStr := cc.Param("memberId")

	memberID, err := strconv.ParseInt(memberIDStr, 10, 64)
	if err != nil {
		return cc.JSON(400, httputil.ErrorResponse("invalid member ID"))
	}

	var data api.UpdateMemberRequest
	if err := validator.BindAndValidate(cc, &data); err != nil {
		logrus.WithError(err).Error("failed to bind and validate update member request")
		return cc.JSON(400, httputil.ErrorResponse(err.Error()))
	}

	updated, err := r.memberService.UpdateMember(ctx, organizationID, memberID, &data)
	if err != nil {
		logrus.WithError(err).Error("failed to update member")
		if errors.Is(err, memberservice.ErrLastAdmin) {
			return cc.JSON(400, httputil.ErrorResponse(err.Error()))
		}
		return cc.JSON(500, httputil.ErrorResponse(err.Error()))
	}

	return cc.JSON(200, httputil.Response(updated))
}

// removeMemberHandler godoc
// @Summary      Remove member
// @Description  Remove a member from the organization
// @Tags         member
// @Produce      json
// @Param        organizationId   path      string  true  "Organization ID"
// @Param        memberId         path      string  true  "Member ID"
// @Success      200  {object}  httputil.ResponseData{data=nil}
// @Failure      500  {object}  httputil.ErrorResponseData
// @Router       /dash/organization/{organizationId}/member/{memberId} [delete]
func (r *handler) removeMemberHandler(c echo.Context) error {
	cc := c.(*appctx.Context)
	ctx := cc.Request().Context()

	organizationID := cc.Param("organizationId")
	memberIDStr := cc.Param("memberId")

	memberID, err := strconv.ParseInt(memberIDStr, 10, 64)
	if err != nil {
		return cc.JSON(400, httputil.ErrorResponse("invalid member ID"))
	}

	if err := r.memberService.RemoveMember(ctx, organizationID, memberID); err != nil {
		logrus.WithError(err).Error("failed to remove member")
		if errors.Is(err, memberservice.ErrLastAdmin) {
			return cc.JSON(400, httputil.ErrorResponse(err.Error()))
		}
		return cc.JSON(500, httputil.ErrorResponse(err.Error()))
	}

	return cc.JSON(200, httputil.Response(nil))
}
