package main

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/canonical/microcloud-cluster-manager/test/helpers"
	"github.com/jmoiron/sqlx"
)

// returnsWithin runs fn and returns an error if it doesn't return within the given duration.
func returnsWithin(d time.Duration, fn func() error) error {
	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()

	select {
	case err := <-done:
		if err == nil {
			return errors.New("expected an error while waiting for a connection, got nil")
		}

		return nil
	case <-time.After(d):
		return fmt.Errorf("did not return within %s", d)
	}
}

func testDatabase_StatusCheckHonoursContext() (testName string, testFunc func(t *testing.T)) {
	return "database status check stops waiting for a connection when the context is done", func(t *testing.T) {
		var condition string

		db, cleanup, err := helpers.NewDBWithExhaustedPool()
		helpers.LogTestOutcome(t, "Should set up a database with no free connections", err)
		defer cleanup()

		{
			condition = "Should return once the context times out"

			err = returnsWithin(5*time.Second, func() error {
				ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
				defer cancel()

				return db.StatusCheck(ctx)
			})

			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

func testDatabase_TransactionHonoursContext() (testName string, testFunc func(t *testing.T)) {
	return "database transaction stops waiting for a connection when the context is done", func(t *testing.T) {
		var condition string

		db, cleanup, err := helpers.NewDBWithExhaustedPool()
		helpers.LogTestOutcome(t, "Should set up a database with no free connections", err)
		defer cleanup()

		{
			condition = "Should return once the context times out"

			err = returnsWithin(5*time.Second, func() error {
				ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
				defer cancel()

				return db.Transaction(ctx, func(ctx context.Context, tx *sqlx.Tx) error {
					return nil
				})
			})

			helpers.LogTestOutcome(t, condition, err)
		}
	}
}
