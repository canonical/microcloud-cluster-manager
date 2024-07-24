package types

// StatusDistribution is a status and count pair used by the site status endpoint.
type StatusDistribution struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

// SitePost represents the fields required to create a new site.
type SitePost struct {
	SiteName        string `json:"site_name"`
	SiteCertificate string `json:"site_certificate"`
}

// SiteStatusPost is sent by LXD with to inform about its current status.
type SiteStatusPost struct {
	CPUTotalCount     int64                `json:"cpu_total_count"`
	CPULoad1          string               `json:"cpu_load_1"`
	CPULoad5          string               `json:"cpu_load_5"`
	CPULoad15         string               `json:"cpu_load_15"`
	MemoryTotalAmount int64                `json:"memory_total_amount"`
	MemoryUsage       int64                `json:"memory_usage"`
	DiskTotalSize     int64                `json:"disk_total_size"`
	DiskUsage         int64                `json:"disk_usage"`
	MemberStatuses    []StatusDistribution `json:"member_statuses"`
	InstanceStatuses  []StatusDistribution `json:"instance_status"`
}

// SiteStatusPostResponse is sent to LXD in response to a site status update.
type SiteStatusPostResponse struct {
	SiteManagerAddresses []string `json:"site_manager_addresses"`
}
