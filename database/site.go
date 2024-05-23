package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/canonical/microcluster/cluster"

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

var siteCreateStmt = cluster.RegisterStmt(`
INSERT INTO sites (name, status) VALUES (?, ?)
`)

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
		return nil, fmt.Errorf("failed to list sites %w", err)
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
		return nil, fmt.Errorf("failed to list sites %w", err)
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

func SiteExists(ctx context.Context, tx *sql.Tx, siteName string) (bool, error) {
	_, err := GetSite(ctx, tx, siteName)
	if err != nil {
		if api.StatusErrorCheck(err, http.StatusNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func CreateSite(ctx context.Context, tx *sql.Tx, site *Site) (int64, error) {
	// Check if a Site with the same name exists.
	exists, err := SiteExists(ctx, tx, site.Name)
	if err != nil {
		return -1, fmt.Errorf("failed to check for duplicates: %w", err)
	}

	if exists {
		return -1, api.StatusErrorf(http.StatusConflict, "This \"site\" entry already exists")
	}

	args := make([]any, 2)

	// Populate the statement arguments.
	args[0] = site.Name
	args[1] = site.Status

	// Prepared statement to use.
	stmt, err := cluster.Stmt(tx, siteCreateStmt)
	if err != nil {
		return -1, fmt.Errorf("failed to get \"siteCreateStmt\" prepared statement: %w", err)
	}

	// Execute the statement.
	result, err := stmt.Exec(args...)
	if err != nil {
		return -1, fmt.Errorf("failed to create \"sites\" entry: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("failed to fetch \"sites\" entry ID: %w", err)
	}

	return id, nil
}
