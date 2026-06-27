package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/fivemanage/lite/api"
	"github.com/fivemanage/lite/internal/crypt"
	"github.com/fivemanage/lite/internal/http/appctx"
	"github.com/fivemanage/lite/internal/http/httputil"
	"github.com/fivemanage/lite/internal/http/validator"
	"github.com/fivemanage/lite/internal/service/auth"
	"github.com/fivemanage/lite/internal/service/invite"
	"github.com/labstack/echo/v4"

	internalauth "github.com/fivemanage/lite/internal/auth"
)

// getSessionHandler godoc
// @Summary      Get current session
// @Description  Get the current user session from the session cookie
// @Tags         auth
// @Produce      json
// @Success      200  {object}  httputil.ResponseData{data=api.User}
// @Failure      401  {object}  httputil.ErrorResponseData
// @Router       /dash/auth/session [get]
func (r *handler) getSessionHandler(c echo.Context) error {
	cc := c.(*appctx.Context)
	ctx := cc.Request().Context()

	sessionCookie, err := cc.Cookie(internalauth.SessionCookieName)
	if err != nil {
		return cc.JSON(http.StatusUnauthorized, httputil.ErrorResponse("Already logged out"))
	}

	user, err := r.authService.UserBySession(ctx, sessionCookie.Value)
	if err != nil {
		return cc.JSON(http.StatusUnauthorized, httputil.ErrorResponse("Invalid session"))
	}

	csrfToken, err := crypt.GenerateSessionID()
	if err != nil {
		return cc.JSON(http.StatusInternalServerError, httputil.ErrorResponse("Failed to create csrf token"))
	}
	cc.SetCookie(r.authService.CreateCSRFCookie(csrfToken))

	return cc.JSON(http.StatusOK, httputil.Response(user))
}

// loginHandler godoc
// @Summary      Login
// @Description  Login with username and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        login  body      api.LoginRequest  true  "Login Request"
// @Success      200    {object}  httputil.ResponseData{data=string}
// @Failure      400    {object}  httputil.ErrorResponseData
// @Failure      403    {object}  httputil.ErrorResponseData
// @Router       /dash/auth/login [post]
func (r *handler) loginHandler(c echo.Context) error {
	cc := c.(*appctx.Context)
	ctx := cc.Request().Context()

	var login api.LoginRequest
	if err := validator.BindAndValidate(cc, &login); err != nil {
		return cc.JSON(http.StatusBadRequest, httputil.ErrorResponse(err.Error()))
	}

	sessionID, err := r.authService.LoginUser(ctx, login.Username, login.Password)
	if err != nil {
		// this might not be the way we want to handle these errors
		// might be better to send some sort of code instead that we can map on the client?
		// as this could get veryyy long some places
		// errors.As() is also an option
		if errors.Is(err, auth.ErrUserCredentials{}) {
			return c.JSON(http.StatusForbidden, httputil.ErrorResponse("The username or password is wrong. Please try again"))
		}
		// this also bad
		return c.JSON(http.StatusForbidden, httputil.ErrorResponse(err.Error()))

	}

	sessionCookie := r.authService.CreateSessionCookie(sessionID)
	csrfToken, err := crypt.GenerateSessionID()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, httputil.ErrorResponse("Failed to create csrf token"))
	}

	c.SetCookie(sessionCookie)
	c.SetCookie(r.authService.CreateCSRFCookie(csrfToken))
	return c.JSON(http.StatusOK, httputil.Response("Login successful"))
}

const (
	discordStateCookie = "fmlite_oauth_state"
	inviteCookie       = "fmlite_invite"
	oauthModeCookie    = "fmlite_oauth_mode"
	oauthModeLink      = "link"
)

