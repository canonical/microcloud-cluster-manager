// Package types provides shared types and structs.
package types

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

// ExternalSiteTokenPost is the request body for creating a new external site token.
type ExternalSiteTokenPost struct {
	Expiry   time.Time `json:"expiry"`
	SiteName string    `json:"site_name"`
}

// ExternalSiteTokenPostResponse is the response body for creating a new external site token.
type ExternalSiteTokenPostResponse struct {
	Token string `json:"token"`
}

// ExternalSiteToken is the response body for an external site token.
type ExternalSiteToken struct {
	Expiry   time.Time `json:"expiry"`
	SiteName string    `json:"site_name"`
	CreateAt time.Time `json:"created_at"`
}

// ExternalSiteTokenBody is the body of the external site token.
type ExternalSiteTokenBody struct {
	Secret      string    `json:"secret"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	Addresses   []string  `json:"addresses"`
	ServerName  string    `json:"server_name"`
	Fingerprint string    `json:"fingerprint"`
}

// Encode returns the base64 encoded string of the token body.
func (t ExternalSiteTokenBody) Encode() (string, error) {
	tokenData, err := json.Marshal(t)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(tokenData), nil
}
