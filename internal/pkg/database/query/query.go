package query

import (
	"context"
	"database/sql"

	"github.com/canonical/microcloud-cluster-manager/internal/pkg/logger"
	"github.com/jmoiron/sqlx"
)

// Dest is a function that is expected to return the objects to pass to the
// 'dest' argument of sql.Rows.Scan(). It is invoked by SelectObjects once per
// yielded row, and it will be passed the index of the row being scanned.
type Dest func(scan func(dest ...any) error) error

// Scan runs a query with inArgs and provides the rowFunc with the scan function for each row.
// It handles closing the rows and errors from the result set.
func Scan(ctx context.Context, tx *sqlx.Tx, sql string, rowFunc Dest, inArgs ...any) error {
	rows, err := tx.QueryContext(ctx, sql, inArgs...)
	if err != nil {
		return err
	}

	defer ScanCleanup(rows)

	for rows.Next() {
		err = rowFunc(rows.Scan)
		if err != nil {
			return err
		}
	}

	return rows.Err()
}

func ScanCleanup(rows *sql.Rows) {
	err := rows.Close()
	if err != nil {
		logger.Log.Errorw("query scan cleanup: failed to close rows", err)
	}
}
