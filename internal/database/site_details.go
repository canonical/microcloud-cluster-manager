package database

import (
	"time"
)

//go:generate -command mapper lxd-generate db mapper -t site_details.mapper.go
//go:generate mapper reset
//
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e site_detail objects table=site_details
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e site_detail objects-by-Status table=site_details
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
	CoreSiteID        int64  `db:"primary=true"`
	Status            string `db:"primary=true"`
	ID                int
	CPUTotalCount     float64
	CPULoad1          string `db:"sql=site_details.cpu_load_1"`
	CPULoad5          string `db:"sql=site_details.cpu_load_5"`
	CPULoad15         string `db:"sql=site_details.cpu_load_15"`
	MemoryTotalAmount float64
	MemoryUsage       float64
	DiskTotalSize     float64
	DiskUsage         float64
	InstanceCount     int
	InstanceStatuses  string
	MemberCount       int
	MemberStatuses    string
	JoinedAt          time.Time
	UpdatedAt         time.Time
}

// SiteDetailFilter is a required struct for use with lxd-generate. It is used for filtering fields on database fetches.
type SiteDetailFilter struct {
	Status     *string
	CoreSiteID *int64
}
