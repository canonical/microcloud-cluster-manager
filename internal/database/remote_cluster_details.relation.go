package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/canonical/lxd/lxd/db/query"
	"github.com/canonical/microcluster/cluster"
)

// CoreRemoteClusterWithDetail is a struct that contains all the information about a remote cluster directly queried from the database.
type CoreRemoteClusterWithDetail struct {
	ID                 int64     `db:"id"`
	Name               string    `db:"name"`
	ClusterCertificate string    `db:"cluster_certificate"`
	ClusterCreatedAt   time.Time `db:"created_at"`
	Status             string    `db:"status"`
	CPUTotalCount      int64     `db:"cpu_total_count"`
	CPULoad1           string    `db:"cpu_load_1"`
	CPULoad5           string    `db:"cpu_load_5"`
	CPULoad15          string    `db:"cpu_load_15"`
	MemoryTotalAmount  int64     `db:"memory_total_amount"`
	MemoryUsage        int64     `db:"memory_usage"`
	DiskTotalSize      int64     `db:"disk_total_size"`
	DiskUsage          int64     `db:"disk_usage"`
	InstanceCount      int64     `db:"instance_count"`
	InstanceStatuses   string    `db:"instance_statuses"`
	MemberCount        int64     `db:"member_count"`
	MemberStatuses     string    `db:"member_statuses"`
	ClusterJoinedAt    time.Time `db:"joined_at"`
	ClusterUpdatedAt   time.Time `db:"updated_at"`
}

func mainCoreRemoteClusterDetailQuery() string {
	return `
		SELECT
			core_remote_clusters.id, core_remote_clusters.name, core_remote_clusters.cluster_certificate, core_remote_clusters.created_at,
			remote_cluster_details.status, remote_cluster_details.cpu_total_count, remote_cluster_details.cpu_load_1, remote_cluster_details.cpu_load_5, remote_cluster_details.cpu_load_15, remote_cluster_details.memory_total_amount, remote_cluster_details.memory_usage, 
			remote_cluster_details.disk_total_size, remote_cluster_details.disk_usage, remote_cluster_details.instance_count, remote_cluster_details.instance_statuses, 
			remote_cluster_details.member_count, remote_cluster_details.member_statuses, remote_cluster_details.joined_at, remote_cluster_details.updated_at
		FROM remote_cluster_details
		JOIN core_remote_clusters ON remote_cluster_details.core_remote_cluster_id = core_remote_clusters.id
	`
}

var coreRemoteClusterDetailObjects = cluster.RegisterStmt(
	fmt.Sprintf(`%s ORDER BY core_remote_clusters.name`, mainCoreRemoteClusterDetailQuery()),
)

var coreRemoteClusterDetailByNameObjects = cluster.RegisterStmt(
	fmt.Sprintf(`%s WHERE core_remote_clusters.name = ?`, mainCoreRemoteClusterDetailQuery()),
)

// GetCoreRemoteClustersWithDetails fetches all remote cluster details with core remote cluster information from the database.
func GetCoreRemoteClustersWithDetails(ctx context.Context, tx *sql.Tx) ([]CoreRemoteClusterWithDetail, error) {
	var err error
	objects := make([]CoreRemoteClusterWithDetail, 0)
	sqlStmt, err := cluster.Stmt(tx, coreRemoteClusterDetailObjects)
	if err != nil {
		return nil, fmt.Errorf("Failed to prepare statement: %w", err)
	}

	dest := func(scan func(dest ...any) error) error {
		c := CoreRemoteClusterWithDetail{}
		err := scan(
			&c.ID,
			&c.Name,
			&c.ClusterCertificate,
			&c.ClusterCreatedAt,
			&c.Status,
			&c.CPUTotalCount,
			&c.CPULoad1,
			&c.CPULoad5,
			&c.CPULoad15,
			&c.MemoryTotalAmount,
			&c.MemoryUsage,
			&c.DiskTotalSize,
			&c.DiskUsage,
			&c.InstanceCount,
			&c.InstanceStatuses,
			&c.MemberCount,
			&c.MemberStatuses,
			&c.ClusterJoinedAt,
			&c.ClusterUpdatedAt,
		)

		if err != nil {
			return err
		}

		objects = append(objects, c)

		return nil
	}

	err = query.SelectObjects(ctx, sqlStmt, dest)
	if err != nil {
		return nil, fmt.Errorf("Failed to do a joint fetch from \"core_remote_clusters\" and \"remote_cluster\" tables: %w", err)
	}

	return objects, nil
}

// GetCoreRemoteClusterWithDetailByName fetches the remote cluster detail with core remote cluster information from the database filtered by remote cluster name.
func GetCoreRemoteClusterWithDetailByName(ctx context.Context, tx *sql.Tx, remoteClusterName string) ([]CoreRemoteClusterWithDetail, error) {
	var err error
	objects := make([]CoreRemoteClusterWithDetail, 0)
	sqlStmt, err := cluster.Stmt(tx, coreRemoteClusterDetailByNameObjects)
	if err != nil {
		return nil, fmt.Errorf("Failed to prepare statement: %w", err)
	}

	dest := func(scan func(dest ...any) error) error {
		c := CoreRemoteClusterWithDetail{}
		err := scan(
			&c.ID,
			&c.Name,
			&c.ClusterCertificate,
			&c.ClusterCreatedAt,
			&c.Status,
			&c.CPUTotalCount,
			&c.CPULoad1,
			&c.CPULoad5,
			&c.CPULoad15,
			&c.MemoryTotalAmount,
			&c.MemoryUsage,
			&c.DiskTotalSize,
			&c.DiskUsage,
			&c.InstanceCount,
			&c.InstanceStatuses,
			&c.MemberCount,
			&c.MemberStatuses,
			&c.ClusterJoinedAt,
			&c.ClusterUpdatedAt,
		)

		if err != nil {
			return err
		}

		objects = append(objects, c)

		return nil
	}

	err = query.SelectObjects(ctx, sqlStmt, dest, remoteClusterName)
	if err != nil {
		return nil, fmt.Errorf("Failed to do a joint fetch from \"core_remote_clusters\" and \"remote_cluster\" tables: %w", err)
	}

	return objects, nil
}
