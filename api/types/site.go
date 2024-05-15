// Package types provides shared types and structs.
package types

// Site is a standalone or clustered LXD site.
type Site struct {
	Name      string
	Addresses []string
	Status    string
}
