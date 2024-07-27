package database

import (
	"time"
)

//go:generate -command mapper lxd-generate db mapper -t remote_cluster_details.mapper.go
//go:generate mapper reset
//
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e remote_cluster_detail objects table=remote_cluster_details
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e remote_cluster_detail objects-by-CoreRemoteClusterID table=remote_cluster_details
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e remote_cluster_detail id table=remote_cluster_details
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e remote_cluster_detail create table=remote_cluster_details
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e remote_cluster_detail update table=remote_cluster_details
//
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e remote_cluster_detail GetMany
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e remote_cluster_detail GetOne
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e remote_cluster_detail ID
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e remote_cluster_detail Exists
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e remote_cluster_detail Create
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e remote_cluster_detail Update

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

// RemoteClusterDetailFilter is a required struct for use with lxd-generate. It is used for filtering fields on database fetches.
type RemoteClusterDetailFilter struct {
	CoreRemoteClusterID *int64
}
