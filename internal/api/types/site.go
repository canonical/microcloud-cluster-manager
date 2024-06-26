// Package types provides shared types and structs.
package types

import "time"

// MemberStatus is the status of a member in a site.
type MemberStatus struct {
	Address      string  `json:"address"`
	Architecture string  `json:"architecture"`
	Role         string  `json:"role"`
	UsageCPU     float64 `json:"usage_cpu"`
	UsageMemory  float64 `json:"usage_memory"`
	UsageDisk    float64 `json:"usage_disk"`
	Status       string  `json:"status"`
}

// Site is a standalone or clustered LXD site.
type Site struct {
	Name               string         `json:"name"`
	SiteCertificate    string         `json:"site_certificate"`
	Status             string         `json:"status"`
	InstanceCount      int            `json:"instance_count"`
	InstanceStatuses   string         `json:"instance_statuses"`
	JoinedAt           time.Time      `json:"joined_at"`
	CreatedAt          time.Time      `json:"created_at"`
	LastStatusUpdateAt time.Time      `json:"last_status_update_at"`
	MemberStatuses     []MemberStatus `json:"member_statuses"`
}
