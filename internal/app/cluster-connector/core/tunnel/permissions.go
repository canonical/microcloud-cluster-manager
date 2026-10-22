// Package tunnel provides shared reverse-tunnel permission helpers.
package tunnel

import (
	"fmt"
	"slices"

	"github.com/canonical/lxd/lxd/auth"
	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/lxd/shared/entity"
)

// AccessManagementAPIExtension adds the endpoint for retrieving effective permissions.
const AccessManagementAPIExtension = "access_management"

// CheckPermissions returns an error if LXD's effective permissions do not grant every required permission.
// A server admin entitlement grants full access, matching LXD's authorization semantics.
func CheckPermissions(effectivePermissions []api.Permission, requiredPermissions []api.Permission) error {
	serverAdmin := api.Permission{
		EntityType:      string(entity.TypeServer),
		EntityReference: entity.ServerURL().String(),
		Entitlement:     string(auth.EntitlementAdmin),
	}

	if slices.Contains(effectivePermissions, serverAdmin) {
		return nil
	}

	for _, requiredPermission := range requiredPermissions {
		if !slices.Contains(effectivePermissions, requiredPermission) {
			return fmt.Errorf("missing %q entitlement on %q", requiredPermission.Entitlement, requiredPermission.EntityReference)
		}
	}

	return nil
}
