package auth

import (
	"crypto/hmac"
	"net/http"

	"github.com/canonical/lxd-cluster-manager/internal/pkg/api/models/v1"
	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/lxd/shared/trust"
)

const HMACClusterManager10 trust.HMACVersion = "ClusterManager-1.0"

// VerifyHMAC verifies the HMAC signature of a request.
func VerifyHMAC(payload models.RemoteClusterPost, r *http.Request, secret string) (bool, error) {
	h := trust.NewHMAC([]byte(secret), trust.NewDefaultHMACConf(HMACClusterManager10))

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false, api.NewStatusError(http.StatusBadRequest, "Authorization header is missing")
	}
	hFromHeader, hmacFromHeader, err := h.ParseHTTPHeader(authHeader)
	if err != nil {
		return false, api.StatusErrorf(http.StatusBadRequest, "Failed to parse Authorization header: %w", err)
	}

	hmacFromBody, err := hFromHeader.WriteJSON(payload)
	if err != nil {
		return false, api.StatusErrorf(http.StatusBadRequest, "Failed to calculate HMAC from payload: %w", err)
	}

	return hmac.Equal(hmacFromHeader, hmacFromBody), nil
}
