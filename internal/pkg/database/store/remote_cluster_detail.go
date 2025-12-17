package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/api/models/v1"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database/query"
	"github.com/jmoiron/sqlx"
)

// RemoteClusterDetail represents detailed information about a remote LXD cluster.
type RemoteClusterDetail struct {
	ID                int             `db:"id"`                    // Primary key
	RemoteClusterID   int             `db:"remote_cluster_id"`     // Foreign key to remote_clusters
	CPUTotalCount     int64           `db:"cpu_total_count"`       // Total CPU count
	CPULoad1          string          `db:"cpu_load_1"`            // CPU load (1 minute average)
	CPULoad5          string          `db:"cpu_load_5"`            // CPU load (5 minute average)
	CPULoad15         string          `db:"cpu_load_15"`           // CPU load (15 minute average)
	MemoryTotalAmount int64           `db:"memory_total_amount"`   // Total memory in bytes
	MemoryUsage       int64           `db:"memory_usage"`          // Memory usage in bytes
	InstanceCount     int64           `db:"instance_count"`        // Number of instances
	InstanceStatuses  json.RawMessage `db:"instance_statuses"`     // JSON array of instance statuses
	MemberCount       int64           `db:"member_count"`          // Number of members
	MemberStatuses    json.RawMessage `db:"member_statuses"`       // JSON array of member statuses
	StoragePoolUsages json.RawMessage `json:"storage_pool_usages"` // JSON array of storage pool usages
	UIURL             string          `db:"ui_url"`                // UI URL
	CreatedAt         time.Time       `db:"created_at"`            // Creation timestamp
	UpdatedAt         time.Time       `db:"updated_at"`            // Update timestamp
}

// Put updates the RemoteClusterDetail with the provided payload.
func (r *RemoteClusterDetail) Put(payload models.RemoteClusterStatusPost) {
	r.CPULoad1 = payload.CPULoad1
	r.CPULoad5 = payload.CPULoad5
	r.CPULoad15 = payload.CPULoad15
	r.CPUTotalCount = payload.CPUTotalCount
	r.InstanceCount, r.InstanceStatuses = parseStatusDistribution(payload.InstanceStatuses)
	r.MemberCount, r.MemberStatuses = parseStatusDistribution(payload.MemberStatuses)
	r.MemoryTotalAmount = payload.MemoryTotalAmount
	r.MemoryUsage = payload.MemoryUsage
	r.StoragePoolUsages = parsePoolUsage(payload.StoragePoolUsages)
	r.UIURL = payload.UIURL
	r.UpdatedAt = time.Now()
}

// RemoteClusterWithDetail is a struct that contains all the information about a remote cluster directly queried from the database.
type RemoteClusterWithDetail struct {
	ID                 int             `db:"id"`
	Name               string          `db:"name"`
	Description        string          `db:"description"`
	ClusterCertificate string          `db:"cluster_certificate"`
	DiskThreshold      int64           `db:"disk_threshold"`
	MemoryThreshold    int64           `db:"memory_threshold"`
	ClusterCreatedAt   time.Time       `db:"created_at"`
	Status             string          `db:"status"`
	CPUTotalCount      int64           `db:"cpu_total_count"`
	CPULoad1           string          `db:"cpu_load_1"`
	CPULoad5           string          `db:"cpu_load_5"`
	CPULoad15          string          `db:"cpu_load_15"`
	MemoryTotalAmount  int64           `db:"memory_total_amount"`
	MemoryUsage        int64           `db:"memory_usage"`
	InstanceCount      int64           `db:"instance_count"`
	InstanceStatuses   json.RawMessage `db:"instance_statuses"`
	MemberCount        int64           `db:"member_count"`
	MemberStatuses     json.RawMessage `db:"member_statuses"`
	StoragePoolUsages  json.RawMessage `json:"storage_pool_usages"`
	UIURL              string          `db:"ui_url"`
	ClusterJoinedAt    time.Time       `db:"joined_at"`
	ClusterUpdatedAt   time.Time       `db:"updated_at"`
}

