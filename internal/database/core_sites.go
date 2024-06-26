package database

import (
	"time"
)

//go:generate -command mapper lxd-generate db mapper -t core_sites.mapper.go
//go:generate mapper reset
//
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e core_site objects table=core_sites
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e core_site objects-by-Name table=core_sites
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e core_site objects-by-SiteCertificate table=core_sites
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e core_site id table=core_sites
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e core_site create table=core_sites
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e core_site delete-by-Name table=core_sites
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e core_site update table=core_sites
//
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_site GetMany
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_site GetOne
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_site ID
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_site Exists
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_site Create
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_site DeleteOne-by-Name
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_site Update

// CoreSite represents data about a site that would be common across all microcluster based projects.
type CoreSite struct {
	Name            string `db:"primary=yes"`
	SiteCertificate string `db:"primary=yes"`
	ID              int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// CoreSiteFilter is a required struct for use with lxd-generate. It is used for filtering fields on database fetches.
type CoreSiteFilter struct {
	Name            *string
	SiteCertificate *string
}
