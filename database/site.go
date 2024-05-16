package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/canonical/lxd/lxd/db/query"
	"github.com/canonical/lxd/shared/api"
)

// Site represents a single LXD site.
type Site struct {
	ID        int
	Name      string
	Addresses []string
	Status    string
}

// GetSites returns all sites from the database.
func GetSites(ctx context.Context, tx *sql.Tx) ([]Site, error) {
	stmt := `
SELECT sites.id, sites.name, sites.status, sites_addresses.address 
FROM sites_addresses
JOIN sites ON sites_addresses.site_id = sites.id`

	result := make(map[int]*Site)
	dest := func(scan func(dest ...any) error) error {
		s := Site{}
		var addr string
		err := scan(&s.ID, &s.Name, &s.Status, &addr)
		if err != nil {
			return err
		}

		existingSite, ok := result[s.ID]
		if !ok {
			s.Addresses = []string{addr}
			result[s.ID] = &s
			return nil
		}

		existingSite.Addresses = append(existingSite.Addresses, addr)
		return nil
	}

	err := query.Scan(ctx, tx, stmt, dest)
	if err != nil {
		return nil, fmt.Errorf("Failed to list sites %w", err)
	}

	sites := make([]Site, 0, len(result))
	for _, site := range result {
		sites = append(sites, *site)
	}

	// TODO: Maybe sort this.
	return sites, nil
}

// GetSite returns a single site by name.
func GetSite(ctx context.Context, tx *sql.Tx, siteName string) (*Site, error) {
	stmt := `
SELECT sites.id, sites.name, sites.status, sites_addresses.address 
FROM sites_addresses
JOIN sites ON sites_addresses.site_id = sites.id WHERE sites.name = ?`

	result := make(map[int]*Site)
	dest := func(scan func(dest ...any) error) error {
		s := Site{}
		var addr string
		err := scan(&s.ID, &s.Name, &s.Status, &addr)
		if err != nil {
			return err
		}

		existingSite, ok := result[s.ID]
		if !ok {
			s.Addresses = []string{addr}
			result[s.ID] = &s
			return nil
		}

		existingSite.Addresses = append(existingSite.Addresses, addr)
		return nil
	}

	err := query.Scan(ctx, tx, stmt, dest, siteName)
	if err != nil {
		return nil, fmt.Errorf("Failed to list sites %w", err)
	}

	if len(result) == 0 {
		return nil, api.StatusErrorf(http.StatusBadRequest, "Site %q not found", siteName)
	} else if len(result) > 1 {
		return nil, api.StatusErrorf(http.StatusInternalServerError, "Multiple sites found for name %q", siteName)
	}

	var site *Site
	for _, s := range result {
		site = s
	}

	return site, nil
}
