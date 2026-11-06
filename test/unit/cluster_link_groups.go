package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/canonical/lxd/shared/api"
	apiV1 "github.com/canonical/microcloud-cluster-manager/internal/app/management-api/api/v1"
	"github.com/canonical/microcloud-cluster-manager/test/helpers"
)

func testClusterLinkGroups_PartialFailures() (testName string, testFunc func(t *testing.T)) {
	return "collecting cluster links reports failures per cluster", func(t *testing.T) {
		var condition string

		results := apiV1.CollectClusterLinks(context.Background(), []string{"healthy", "failed", "empty"}, time.Second, func(ctx context.Context, name string) ([]api.ClusterLink, error) {
			switch name {
			case "healthy":
				return []api.ClusterLink{{Name: "target"}}, nil
			case "failed":
				return nil, errors.New("cluster unavailable")
			default:
				return nil, nil
			}
		})

		encoded, err := json.Marshal(results)
		helpers.LogTestOutcome(t, "Should encode the groups as JSON", err)

		var entries []map[string]json.RawMessage
		err = json.Unmarshal(encoded, &entries)
		helpers.LogTestOutcome(t, "Should decode the groups as a list of objects", err)

		{
			condition = "Should return a group per cluster, in the order of the names, with the error of the failed one"

			err = nil
			if len(results) != 3 || results[0].Cluster != "healthy" || len(results[0].Links) != 1 || results[1].Error != "cluster unavailable" {
				err = fmt.Errorf("unexpected groups: %s", encoded)
			}

			helpers.LogTestOutcome(t, condition, err)
		}

		{
			condition = "Should omit the error of a cluster that was loaded"

			err = nil
			if _, exists := entries[0]["error"]; exists {
				err = fmt.Errorf("the error is set: %s", encoded)
			}

			helpers.LogTestOutcome(t, condition, err)
		}

		{
			condition = "Should encode the links of failed and empty clusters as an empty list"

			err = nil
			if string(entries[1]["links"]) != "[]" || string(entries[2]["links"]) != "[]" {
				err = fmt.Errorf("links are not empty lists: %s", encoded)
			}

			helpers.LogTestOutcome(t, condition, err)
		}
	}
}

func testClusterLinkGroups_Timeout() (testName string, testFunc func(t *testing.T)) {
	return "collecting cluster links times out slow clusters", func(t *testing.T) {
		var condition string

		const timeout = 50 * time.Millisecond

		results := apiV1.CollectClusterLinks(context.Background(), []string{"slow", "healthy"}, timeout, func(ctx context.Context, name string) ([]api.ClusterLink, error) {
			if name == "healthy" {
				return []api.ClusterLink{{Name: "target"}}, nil
			}

			// The slow fetch only returns once its context ends. The test context is never
			// cancelled, so this can only be the per-cluster timeout.
			<-ctx.Done()
			return nil, ctx.Err()
		})

		{
			condition = "Should report the timeout of the slow cluster and keep the links of the healthy one"

			var err error
			if results[0].Error != context.DeadlineExceeded.Error() || len(results[1].Links) != 1 {
				err = fmt.Errorf("unexpected groups: %#v", results)
			}

			helpers.LogTestOutcome(t, condition, err)
		}
	}
}
