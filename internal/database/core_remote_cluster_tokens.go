package database

import (
	"time"
)

//go:generate -command mapper lxd-generate db mapper -t core_remote_cluster_tokens.mapper.go
//go:generate mapper reset
//
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e core_remote_cluster_token objects table=core_remote_cluster_tokens
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e core_remote_cluster_token objects-by-ClusterName table=core_remote_cluster_tokens
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e core_remote_cluster_token id table=core_remote_cluster_tokens
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e core_remote_cluster_token create table=core_remote_cluster_tokens
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e core_remote_cluster_token delete-by-ClusterName table=core_remote_cluster_tokens
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e core_remote_cluster_token update table=core_remote_cluster_tokens
//
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_remote_cluster_token GetMany
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_remote_cluster_token GetOne
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_remote_cluster_token ID
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_remote_cluster_token Exists
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_remote_cluster_token Create
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_remote_cluster_token DeleteOne-by-ClusterName
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_remote_cluster_token Update

// CoreRemoteClusterToken is the join token for a remote cluster.
type CoreRemoteClusterToken struct {
	ClusterName string `db:"primary=true"`
	ID          int
	Secret      string
	Expiry      time.Time
	CreatedAt   time.Time
}

// CoreRemoteClusterTokenFilter is a required struct for use with lxd-generate. It is used for filtering fields on database fetches.
type CoreRemoteClusterTokenFilter struct {
	ClusterName *string
}