// setOAuthStateCookie stores the CSRF state for an in-flight OAuth exchange.
func setOAuthStateCookie(cc *appctx.Context, state string) {
	cc.SetCookie(&http.Cookie{
		Name:     discordStateCookie,
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// discordLoginHandler godoc
// @Summary      Begin Discord OAuth login
// @Description  Redirect the user to Discord to authorize the application
// @Tags         auth
// @Success      302
// @Router       /dash/auth/discord [get]
func (r *handler) discordLoginHandler(c echo.Context) error {
	cc := c.(*appctx.Context)

	state, err := crypt.GenerateSessionID()
	if err != nil {
		return cc.JSON(http.StatusInternalServerError, httputil.ErrorResponse("Failed to start login"))
	}

	setOAuthStateCookie(cc, state)

	return cc.Redirect(http.StatusFound, r.authService.DiscordAuthURL(state))
}

// discordLinkHandler godoc
// @Summary      Link Discord to the current account
// @Description  Begin Discord OAuth to attach a Discord account to the logged-in user
// @Tags         auth
// @Success      302
// @Router       /dash/auth/discord/link [get]
func (r *handler) discordLinkHandler(c echo.Context) error {
	cc := c.(*appctx.Context)

	state, err := crypt.GenerateSessionID()
	if err != nil {
		return cc.JSON(http.StatusInternalServerError, httputil.ErrorResponse("Failed to start linking"))
	}

	setOAuthStateCookie(cc, state)
	// Mark this exchange as a link (not a login) so the shared callback attaches
	// the Discord account to the current session instead of logging in.
	cc.SetCookie(&http.Cookie{
		Name:     oauthModeCookie,
		Value:    oauthModeLink,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	return cc.Redirect(http.StatusFound, r.authService.DiscordAuthURL(state))
}

// discordCallbackHandler godoc
// @Summary      Discord OAuth callback
// @Description  Handle the Discord redirect, create a session, and return to the dashboard
// @Tags         auth
// @Success      302
// @Failure      400  {object}  httputil.ErrorResponseData
// @Failure      401  {object}  httputil.ErrorResponseData
// @Router       /dash/auth/discord/callback [get]
func (r *handler) discordCallbackHandler(c echo.Context) error {
	cc := c.(*appctx.Context)
	ctx := cc.Request().Context()

	code := cc.QueryParam("code")
	state := cc.QueryParam("state")
	if code == "" || state == "" {
		return cc.JSON(http.StatusBadRequest, httputil.ErrorResponse("Missing code or state"))
	}

	stateCookie, err := cc.Cookie(discordStateCookie)
	if err != nil || stateCookie.Value == "" || stateCookie.Value != state {
		return cc.JSON(http.StatusBadRequest, httputil.ErrorResponse("Invalid OAuth state"))
	}
	cc.SetCookie(&http.Cookie{Name: discordStateCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true})

	// Linking flow: attach Discord to the already-logged-in account.
	if modeCookie, mErr := cc.Cookie(oauthModeCookie); mErr == nil && modeCookie.Value == oauthModeLink {
		return r.discordLinkCallback(cc, ctx, code)
	}

	sessionID, userID, discordID, err := r.authService.LoginWithDiscord(ctx, code)
	if err != nil {
		return cc.JSON(http.StatusUnauthorized, httputil.ErrorResponse("Discord login failed"))
	}

	csrfToken, err := crypt.GenerateSessionID()
	if err != nil {
		return cc.JSON(http.StatusInternalServerError, httputil.ErrorResponse("Failed to create csrf token"))
	}

	cc.SetCookie(r.authService.CreateSessionCookie(sessionID))
	cc.SetCookie(r.authService.CreateCSRFCookie(csrfToken))

	// If the user arrived via an invite link, redeem it now that they have a session.
	if inviteToken := inviteTokenFromCookie(cc); inviteToken != "" {
		cc.SetCookie(&http.Cookie{Name: inviteCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true})

		err := r.inviteService.Accept(ctx, inviteToken, userID, discordID)
		if errors.Is(err, invite.ErrInviteWrongDiscord) {
			return cc.Redirect(http.StatusFound, "/invite/"+inviteToken+"?error=wrong_discord")
		}
	}

	return cc.Redirect(http.StatusFound, "/")
}

// discordLinkCallback attaches the authorized Discord account to the currently
// logged-in user, then redirects back to the app with a status flag.
func (r *handler) discordLinkCallback(cc *appctx.Context, ctx context.Context, code string) error {
	cc.SetCookie(&http.Cookie{Name: oauthModeCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true})

	sessionCookie, err := cc.Cookie(internalauth.SessionCookieName)
	if err != nil || sessionCookie.Value == "" {
		return cc.Redirect(http.StatusFound, "/auth")
	}
	user, err := r.authService.UserBySession(ctx, sessionCookie.Value)
	if err != nil {
		return cc.Redirect(http.StatusFound, "/auth")
	}

	err = r.authService.LinkDiscord(ctx, code, user.ID)
	if errors.Is(err, auth.ErrDiscordAlreadyLinked) {
		return cc.Redirect(http.StatusFound, "/app?discord=taken")
	}
	if err != nil {
		return cc.Redirect(http.StatusFound, "/app?discord=error")
	}

	return cc.Redirect(http.StatusFound, "/app?discord=linked")
}

// inviteTokenFromCookie returns the pending invite token, if any.
func inviteTokenFromCookie(cc *appctx.Context) string {
	cookie, err := cc.Cookie(inviteCookie)
	if err != nil || cookie.Value == "" {
		return ""
	}
	return cookie.Value
}

// updateProfileHandler godoc
// @Summary      Update profile
// @Description  Update the current user's display name and/or avatar
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        data  body      api.UpdateProfileRequest  true  "Update Profile Request"
// @Success      200   {object}  httputil.ResponseData{data=api.User}
// @Failure      400   {object}  httputil.ErrorResponseData
// @Router       /dash/auth/me [patch]
func (r *handler) updateProfileHandler(c echo.Context) error {
	cc := c.(*appctx.Context)
	ctx := cc.Request().Context()

	var req api.UpdateProfileRequest
	if err := validator.BindAndValidate(cc, &req); err != nil {
		return cc.JSON(http.StatusBadRequest, httputil.ErrorResponse(err.Error()))
	}

	user := cc.User()
	if err := r.authService.UpdateProfile(ctx, user.ID, req.Name, req.Avatar); err != nil {
		return cc.JSON(http.StatusInternalServerError, httputil.ErrorResponse("Failed to update profile"))
	}

	return cc.JSON(http.StatusOK, httputil.Response("Profile updated"))
}

// changePasswordHandler godoc
// @Summary      Change password
// @Description  Change the current user's password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        data  body      api.ChangePasswordRequest  true  "Change Password Request"
// @Success      200   {object}  httputil.ResponseData{data=string}
// @Failure      400   {object}  httputil.ErrorResponseData
// @Failure      403   {object}  httputil.ErrorResponseData
// @Router       /dash/auth/password [post]
func (r *handler) changePasswordHandler(c echo.Context) error {
	cc := c.(*appctx.Context)
	ctx := cc.Request().Context()

	var req api.ChangePasswordRequest
	if err := validator.BindAndValidate(cc, &req); err != nil {
		return cc.JSON(http.StatusBadRequest, httputil.ErrorResponse(err.Error()))
	}

	user := cc.User()
	err := r.authService.ChangePassword(ctx, user.ID, req.CurrentPassword, req.NewPassword)
	if errors.Is(err, auth.ErrUserCredentials{}) {
		return cc.JSON(http.StatusForbidden, httputil.ErrorResponse("Current password is incorrect"))
	}
	if err != nil {
		return cc.JSON(http.StatusInternalServerError, httputil.ErrorResponse("Failed to change password"))
	}

	return cc.JSON(http.StatusOK, httputil.Response("Password changed"))
}

// unlinkDiscordHandler godoc
// @Summary      Unlink Discord
// @Description  Remove the Discord link from the current account
// @Tags         auth
// @Produce      json
// @Success      200  {object}  httputil.ResponseData{data=string}
// @Failure      400  {object}  httputil.ErrorResponseData
// @Router       /dash/auth/discord/unlink [post]
func (r *handler) unlinkDiscordHandler(c echo.Context) error {
	cc := c.(*appctx.Context)
	ctx := cc.Request().Context()

	user := cc.User()
	err := r.authService.UnlinkDiscord(ctx, user.ID)
	if errors.Is(err, auth.ErrUnlinkWouldLockOut) {
		return cc.JSON(http.StatusBadRequest, httputil.ErrorResponse("Set a password before unlinking Discord"))
	}
	if err != nil {
		return cc.JSON(http.StatusInternalServerError, httputil.ErrorResponse("Failed to unlink Discord"))
	}

	return cc.JSON(http.StatusOK, httputil.Response("Discord unlinked"))
}

// logoutHandler godoc
// @Summary      Logout
// @Description  Logout current user and clear session cookie
// @Tags         auth
// @Produce      json
// @Success      200  {object}  httputil.ResponseData{data=string}
// @Failure      500  {object}  httputil.ErrorResponseData
// @Router       /dash/auth/logout [post]
func (r *handler) logoutHandler(c echo.Context) error {
	cc := c.(*appctx.Context)
	ctx := cc.Request().Context()

	sessionCookie, err := cc.Cookie(internalauth.SessionCookieName)
	if err != nil {
		// better error handling here
		return cc.JSON(http.StatusOK, httputil.Response("Already logged out"))
	}

	err = r.authService.LogoutUser(ctx, sessionCookie.Value)
	if err != nil {
		// here too
		return cc.JSON(http.StatusInternalServerError, httputil.ErrorResponse("Failed to logout"))
	}

	logoutCookie := r.authService.CreateLogoutCookie()
	cc.SetCookie(logoutCookie)
	cc.SetCookie(&http.Cookie{
		Name:     internalauth.CSRFCookieName,
		Value:    "",
		HttpOnly: false,
		Path:     "/",
		Expires:  logoutCookie.Expires,
		MaxAge:   -1,
	})

	// change this dumb ahhh response
	return cc.JSON(http.StatusOK, httputil.Response("Logout successful"))
}
