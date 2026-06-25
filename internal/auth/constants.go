package auth

import "time"

const (
	SessionCookieName = "fmlite_session"
	CSRFCookieName    = "fmlite_csrf"
	CSRFHeaderName    = "X-CSRF-Token"
	UserContextKey    = "user"
	OrgIDContextKey   = "org_id"
	SessionDuration   = 24 * time.Hour
)
