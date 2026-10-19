package tunnel

import (
	"encoding/json"
	"net/http"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
)

// ResponseError maps an upstream failure to a client status and forwards only the JSON error message.
func ResponseError(resp types.ClusterManagerTunnelResponse, clusterName string, operation string) error {
	status := http.StatusBadGateway
	switch resp.Status {
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound,
		http.StatusConflict, http.StatusUnprocessableEntity, http.StatusTooManyRequests, http.StatusGatewayTimeout:
		status = resp.Status
	}

	message := upstreamErrorMessage(resp.Body)
	if message == "" {
		return api.StatusErrorf(status, "Failed to %s on cluster %q (upstream HTTP status %d)", operation, clusterName, resp.Status)
	}
	return api.StatusErrorf(status, "Failed to %s on cluster %q (upstream HTTP status %d): %s", operation, clusterName, resp.Status, message)
}

// upstreamErrorMessage returns the error field of a JSON body, or an empty string for non-JSON bodies.
func upstreamErrorMessage(body []byte) string {
	var upstream struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &upstream); err != nil {
		return ""
	}

	return upstream.Error
}
