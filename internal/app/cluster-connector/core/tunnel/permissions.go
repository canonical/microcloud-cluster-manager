// Package tunnel provides shared reverse-tunnel permission helpers.
package tunnel

import (
	"github.com/canonical/lxd/lxd/auth"
	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/lxd/shared/entity"
)

// AccessManagementAPIExtension adds the endpoint for retrieving effective permissions.
const AccessManagementAPIExtension = "access_management"

// HasPermission reports whether LXD's effective permissions grant the requested permission.
// A server admin entitlement grants full access, matching LXD's authorization semantics.
func HasPermission(effectivePermissions []api.Permission, requiredPermission api.Permission) bool {
	for _, effectivePermission := range effectivePermissions {
		if effectivePermission.EntityType == string(entity.TypeServer) &&
			effectivePermission.EntityReference == entity.ServerURL().String() &&
			effectivePermission.Entitlement == string(auth.EntitlementAdmin) {
			return true
		}
	}

	for _, effectivePermission := range effectivePermissions {
		if effectivePermission == requiredPermission {
			return true
		}
	}

	return false
}
