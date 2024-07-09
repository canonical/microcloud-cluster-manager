package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/canonical/lxd/lxd/db/query"
	"github.com/canonical/microcluster/cluster"
)

// CoreSiteWithDetails is a struct that contains all the information about a site directly queried from the database.
type CoreSiteWithDetails struct {
	ID                int       `db:"id"`
	Name              string    `db:"name"`
	SiteCertificate   string    `db:"site_certificate"`
	SiteCreatedAt     time.Time `db:"created_at"`
	Status            string    `db:"status"`
	CPUTotalCount     float64   `db:"cpu_total_count"`
	CPULoad1          string    `db:"cpu_load_1"`
	CPULoad5          string    `db:"cpu_load_5"`
	CPULoad15         string    `db:"cpu_load_15"`
	MemoryTotalAmount float64   `db:"memory_total_amount"`
	MemoryUsage       float64   `db:"memory_usage"`
	DiskTotalSize     float64   `db:"disk_total_size"`
	DiskUsage         float64   `db:"disk_usage"`
	InstanceCount     int       `db:"instance_count"`
	InstanceStatuses  string    `db:"instance_statuses"`
	MemberCount       int       `db:"member_count"`
	MemberStatuses    string    `db:"member_statuses"`
	SiteJoinedAt      time.Time `db:"joined_at"`
	SiteUpdatedAt     time.Time `db:"updated_at"`
}

func mainCoreSiteDetailQuery() string {
	return `
		SELECT
			core_sites.id, core_sites.name, core_sites.site_certificate, core_sites.created_at,
			site_details.status, site_details.cpu_total_count, site_details.cpu_load_1, site_details.cpu_load_5, site_details.cpu_load_15, site_details.memory_total_amount, site_details.memory_usage, 
			site_details.disk_total_size, site_details.disk_usage, site_details.instance_count, site_details.instance_statuses, 
			site_details.member_count, site_details.member_statuses, site_details.joined_at, site_details.updated_at
		FROM site_details
		JOIN core_sites ON site_details.core_site_id = core_sites.id
	`
}

var coreSiteDetailObjects = cluster.RegisterStmt(
	fmt.Sprintf(`%s ORDER BY core_sites.name`, mainCoreSiteDetailQuery()),
)

var coreSiteDetailBySiteNameObjects = cluster.RegisterStmt(
	fmt.Sprintf(`%s WHERE core_sites.name = ?`, mainCoreSiteDetailQuery()),
)

// GetCoreSitesWithDetails fetches all site details with core site information from the database.
func GetCoreSitesWithDetails(ctx context.Context, tx *sql.Tx) ([]CoreSiteWithDetails, error) {
	var err error
	objects := make([]CoreSiteWithDetails, 0)
	sqlStmt, err := cluster.Stmt(tx, coreSiteDetailObjects)
	if err != nil {
		return nil, fmt.Errorf("Failed to prepare statement: %w", err)
	}

	dest := func(scan func(dest ...any) error) error {
		c := CoreSiteWithDetails{}
		err := scan(
			&c.ID,
			&c.Name,
			&c.SiteCertificate,
			&c.SiteCreatedAt,
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
			&c.SiteJoinedAt,
			&c.SiteUpdatedAt,
		)

		if err != nil {
			return err
		}

		objects = append(objects, c)

		return nil
	}

	err = query.SelectObjects(ctx, sqlStmt, dest)
	if err != nil {
		return nil, fmt.Errorf("Failed to do a joint fetch from \"cores_sites\" and \"site_details\" tables: %w", err)
	}

	return objects, nil
}

// GetCoreSiteWithDetailBySiteName fetches the site detail with core site information from the database filtered by site name.
func GetCoreSiteWithDetailBySiteName(ctx context.Context, tx *sql.Tx, siteName string) ([]CoreSiteWithDetails, error) {
	var err error
	objects := make([]CoreSiteWithDetails, 0)
	sqlStmt, err := cluster.Stmt(tx, coreSiteDetailBySiteNameObjects)
	if err != nil {
		return nil, fmt.Errorf("Failed to prepare statement: %w", err)
	}

	dest := func(scan func(dest ...any) error) error {
		c := CoreSiteWithDetails{}
		err := scan(
			&c.ID,
			&c.Name,
			&c.SiteCertificate,
			&c.SiteCreatedAt,
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
			&c.SiteJoinedAt,
			&c.SiteUpdatedAt,
		)

		if err != nil {
			return err
		}

		objects = append(objects, c)

		return nil
	}

	err = query.SelectObjects(ctx, sqlStmt, dest, siteName)
	if err != nil {
		return nil, fmt.Errorf("Failed to do a joint fetch from \"cores_sites\" and \"site_details\" tables: %w", err)
	}

	return objects, nil
}
