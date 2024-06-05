// Package types provides shared types and structs.
package types

// Site is a standalone or clustered LXD site.
type Site struct {
	Name      string   `json:"name"`
	Addresses []string `json:"addresses"`
	Status    string   `json:"status"`
}
