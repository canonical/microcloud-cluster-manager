package seed

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/canonical/lxd-cluster-manager/internal/pkg/database"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/database/store"
	"github.com/jmoiron/sqlx"
)

// SeedDatabase seeds the database with sample data.
func SeedDatabase(ctx context.Context, db *database.DB) error {
	// Seed remote clusters
	if err := seedRemoteClusters(ctx, db); err != nil {
		return fmt.Errorf("failed to seed remote clusters: %w", err)
	}

	// Seed remote cluster tokens
	if err := seedRemoteClusterTokens(ctx, db); err != nil {
		return fmt.Errorf("failed to seed remote cluster tokens: %w", err)
	}

	// Seed remote cluster details
	if err := seedRemoteClusterDetails(ctx, db); err != nil {
		return fmt.Errorf("failed to seed remote cluster details: %w", err)
	}

	return nil
}

// generateRemoteClusters generates a slice of remote clusters with the given count.
func generateRemoteClusters(count int) []store.RemoteCluster {
	clusters := make([]store.RemoteCluster, count)

	for i := 0; i < count; i++ {
		status := "ACTIVE"
		clusters[i] = store.RemoteCluster{
			Name:               fmt.Sprintf("cluster-%02d", i+1),
			Status:             status,
			ClusterCertificate: fmt.Sprintf("cert-%02d", i+1),
			JoinedAt:           time.Now(),
			CreatedAt:          time.Now(),
		}
	}
	return clusters
}

// seedRemoteClusters inserts multiple remote clusters into the database.
func seedRemoteClusters(ctx context.Context, db *database.DB) error {
	remoteClusters := generateRemoteClusters(20)

	return db.Transaction(ctx, func(ctx context.Context, tx *sqlx.Tx) error {
		for _, cluster := range remoteClusters {
			q := `
				INSERT INTO remote_clusters (name, status, cluster_certificate, joined_at, created_at)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (name)
				DO UPDATE SET
					name = $1,
					status = $2,
					cluster_certificate = $3,
					joined_at = $4
				RETURNING id, name, status, cluster_certificate, joined_at, created_at, updated_at;
			`

			var result store.RemoteCluster
			err := tx.QueryRowContext(ctx, q,
				cluster.Name,
				cluster.Status,
				cluster.ClusterCertificate,
				cluster.JoinedAt,
				cluster.CreatedAt,
			).Scan(
				&result.ID,
				&result.Name,
				&result.Status,
				&result.ClusterCertificate,
				&result.JoinedAt,
				&result.CreatedAt,
				&result.UpdatedAt,
			)

			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("failed to create \"remote_clusters\" entry: %w", err)
			}
		}

		return nil
	})
}

// generateRemoteClusterTokens generates a slice of RemoteClusterToken with the specified number of entries.
func generateRemoteClusterTokens(count int) []store.RemoteClusterToken {
	tokens := make([]store.RemoteClusterToken, count)

	for i := 0; i < count; i++ {
		tokens[i] = store.RemoteClusterToken{
			ClusterName:  fmt.Sprintf("cluster-%02d", i+1),
			EncodedToken: fmt.Sprintf("encoded-token-%02d", i+1),
			Expiry:       time.Now().Add(30 * time.Hour),
			CreatedAt:    time.Now(),
		}
	}
	return tokens
}

// seedRemoteClusterTokens inserts multiple remote cluster tokens into the database.
func seedRemoteClusterTokens(ctx context.Context, db *database.DB) error {
	remoteClusterTokens := generateRemoteClusterTokens(20)

	return db.Transaction(ctx, func(ctx context.Context, tx *sqlx.Tx) error {
		for _, token := range remoteClusterTokens {
			q := `
				INSERT INTO remote_cluster_tokens (cluster_name, encoded_token, expiry, created_at)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (cluster_name)
				DO UPDATE SET
					cluster_name = $1,
					encoded_token = $2,
					expiry = $3,
					created_at = $4
				RETURNING id, cluster_name, encoded_token, expiry, created_at;
			`

			var result store.RemoteClusterToken
			err := tx.QueryRowContext(ctx, q,
				token.ClusterName,
				token.EncodedToken,
				token.Expiry,
				token.CreatedAt,
			).Scan(
				&result.ID,
				&result.ClusterName,
				&result.EncodedToken,
				&result.Expiry,
				&result.CreatedAt,
			)

			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("failed to create \"remote_cluster_tokens\" entry: %w", err)
			}
		}

		return nil
	})
}

// generateRandomStatuses generates a JSON array of statuses with random counts.
func generateRandomStatuses(status1, status2, status3, status4 string, maxCount int) []byte {
	// Generate random values for the first three statuses
	count1 := rand.IntN(maxCount / 2) // Max is half to balance distribution
	count2 := rand.IntN(maxCount - count1)
	count3 := rand.IntN(maxCount - count1 - count2)
	count4 := maxCount - count1 - count2 - count3

	statuses := []map[string]any{
		{"status": status1, "count": count1},
		{"status": status2, "count": count2},
		{"status": status3, "count": count3},
		{"status": status4, "count": count4},
	}

	result, _ := json.Marshal(statuses)
	return result
}

