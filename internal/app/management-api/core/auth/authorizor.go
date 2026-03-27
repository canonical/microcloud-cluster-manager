package auth

import (
	"context"
	"fmt"

	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
)

type ManagementAPIAuthorizor struct{}

func NewManagementAPIAuthorizor() (*ManagementAPIAuthorizor, error) {
	return &ManagementAPIAuthorizor{}, nil
}

// CheckPermissions checks if the user has permissions to perform an action based on the user information in the context.
// It returns an error if the user does not have permissions or if there was an error getting the user information from the context.
func (a *ManagementAPIAuthorizor) CheckPermissions(ctx context.Context, AllowedEntitlements []string) error {
	userInfo, ok := ctx.Value(types.UserInfoKey).(*types.UserInfo)
	if !ok {
		return fmt.Errorf("failed to get user information from the request context")
	}
	if !userInfo.IsAdmin {
		return fmt.Errorf("user does not have permissions to perform this action")
	}
	return nil
}