// GetRemoteClusterDetailID returns the ID of the detail entry for a remote cluster.
func GetRemoteClusterDetailID(ctx context.Context, tx *sqlx.Tx, remoteClusterID int) (int, error) {
	// Query to check if the entry exists
	q := `
        SELECT id
        FROM remote_cluster_details
        WHERE remote_cluster_id = $1
    `

	var id int
	err := tx.QueryRowContext(ctx, q, remoteClusterID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return -1, api.StatusErrorf(http.StatusNotFound, "detail for remote cluster not found")
	}

	if err != nil {
		return -1, fmt.Errorf("failed to get \"remote_cluster_details\" ID: %w", err)
	}

	return id, nil
}

// RemoteClusterDetailExists checks detail exists for a given remote cluster id.
func RemoteClusterDetailExists(ctx context.Context, tx *sqlx.Tx, remoteClusterID int) (bool, error) {
	_, err := GetRemoteClusterDetailID(ctx, tx, remoteClusterID)
	if err != nil {
		if api.StatusErrorCheck(err, http.StatusNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// GetRemoteClusterDetail returns the detail for a remote cluster.
func GetRemoteClusterDetail(ctx context.Context, tx *sqlx.Tx, remoteClusterID int) (*RemoteClusterDetail, error) {
	q := `
        SELECT 
			id, remote_cluster_id, cpu_total_count, cpu_load_1, cpu_load_5, 
			cpu_load_15, memory_total_amount, memory_usage, 
			instance_count, instance_statuses, member_count, 
			member_statuses, storage_pool_usages, ui_url, created_at, updated_at
        FROM remote_cluster_details
		WHERE remote_cluster_id = $1;
    `

	var result RemoteClusterDetail
	err := tx.QueryRowContext(ctx, q, remoteClusterID).Scan(
		&result.ID,
		&result.RemoteClusterID,
		&result.CPUTotalCount,
		&result.CPULoad1,
		&result.CPULoad5,
		&result.CPULoad15,
		&result.MemoryTotalAmount,
		&result.MemoryUsage,
		&result.InstanceCount,
		&result.InstanceStatuses,
		&result.MemberCount,
		&result.MemberStatuses,
		&result.StoragePoolUsages,
		&result.UIURL,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, api.StatusErrorf(http.StatusNotFound, "detail for remote cluster not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch from \"remote_cluster_details\" table: %w", err)
	}

	return &result, nil
}

// CreateRemoteClusterDetail creates a new detail entry for a remote cluster.
func CreateRemoteClusterDetail(ctx context.Context, tx *sqlx.Tx, data RemoteClusterDetail) (*RemoteClusterDetail, error) {
	exists, err := RemoteClusterDetailExists(ctx, tx, data.RemoteClusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to check for duplicates: %w", err)
	}

	if exists {
		return nil, api.StatusErrorf(http.StatusConflict, "This \"remote_cluster_details\" entry already exists")
	}

	q := `
        INSERT INTO remote_cluster_details 
			(remote_cluster_id, cpu_total_count, cpu_load_1, cpu_load_5, cpu_load_15, memory_total_amount, memory_usage, instance_count, instance_statuses, member_count, member_statuses, storage_pool_usages, ui_url)
        VALUES 
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
        RETURNING 
			id, remote_cluster_id, cpu_total_count, cpu_load_1, cpu_load_5, cpu_load_15, memory_total_amount, memory_usage, instance_count, instance_statuses, member_count, member_statuses, storage_pool_usages, ui_url, created_at, updated_at;
    `

	var result RemoteClusterDetail
	err = tx.QueryRowContext(ctx, q,
		data.RemoteClusterID,
		data.CPUTotalCount,
		data.CPULoad1,
		data.CPULoad5,
		data.CPULoad15,
		data.MemoryTotalAmount,
		data.MemoryUsage,
		data.InstanceCount,
		data.InstanceStatuses,
		data.MemberCount,
		data.MemberStatuses,
		data.StoragePoolUsages,
		data.UIURL,
	).Scan(
		&result.ID,
		&result.RemoteClusterID,
		&result.CPUTotalCount,
		&result.CPULoad1,
		&result.CPULoad5,
		&result.CPULoad15,
		&result.MemoryTotalAmount,
		&result.MemoryUsage,
		&result.InstanceCount,
		&result.InstanceStatuses,
		&result.MemberCount,
		&result.MemberStatuses,
		&result.StoragePoolUsages,
		&result.UIURL,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create \"remote_cluster_details\" entry: %w", err)
	}

	return &result, nil
}

// UpdateRemoteClusterDetail updates a detail entry for a remote cluster.
func UpdateRemoteClusterDetail(ctx context.Context, tx *sqlx.Tx, remoteClusterID int, data RemoteClusterDetail) error {
	id, err := GetRemoteClusterDetailID(ctx, tx, remoteClusterID)
	if err != nil {
		return err
	}

	q := `
        UPDATE remote_cluster_details
        SET cpu_total_count = $1, cpu_load_1 = $2, cpu_load_5 = $3, cpu_load_15 = $4, memory_total_amount = $5, memory_usage = $6, instance_count = $7, instance_statuses = $8, member_count = $9, member_statuses = $10, storage_pool_usages = $11, ui_url = $12, updated_at = NOW()
        WHERE id = $13;
    `

	result, err := tx.ExecContext(ctx, q,
		data.CPUTotalCount,
		data.CPULoad1,
		data.CPULoad5,
		data.CPULoad15,
		data.MemoryTotalAmount,
		data.MemoryUsage,
		data.InstanceCount,
		data.InstanceStatuses,
		data.MemberCount,
		data.MemberStatuses,
		data.StoragePoolUsages,
		data.UIURL,
		id,
	)
	if err != nil {
		return fmt.Errorf("update remote_cluster_details entry failed: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("fetch affected rows: %w", err)
	}

	if n != 1 {
		return fmt.Errorf("query updated %d rows instead of 1", n)
	}

	return nil
}

var baseDetailQuery = `
	SELECT DISTINCT ON (remote_clusters.name)
		remote_clusters.id,
		remote_clusters.name,
		remote_clusters.description,
		remote_clusters.status,
		remote_clusters.cluster_certificate,
		remote_clusters.joined_at,
		remote_clusters.created_at,
		remote_cluster_details.cpu_total_count,
		remote_cluster_details.cpu_load_1,
		remote_cluster_details.cpu_load_5,
		remote_cluster_details.cpu_load_15,
		remote_cluster_details.memory_total_amount,
		remote_cluster_details.memory_usage,
		remote_cluster_details.instance_count,
		remote_cluster_details.instance_statuses,
		remote_cluster_details.member_count,
		remote_cluster_details.member_statuses,
		remote_cluster_details.storage_pool_usages,
		remote_cluster_details.ui_url,
		remote_cluster_details.updated_at,
		COALESCE(
			(SELECT value::bigint FROM remote_cluster_config cfg
				WHERE cfg.remote_cluster_id = remote_clusters.id AND cfg.key = 'disk_threshold'
				LIMIT 1),
	    	0
	    ) AS disk_threshold,
	    COALESCE(
			(SELECT value::bigint FROM remote_cluster_config cfg
				WHERE cfg.remote_cluster_id = remote_clusters.id AND cfg.key = 'memory_threshold'
				LIMIT 1),
	    	0
	    ) AS memory_threshold
	FROM remote_cluster_details
	JOIN remote_clusters ON remote_cluster_details.remote_cluster_id = remote_clusters.id
`

func getRemoteClusterWithDetails(ctx context.Context, tx *sqlx.Tx, sql string, args ...any) ([]RemoteClusterWithDetail, error) {
	objects := make([]RemoteClusterWithDetail, 0)
	dest := func(scan func(dest ...any) error) error {
		c := RemoteClusterWithDetail{}
		err := scan(
			&c.ID,
			&c.Name,
			&c.Description,
			&c.Status,
			&c.ClusterCertificate,
			&c.ClusterJoinedAt,
			&c.ClusterCreatedAt,
			&c.CPUTotalCount,
			&c.CPULoad1,
			&c.CPULoad5,
			&c.CPULoad15,
			&c.MemoryTotalAmount,
			&c.MemoryUsage,
			&c.InstanceCount,
			&c.InstanceStatuses,
			&c.MemberCount,
			&c.MemberStatuses,
			&c.StoragePoolUsages,
			&c.UIURL,
			&c.ClusterUpdatedAt,
			&c.DiskThreshold,
			&c.MemoryThreshold,
		)
		if err != nil {
			return err
		}

		objects = append(objects, c)

		return nil
	}

	err := query.Scan(ctx, tx, sql, dest, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to do a joint fetch from \"remote_clusters\" and \"remote_cluster_details\" tables: %w", err)
	}

	return objects, nil
}

// GetRemoteClustersWithDetails fetches all remote cluster details with remote cluster information from the database.
func GetRemoteClustersWithDetails(ctx context.Context, tx *sqlx.Tx) ([]RemoteClusterWithDetail, error) {
	q := fmt.Sprintf(`%s ORDER BY remote_clusters.name`, baseDetailQuery)
	remoteClusterDetails, err := getRemoteClusterWithDetails(ctx, tx, q)
	if err != nil {
		return nil, err
	}

	return remoteClusterDetails, nil
}

// GetRemoteClusterWithDetailByName fetches the remote cluster detail with remote cluster information from the database filtered by remote cluster name.
func GetRemoteClusterWithDetailByName(ctx context.Context, tx *sqlx.Tx, remoteClusterName string) (*RemoteClusterWithDetail, error) {
	q := fmt.Sprintf(`%s WHERE remote_clusters.name = $1`, baseDetailQuery)
	remoteClusterDetails, err := getRemoteClusterWithDetails(ctx, tx, q, remoteClusterName)
	if err != nil {
		return nil, err
	}

	if len(remoteClusterDetails) == 0 {
		return nil, api.StatusErrorf(http.StatusNotFound, "remote cluster with name %s not found", remoteClusterName)
	}

	return &remoteClusterDetails[0], nil
}

// GetRemoteClusterWithDetailByID fetches the remote cluster detail with information from the database filtered by site id.
func GetRemoteClusterWithDetailByID(ctx context.Context, tx *sqlx.Tx, remoteClusterID int) (*RemoteClusterWithDetail, error) {
	q := fmt.Sprintf(`%s WHERE remote_clusters.id = $1`, baseDetailQuery)
	remoteClusterDetails, err := getRemoteClusterWithDetails(ctx, tx, q, remoteClusterID)
	if err != nil {
		return nil, err
	}

	if len(remoteClusterDetails) == 0 {
		return nil, api.StatusErrorf(http.StatusNotFound, "remote cluster with ID %d not found", remoteClusterID)
	}

	return &remoteClusterDetails[0], nil
}

func parseStatusDistribution(statuses []models.StatusDistribution) (int64, json.RawMessage) {
	if len(statuses) == 0 {
		return 0, json.RawMessage("[]")
	}

	parsedStatuses, err := json.Marshal(statuses)
	if err != nil {
		return 0, json.RawMessage("[]")
	}

	var total int64
	for _, s := range statuses {
		total += s.Count
	}

	return total, parsedStatuses
}

func parsePoolUsage(usages []models.StoragePoolUsage) json.RawMessage {
	if len(usages) == 0 {
		return json.RawMessage("[]")
	}

	parsedUsages, err := json.Marshal(usages)
	if err != nil {
		return json.RawMessage("[]")
	}

	return parsedUsages
}
