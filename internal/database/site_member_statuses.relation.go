package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/canonical/lxd/lxd/db/query"
	"github.com/canonical/microcluster/cluster"
)

// MemberStatusWithSiteInfo is a struct that contains all the information about a site member directly queried from the database.
type MemberStatusWithSiteInfo struct {
	ID               int       `db:"id"`
	Name             string    `db:"name"`
	SiteCertificate  string    `db:"site_certificate"`
	SiteCreatedAt    time.Time `db:"created_at"`
	Status           string    `db:"site_status"`
	InstanceCount    int       `db:"instance_count"`
	InstanceStatuses string    `db:"instance_statuses"`
	SiteJoinedAt     time.Time `db:"joined_at"`
	SiteUpdatedAt    time.Time `db:"updated_at"`
	MemberName       string    `db:"member_name"`
	Address          string    `db:"address"`
	Architecture     string    `db:"architecture"`
	Role             string    `db:"role"`
	UsageCPU         float64   `db:"usage_cpu"`
	UsageMemory      float64   `db:"usage_memory"`
	UsageDisk        float64   `db:"usage_disk"`
	MemberStatus     string    `db:"member_status"`
}

func mainSiteMemberStatusQuery() string {
	return `
		SELECT
			core_sites.id, core_sites.name, core_sites.site_certificate, core_sites.created_at,
			site_details.status, site_details.instance_count, site_details.instance_statuses, site_details.joined_at, site_details.updated_at,
			site_member_statuses.member_name, site_member_statuses.address, site_member_statuses.architecture, site_member_statuses.role, site_member_statuses.usage_cpu, site_member_statuses.usage_memory, site_member_statuses.usage_disk, site_member_statuses.status as member_status
		FROM site_member_statuses
		JOIN core_sites ON site_member_statuses.core_site_id = core_sites.id
		JOIN site_details ON site_member_statuses.core_site_id = site_details.core_site_id
	`
}

var memberStatusWithSiteInfoObjects = cluster.RegisterStmt(
	fmt.Sprintf(`%s ORDER BY core_sites.name`, mainSiteMemberStatusQuery()),
)

var memberStatusWithSiteInfoBySiteNameObjects = cluster.RegisterStmt(
	fmt.Sprintf(`%s WHERE core_sites.name = ?`, mainSiteMemberStatusQuery()),
)

// GetMemberStatusesWithSiteInfo fetches all the member statuses with site information from the database.
func GetMemberStatusesWithSiteInfo(ctx context.Context, tx *sql.Tx) ([]MemberStatusWithSiteInfo, error) {
	var err error
	objects := make([]MemberStatusWithSiteInfo, 0)
	sqlStmt, err := cluster.Stmt(tx, memberStatusWithSiteInfoObjects)
	if err != nil {
		return nil, fmt.Errorf("Failed to prepare statement: %w", err)
	}

	dest := func(scan func(dest ...any) error) error {
		c := MemberStatusWithSiteInfo{}
		err := scan(
			&c.ID,
			&c.Name,
			&c.SiteCertificate,
			&c.SiteCreatedAt,
			&c.Status,
			&c.InstanceCount,
			&c.InstanceStatuses,
			&c.SiteJoinedAt,
			&c.SiteUpdatedAt,
			&c.MemberName,
			&c.Address,
			&c.Architecture,
			&c.Role,
			&c.UsageCPU,
			&c.UsageMemory,
			&c.UsageDisk,
			&c.MemberStatus,
		)

		if err != nil {
			return err
		}

		objects = append(objects, c)

		return nil
	}

	err = query.SelectObjects(ctx, sqlStmt, dest)
	if err != nil {
		return nil, fmt.Errorf("Failed to do a joint fetch from \"cores_sites\", \"site_details\" and \"site_member_statuses\" table: %w", err)
	}

	return objects, nil
}

// GetMemberStatusesWithSiteInfoBySiteName fetches all the member statuses with site information from the database filtered by site name.
func GetMemberStatusesWithSiteInfoBySiteName(ctx context.Context, tx *sql.Tx, siteName string) ([]MemberStatusWithSiteInfo, error) {
	var err error
	objects := make([]MemberStatusWithSiteInfo, 0)
	sqlStmt, err := cluster.Stmt(tx, memberStatusWithSiteInfoBySiteNameObjects)
	if err != nil {
		return nil, fmt.Errorf("Failed to prepare statement: %w", err)
	}

	dest := func(scan func(dest ...any) error) error {
		c := MemberStatusWithSiteInfo{}
		err := scan(
			&c.ID,
			&c.Name,
			&c.SiteCertificate,
			&c.SiteCreatedAt,
			&c.Status,
			&c.InstanceCount,
			&c.InstanceStatuses,
			&c.SiteJoinedAt,
			&c.SiteUpdatedAt,
			&c.MemberName,
			&c.Address,
			&c.Architecture,
			&c.Role,
			&c.UsageCPU,
			&c.UsageMemory,
			&c.UsageDisk,
			&c.MemberStatus,
		)

		if err != nil {
			return err
		}

		objects = append(objects, c)

		return nil
	}

	err = query.SelectObjects(ctx, sqlStmt, dest, siteName)
	if err != nil {
		return nil, fmt.Errorf("Failed to do a joint fetch from \"cores_sites\", \"site_details\" and \"site_member_statuses\" table: %w", err)
	}

	return objects, nil
}
