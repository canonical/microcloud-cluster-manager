package auth

import (
	"context"
)

type ClusterConnectorAuthorizor struct{}

func NewClusterConnectorAuthorizor() (*ClusterConnectorAuthorizor, error) {
	return &ClusterConnectorAuthorizor{}, nil
}

// Cluster connector does not have any authorization requirements at the moment, so this method is a no-op.
func (a *ClusterConnectorAuthorizor) CheckPermissions(ctx context.Context, AllowedEntitlements []string) error {
	return nil
}
