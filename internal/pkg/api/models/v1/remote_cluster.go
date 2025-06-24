package models

import "time"

// RemoteClusterStatus represents the status of a remote cluster.
type RemoteClusterStatus string

const (
	// ACTIVE is the status of a remote cluster once its join request has been approved.
	ACTIVE RemoteClusterStatus = "ACTIVE"
)

// StatusDistribution is a status and count pair used by the remote cluster status endpoint.
type StatusDistribution struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

// RemoteCluster is a standalone or clustered LXD cluster.
type RemoteCluster struct {
	Name               string               `json:"name"`
	Description        string               `json:"description"`
	ClusterCertificate string               `json:"cluster_certificate"`
	DiskThreshold      int                  `json:"disk_threshold"`
	MemoryThreshold    int                  `json:"memory_threshold"`
	Status             string               `json:"status"`
	CPUTotalCount      int                  `json:"cpu_total_count"`
	CPULoad1           string               `json:"cpu_load_1"`
	CPULoad5           string               `json:"cpu_load_5"`
	CPULoad15          string               `json:"cpu_load_15"`
	MemoryTotalAmount  int                  `json:"memory_total_amount"`
	MemoryUsage        int                  `json:"memory_usage"`
	DiskTotalSize      int                  `json:"disk_total_size"`
	DiskUsage          int                  `json:"disk_usage"`
	MemberCount        int                  `json:"member_count"`
	MemberStatuses     []StatusDistribution `json:"member_statuses"`
	InstanceCount      int                  `json:"instance_count"`
	InstanceStatuses   []StatusDistribution `json:"instance_statuses"`
	UIURL              string               `json:"ui_url"`
	JoinedAt           time.Time            `json:"joined_at"`
	CreatedAt          time.Time            `json:"created_at"`
	LastStatusUpdateAt time.Time            `json:"last_status_update_at"`
}

// RemoteClusterPatch represents the payload for the PATCH /1.0/remote-clusters/{remoteClusterName} endpoint.
type RemoteClusterPatch struct {
	Status          RemoteClusterStatus `json:"status"`
	Description     string              `json:"description,omitempty"`
	DiskThreshold   int                 `json:"disk_threshold,omitempty"`
	MemoryThreshold int                 `json:"memory_threshold,omitempty"`
}

// RemoteClusterPost represents the fields required to create a new cluster.
type RemoteClusterPost struct {
	ClusterName        string `json:"cluster_name"`
	ClusterCertificate string `json:"cluster_certificate"`
	Token              string `json:"token" yaml:"token"`
}

// RemoteClusterStatusPost is sent by LXD with to inform about its current status.
type RemoteClusterStatusPost struct {
	CPUTotalCount     int                  `json:"cpu_total_count"`
	CPULoad1          string               `json:"cpu_load_1"`
	CPULoad5          string               `json:"cpu_load_5"`
	CPULoad15         string               `json:"cpu_load_15"`
	MemoryTotalAmount int                  `json:"memory_total_amount"`
	MemoryUsage       int                  `json:"memory_usage"`
	DiskTotalSize     int                  `json:"disk_total_size"`
	DiskUsage         int                  `json:"disk_usage"`
	MemberStatuses    []StatusDistribution `json:"member_statuses"`
	InstanceStatuses  []StatusDistribution `json:"instance_statuses"`
	Metrics           string               `json:"metrics"`
	UIURL             string               `json:"ui_url"`
}

// RemoteClusterStatusPostResponse is sent to LXD in response to a remote cluster status update.
type RemoteClusterStatusPostResponse struct {
	NextUpdateInSeconds   string `json:"next_update_in_seconds"`
	ClusterManagerAddress string `json:"cluster_manager_address"`
}
