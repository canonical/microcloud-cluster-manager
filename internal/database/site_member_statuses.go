package database

import (
	"time"
)

//go:generate -command mapper lxd-generate db mapper -t site_member_statuses.mapper.go
//go:generate mapper reset
//
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e site_member_status objects table=site_member_statuses
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e site_member_status objects-by-MemberName table=site_member_statuses
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e site_member_status objects-by-CoreSiteID table=site_member_statuses
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e site_member_status objects-by-MemberName-and-CoreSiteID table=site_member_statuses
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e site_member_status id table=site_member_statuses
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e site_member_status create table=site_member_statuses
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e site_member_status delete-by-MemberName table=site_member_statuses
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e site_member_status delete-by-MemberName-and-CoreSiteID table=site_member_statuses
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e site_member_status update table=site_member_statuses
//
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e site_member_status GetMany
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e site_member_status ID
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e site_member_status Exists
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e site_member_status Create
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e site_member_status DeleteOne-by-MemberName-and-CoreSiteID
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e site_member_status Update

// SiteMemberStatus represents all site member level data.
type SiteMemberStatus struct {
	CoreSiteID   int    `db:"primary=true"`
	MemberName   string `db:"primary=true"`
	ID           int
	Address      string
	Architecture string
	Role         string
	UsageCPU     float64
	UsageMemory  float64
	UsageDisk    float64
	Status       string
	UpdatedAt    time.Time
}

// SiteMemberStatusFilter is a required struct for use with lxd-generate. It is used for filtering fields on database fetches.
type SiteMemberStatusFilter struct {
	MemberName *string
	CoreSiteID *int
}
