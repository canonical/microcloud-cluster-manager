package database

import (
	"encoding/json"
	"time"

	"github.com/canonical/lxd-cluster-manager/internal/api/types"
)

//go:generate -command mapper lxd-generate db mapper -t remote_cluster_details.mapper.go
//go:generate mapper reset
//
//go:generate mapper stmt -d github.com/canonical/microcluster/v2/cluster -e remote_cluster_detail objects table=remote_cluster_details
//go:generate mapper stmt -d github.com/canonical/microcluster/v2/cluster -e remote_cluster_detail objects-by-CoreRemoteClusterID table=remote_cluster_details
//go:generate mapper stmt -d github.com/canonical/microcluster/v2/cluster -e remote_cluster_detail id table=remote_cluster_details
//go:generate mapper stmt -d github.com/canonical/microcluster/v2/cluster -e remote_cluster_detail create table=remote_cluster_details
//go:generate mapper stmt -d github.com/canonical/microcluster/v2/cluster -e remote_cluster_detail update table=remote_cluster_details
//
//go:generate mapper method -i -d github.com/canonical/microcluster/v2/cluster -e remote_cluster_detail GetMany
//go:generate mapper method -i -d github.com/canonical/microcluster/v2/cluster -e remote_cluster_detail GetOne
//go:generate mapper method -i -d github.com/canonical/microcluster/v2/cluster -e remote_cluster_detail ID
//go:generate mapper method -i -d github.com/canonical/microcluster/v2/cluster -e remote_cluster_detail Exists
//go:generate mapper method -i -d github.com/canonical/microcluster/v2/cluster -e remote_cluster_detail Create
//go:generate mapper method -i -d github.com/canonical/microcluster/v2/cluster -e remote_cluster_detail Update

// RemoteClusterDetail represents all remote_cluster level data.
type RemoteClusterDetail struct {
	CoreRemoteClusterID int64 `db:"primary=true"`
	Status              string
	ID                  int
	CPUTotalCount       int64
	CPULoad1            string `db:"sql=remote_cluster_details.cpu_load_1"`
	CPULoad5            string `db:"sql=remote_cluster_details.cpu_load_5"`
	CPULoad15           string `db:"sql=remote_cluster_details.cpu_load_15"`
	MemoryTotalAmount   int64
	MemoryUsage         int64
	DiskTotalSize       int64
	DiskUsage           int64
	InstanceCount       int64
	InstanceStatuses    string
	MemberCount         int64
	MemberStatuses      string
	JoinedAt            time.Time
	UpdatedAt           time.Time
}

// Put updates the RemoteClusterDetail with the provided payload.
func (r *RemoteClusterDetail) Put(payload types.RemoteClusterStatusPost) {
	r.CPULoad1 = payload.CPULoad1
	r.CPULoad5 = payload.CPULoad5
	r.CPULoad15 = payload.CPULoad15
	r.CPUTotalCount = payload.CPUTotalCount
	r.DiskTotalSize = payload.DiskTotalSize
	r.DiskUsage = payload.DiskUsage
	r.InstanceCount, r.InstanceStatuses = parseStatusDistribution(payload.InstanceStatuses)
	r.MemberCount, r.MemberStatuses = parseStatusDistribution(payload.MemberStatuses)
	r.MemoryTotalAmount = payload.MemoryTotalAmount
	r.MemoryUsage = payload.MemoryUsage
	r.UpdatedAt = time.Now()
}

// RemoteClusterDetailFilter is a required struct for use with lxd-generate. It is used for filtering fields on database fetches.
type RemoteClusterDetailFilter struct {
	CoreRemoteClusterID *int64
}

func parseStatusDistribution(statuses []types.StatusDistribution) (int64, string) {
	if len(statuses) == 0 {
		return 0, "[]"
	}

	parsedStatuses, err := json.Marshal(statuses)
	if err != nil {
		return 0, "[]"
	}

	var total int64
	for _, s := range statuses {
		total += s.Count
	}

	return total, string(parsedStatuses)
}
