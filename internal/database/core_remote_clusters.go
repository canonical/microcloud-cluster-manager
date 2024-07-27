package database

import (
	"time"
)

//go:generate -command mapper lxd-generate db mapper -t core_remote_clusters.mapper.go
//go:generate mapper reset
//
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e core_remote_cluster objects table=core_remote_clusters
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e core_remote_cluster objects-by-Name table=core_remote_clusters
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e core_remote_cluster id table=core_remote_clusters
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e core_remote_cluster create table=core_remote_clusters
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e core_remote_cluster delete-by-Name table=core_remote_clusters
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e core_remote_cluster update table=core_remote_clusters
//
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_remote_cluster GetMany
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_remote_cluster GetOne
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_remote_cluster ID
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_remote_cluster Exists
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_remote_cluster Create
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_remote_cluster DeleteOne-by-Name
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_remote_cluster Update

// CoreRemoteCluster represents data about a remote cluster that would be common across all microcluster based projects.
type CoreRemoteCluster struct {
	Name               string `db:"primary=yes"`
	ClusterCertificate string
	ID                 int64
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// CoreRemoteClusterFilter is a required struct for use with lxd-generate. It is used for filtering fields on database fetches.
type CoreRemoteClusterFilter struct {
	Name *string
}
