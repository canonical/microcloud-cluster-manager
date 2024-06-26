package database

import (
	"time"
)

//go:generate -command mapper lxd-generate db mapper -t core_site_tokens.mapper.go
//go:generate mapper reset
//
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e core_site_token objects table=core_site_tokens
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e core_site_token objects-by-SiteName table=core_site_tokens
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e core_site_token id table=core_site_tokens
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e core_site_token create table=core_site_tokens
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e core_site_token delete-by-SiteName table=core_site_tokens
//go:generate mapper stmt -d github.com/canonical/microcluster/cluster -e core_site_token update table=core_site_tokens
//
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_site_token GetMany
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_site_token GetOne
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_site_token ID
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_site_token Exists
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_site_token Create
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_site_token DeleteOne-by-SiteName
//go:generate mapper method -i -d github.com/canonical/microcluster/cluster -e core_site_token Update

// CoreSiteToken is the join token for a site.
type CoreSiteToken struct {
	SiteName  string `db:"primary=true"`
	ID        int
	Secret    string
	Expiry    time.Time
	CreatedAt time.Time
}

// CoreSiteTokenFilter is a required struct for use with lxd-generate. It is used for filtering fields on database fetches.
type CoreSiteTokenFilter struct {
	SiteName *string
}
