package middleware

import (
	"errors"

	internalauth "github.com/fivemanage/lite/internal/auth"
	"github.com/fivemanage/lite/internal/http/appctx"
)

var errOrganizationContextRequired = errors.New("organization context required")

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
		return "", errOrganizationContextRequired
	}
	return orgID, nil
}
