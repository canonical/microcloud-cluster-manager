package database

import (
	"time"
)

//go:generate -command mapper lxd-generate db mapper -t site_details.mapper.go
//go:generate mapper reset
//
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e site_detail objects table=site_details
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e site_detail objects-by-CoreSiteID table=site_details
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e site_detail id table=site_details
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e site_detail create table=site_details
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e site_detail update table=site_details
//
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e site_detail GetMany
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e site_detail GetOne
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e site_detail ID
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e site_detail Exists
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e site_detail Create
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e site_detail Update

// SiteDetail represents all site level data.
type SiteDetail struct {
	CoreSiteID        int64 `db:"primary=true"`
	Status            string
	ID                int
	CPUTotalCount     int64
	CPULoad1          string `db:"sql=site_details.cpu_load_1"`
	CPULoad5          string `db:"sql=site_details.cpu_load_5"`
	CPULoad15         string `db:"sql=site_details.cpu_load_15"`
	MemoryTotalAmount int64
	MemoryUsage       int64
	DiskTotalSize     int64
	DiskUsage         int64
	InstanceCount     int64
	InstanceStatuses  string
	MemberCount       int64
	MemberStatuses    string
	JoinedAt          time.Time
	UpdatedAt         time.Time
}

// SiteDetailFilter is a required struct for use with lxd-generate. It is used for filtering fields on database fetches.
type SiteDetailFilter struct {
	CoreSiteID *int64
}
