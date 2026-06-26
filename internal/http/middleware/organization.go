package middleware

import (
	"errors"

	internalauth "github.com/fivemanage/lite/internal/auth"
	"github.com/fivemanage/lite/internal/http/appctx"
)

func resolveOrgID(cc *appctx.Context) (string, error) {
	orgID := cc.Param("organizationId")
	if orgID == "" {
		orgID = cc.Param("id")
	}
	if orgID == "" {
		if contextOrgID, ok := cc.Get(internalauth.OrgIDContextKey).(string); ok {
			orgID = contextOrgID
		}
	}
	if orgID == "" {
		return "", errors.New("organization context required")
	}
	return orgID, nil
}
