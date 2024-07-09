// Package types provides shared types and structs.
package types

import "time"

// Status is a simple struct that contains a status and a count.
type Status struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

// Site is a standalone or clustered LXD site.
type Site struct {
	Name               string    `json:"name"`
	SiteCertificate    string    `json:"site_certificate"`
	Status             string    `json:"status"`
	CPUTotalCount      float64   `json:"cpu_total_count"`
	CPULoad1           string    `json:"cpu_load_1"`
	CPULoad5           string    `json:"cpu_load_5"`
	CPULoad15          string    `json:"cpu_load_15"`
	MemoryTotalAmount  float64   `json:"memory_total_amount"`
	MemoryUsage        float64   `json:"memory_usage"`
	DiskTotalSize      float64   `json:"disk_total_size"`
	DiskUsage          float64   `json:"disk_usage"`
	MemberCount        int       `json:"member_count"`
	MemberStatuses     []Status  `json:"member_statuses"`
	InstanceCount      int       `json:"instance_count"`
	InstanceStatuses   []Status  `json:"instance_statuses"`
	JoinedAt           time.Time `json:"joined_at"`
	CreatedAt          time.Time `json:"created_at"`
	LastStatusUpdateAt time.Time `json:"last_status_update_at"`
}

// SitePost represents the fields required to create a new site.
type SitePost struct {
	SiteName        string `json:"site_name"`
	SiteCertificate string `json:"site_certificate"`
}

// SiteStatus represents the status of a site.
type SiteStatus string

const (
	// PENDING_APPROVAL is the status of a site that is waiting for approval. A join reuqest would have been received for the site.
	PENDING_APPROVAL SiteStatus = "PENDING_APPROVAL"
	// ACTIVE is the status of a site once its join request has been approved.
	ACTIVE SiteStatus = "ACTIVE"
)
