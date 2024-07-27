package types

import "time"

// RemoteClusterStatus represents the status of a remote cluster.
type RemoteClusterStatus string

const (
	// PENDING_APPROVAL is the status of a remote cluster that is waiting for approval. A join reuqest would have been received for the remote cluster.
	PENDING_APPROVAL RemoteClusterStatus = "PENDING_APPROVAL"
	// ACTIVE is the status of a remote cluster once its join request has been approved.
	ACTIVE RemoteClusterStatus = "ACTIVE"
)

// RemoteCluster is a standalone or clustered LXD cluster.
type RemoteCluster struct {
	Name               string               `json:"name"`
	ClusterCertificate string               `json:"cluster_certificate"`
	Status             string               `json:"status"`
	CPUTotalCount      int64                `json:"cpu_total_count"`
	CPULoad1           string               `json:"cpu_load_1"`
	CPULoad5           string               `json:"cpu_load_5"`
	CPULoad15          string               `json:"cpu_load_15"`
	MemoryTotalAmount  int64                `json:"memory_total_amount"`
	MemoryUsage        int64                `json:"memory_usage"`
	DiskTotalSize      int64                `json:"disk_total_size"`
	DiskUsage          int64                `json:"disk_usage"`
	MemberCount        int64                `json:"member_count"`
	MemberStatuses     []StatusDistribution `json:"member_statuses"`
	InstanceCount      int64                `json:"instance_count"`
	InstanceStatuses   []StatusDistribution `json:"instance_statuses"`
	JoinedAt           time.Time            `json:"joined_at"`
	CreatedAt          time.Time            `json:"created_at"`
	LastStatusUpdateAt time.Time            `json:"last_status_update_at"`
}

// RemoteClusterPatch represents the payload for the PATCH /1.0/remote-clusters/{remoteClusterName} endpoint.
type RemoteClusterPatch struct {
	Status RemoteClusterStatus `json:"status"`
}
