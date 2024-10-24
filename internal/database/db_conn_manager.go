package database

import (
	"context"
	"time"

	"github.com/canonical/lxd-cluster-manager/internal/database/pg"
	"github.com/canonical/microcluster/v2/state"
)

type DBConnManager struct {
	DB Database
}

func NewDBConnManager(ctx context.Context, microState state.State, usePostgres, boostrap bool) (*DBConnManager, error) {
	if usePostgres {
		host := "localhost"
		port := "5432"
		user := "admin"
		password := "admin"
		dbName := "lxd_cluster_manager"
		// TODO: fine tune connection pooling settings
		maxIdleConns := 10
		maxOpenConns := 10
		connMaxLifetime := time.Duration(5) * time.Minute

		conn, err := pg.NewPgConnection(
			ctx,
			host,
			port,
			user,
			password,
			dbName,
			maxIdleConns,
			maxOpenConns,
			connMaxLifetime,
		)

		if err != nil {
			return nil, err
		}

		if boostrap {
			err = conn.Bootstrap(ctx)
			if err != nil {
				return nil, err
			}
		}

		return &DBConnManager{DB: conn}, nil
	}

	return &DBConnManager{
		DB: microState.Database(),
	}, nil
}