// GenerateRemoteClusterDetails generates a slice of RemoteClusterDetail with the specified number of entries.
func generateRemoteClusterDetails(count int) []store.RemoteClusterDetail {
	clusters := make([]store.RemoteClusterDetail, count)

	for i := 0; i < count; i++ {
		totalMemory := (rand.IntN(16) + 1) * 1024 // Random memory in multiples of 1024
		memoryUsage := rand.IntN(totalMemory + 1)
		totalDisk := (rand.IntN(200) + 1) * 1000 // Random disk size in multiples of 1000
		diskUsage := rand.IntN(totalDisk + 1)
		totalInstances := rand.IntN(50) + 2
		totalMembers := rand.IntN(20) + 2

		clusters[i] = store.RemoteClusterDetail{
			RemoteClusterID:   i + 1,
			CPUTotalCount:     rand.IntN(32) + 1, // Random CPU count between 1 and 32
			CPULoad1:          fmt.Sprintf("%.1f", rand.Float64()),
			CPULoad5:          fmt.Sprintf("%.1f", rand.Float64()),
			CPULoad15:         fmt.Sprintf("%.1f", rand.Float64()),
			MemoryTotalAmount: totalMemory,
			MemoryUsage:       memoryUsage,
			DiskTotalSize:     totalDisk,
			DiskUsage:         diskUsage,
			InstanceCount:     totalInstances,
			InstanceStatuses:  generateRandomStatuses("Running", "Stopped", "Frozen", "Error", totalInstances),
			MemberCount:       totalMembers,
			MemberStatuses:    generateRandomStatuses("Online", "Offline", "Evacuated", "Blocked", totalMembers),
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}
	}
	return clusters
}

// seedRemoteClusterDetails inserts multiple remote cluster details into the database.
func seedRemoteClusterDetails(ctx context.Context, db *database.DB) error {
	remoteClusterDetails := generateRemoteClusterDetails(20)

	return db.Transaction(ctx, func(ctx context.Context, tx *sqlx.Tx) error {
		clusters, err := store.GetRemoteClusters(ctx, tx)
		if err != nil {
			return fmt.Errorf("failed to get remote clusters: %w", err)
		}

		for idx, detail := range remoteClusterDetails {
			q := `
				INSERT INTO remote_cluster_details 
					(remote_cluster_id, cpu_total_count, cpu_load_1, cpu_load_5, cpu_load_15, memory_total_amount, memory_usage, disk_total_size, disk_usage, instance_count, instance_statuses, member_count, member_statuses)
				VALUES 
					($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
				ON CONFLICT (remote_cluster_id)
				DO UPDATE SET
					remote_cluster_id = $1,
					cpu_total_count = $2,
					cpu_load_1 = $3,
					cpu_load_5 = $4,
					cpu_load_15 = $5,
					memory_total_amount = $6,
					memory_usage = $7,
					disk_total_size = $8,
					disk_usage = $9,
					instance_count = $10,
					instance_statuses = $11,
					member_count = $12,
					member_statuses = $13
				RETURNING 
					id, remote_cluster_id, cpu_total_count, cpu_load_1, cpu_load_5, cpu_load_15, memory_total_amount, memory_usage, disk_total_size, disk_usage, instance_count, instance_statuses, member_count, member_statuses, created_at, updated_at;
			`

			detail.RemoteClusterID = clusters[idx].ID
			var result store.RemoteClusterDetail
			err := tx.QueryRowContext(ctx, q,
				detail.RemoteClusterID,
				detail.CPUTotalCount,
				detail.CPULoad1,
				detail.CPULoad5,
				detail.CPULoad15,
				detail.MemoryTotalAmount,
				detail.MemoryUsage,
				detail.DiskTotalSize,
				detail.DiskUsage,
				detail.InstanceCount,
				detail.InstanceStatuses,
				detail.MemberCount,
				detail.MemberStatuses,
			).Scan(
				&result.ID,
				&result.RemoteClusterID,
				&result.CPUTotalCount,
				&result.CPULoad1,
				&result.CPULoad5,
				&result.CPULoad15,
				&result.MemoryTotalAmount,
				&result.MemoryUsage,
				&result.DiskTotalSize,
				&result.DiskUsage,
				&result.InstanceCount,
				&result.InstanceStatuses,
				&result.MemberCount,
				&result.MemberStatuses,
				&result.CreatedAt,
				&result.UpdatedAt,
			)

			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("failed to create \"remote_cluster_details\" entry: %w", err)
			}
		}
		return nil
	})
}
