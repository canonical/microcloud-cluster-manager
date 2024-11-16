package models

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

// RemoteClusterTokenPost is the request body for creating a new remote cluster token.
type RemoteClusterTokenPost struct {
	Expiry      time.Time `json:"expiry"`
	ClusterName string    `json:"cluster_name"`
}

// RemoteClusterTokenPostResponse is the response body for creating a new remote cluster token.
type RemoteClusterTokenPostResponse struct {
	Token string `json:"token"`
}

// RemoteClusterToken is the response body for an remote cluster token.
type RemoteClusterToken struct {
	Expiry      time.Time `json:"expiry"`
	ClusterName string    `json:"cluster_name"`
	CreateAt    time.Time `json:"created_at"`
}

// RemoteClusterTokenBody is the body of the remote cluster token.
type RemoteClusterTokenBody struct {
	Secret      string    `json:"secret"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	Address     string    `json:"address"`
	ServerName  string    `json:"server_name"`
	Fingerprint string    `json:"fingerprint"`
}

// Encode returns the base64 encoded string of the token body.
func (t RemoteClusterTokenBody) Encode() (string, error) {
	tokenData, err := json.Marshal(t)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(tokenData), nil
}
