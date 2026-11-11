package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/request"
	"github.com/canonical/microcloud-cluster-manager/test/helpers"
	"github.com/gorilla/mux"
)

// newParamsRequest builds a request with the given path variables and raw query.
func newParamsRequest(vars map[string]string, rawQuery string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/?"+rawQuery, nil)
	return mux.SetURLVars(r, vars)
}

// expectBadRequest checks that err is a bad request status error.
func expectBadRequest(name string, err error) error {
	if !api.StatusErrorCheck(err, http.StatusBadRequest) {
		return fmt.Errorf("%s: expected bad request error, got %v", name, err)
	}

	return nil
}

func testRequest_PathParams() (testName string, testFunc func(t *testing.T)) {
	return "Request path parameters are unescaped and validated", func(t *testing.T) {
		condition := "path parameters should be unescaped, and invalid or missing required ones rejected"

		var err error

		r := newParamsRequest(map[string]string{request.RemoteClusterNameParam: "my%2Fcluster", request.ProjectNameParam: ""}, "")
		name, e := request.RemoteClusterName(r)
		if e != nil || name != "my/cluster" {
			err = errors.Join(err, fmt.Errorf("expected remote cluster name %q, got %q (%v)", "my/cluster", name, e))
		}

		_, e = request.RequiredPathParam(r, request.ProjectNameParam, "project name")
		err = errors.Join(err, expectBadRequest("empty required path parameter", e))

		_, hasReplicator, e := request.OptionalPathParam(r, request.ReplicatorNameParam, "replicator name")
		if e != nil || hasReplicator {
			err = errors.Join(err, fmt.Errorf("expected absent optional path parameter, got present=%v (%v)", hasReplicator, e))
		}

		invalid := newParamsRequest(map[string]string{request.RemoteClusterNameParam: "bad%zz"}, "")
		_, e = request.RemoteClusterName(invalid)
		err = errors.Join(err, expectBadRequest("invalid escaped path parameter", e))

		helpers.LogTestOutcome(t, condition, err)
	}
}

func testRequest_QueryParams() (testName string, testFunc func(t *testing.T)) {
	return "Request query parameters are parsed and validated", func(t *testing.T) {
		condition := "query parameters should be parsed, and invalid or missing required ones rejected"

		var err error

		cases := []struct {
			rawQuery string
			want     bool
		}{
			{"", false},
			{"force=true", true},
			{"force=0", false},
		}

		for _, c := range cases {
			got, e := request.BoolQueryParam(newParamsRequest(nil, c.rawQuery), "force")
			if e != nil || got != c.want {
				err = errors.Join(err, fmt.Errorf("query %q: expected %v, got %v (%v)", c.rawQuery, c.want, got, e))
			}
		}

		_, e := request.BoolQueryParam(newParamsRequest(nil, "force=maybe"), "force")
		err = errors.Join(err, expectBadRequest("invalid boolean query parameter", e))

		_, e = request.RequiredQueryParam(newParamsRequest(nil, ""), "project")
		err = errors.Join(err, expectBadRequest("missing required query parameter", e))

		project, e := request.RequiredQueryParam(newParamsRequest(nil, "project=default"), "project")
		if e != nil || project != "default" {
			err = errors.Join(err, fmt.Errorf("expected project %q, got %q (%v)", "default", project, e))
		}

		helpers.LogTestOutcome(t, condition, err)
	}
}
