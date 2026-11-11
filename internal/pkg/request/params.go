package request

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/canonical/lxd/shared/api"
	"github.com/gorilla/mux"
)

// Path parameter names shared by the API routes.
const (
	RemoteClusterNameParam = "remoteClusterName"
	ProjectNameParam       = "projectName"
	ReplicatorNameParam    = "replicatorName"
	ClusterLinkNameParam   = "clusterLinkName"
	OperationIDParam       = "operationID"
)

// PathParam reads and unescapes the path parameter with the given name.
// The label describes the parameter in error messages, e.g. "remote cluster name".
// Returns a bad request error if the parameter can not be unescaped.
func PathParam(r *http.Request, name string, label string) (string, error) {
	value, err := url.PathUnescape(mux.Vars(r)[name])
	if err != nil {
		return "", api.StatusErrorf(http.StatusBadRequest, "Invalid %s: %v", label, err)
	}

	return value, nil
}

// OptionalPathParam reads and unescapes the path parameter with the given name, if it is part of the route.
// The returned bool reports whether the parameter is present.
func OptionalPathParam(r *http.Request, name string, label string) (string, bool, error) {
	_, ok := mux.Vars(r)[name]
	if !ok {
		return "", false, nil
	}

	value, err := PathParam(r, name, label)
	if err != nil {
		return "", true, err
	}

	return value, true, nil
}

// RequiredPathParam reads and unescapes the path parameter with the given name.
// Returns a bad request error if the parameter is empty or can not be unescaped.
func RequiredPathParam(r *http.Request, name string, label string) (string, error) {
	value, err := PathParam(r, name, label)
	if err != nil {
		return "", err
	}

	if value == "" {
		return "", api.StatusErrorf(http.StatusBadRequest, "Missing %s", label)
	}

	return value, nil
}

// RemoteClusterName reads the remote cluster name from the request path.
func RemoteClusterName(r *http.Request) (string, error) {
	return PathParam(r, RemoteClusterNameParam, "remote cluster name")
}

// QueryParam returns the value of the query parameter with the given name.
func QueryParam(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

// RequiredQueryParam returns the value of the query parameter with the given name.
// Returns a bad request error if the parameter is empty.
func RequiredQueryParam(r *http.Request, name string) (string, error) {
	value := QueryParam(r, name)
	if value == "" {
		return "", api.StatusErrorf(http.StatusBadRequest, "Missing %s", name)
	}

	return value, nil
}

// BoolQueryParam reads the optional boolean query parameter with the given name, defaulting to false.
// Returns a bad request error if the value is not a valid boolean.
func BoolQueryParam(r *http.Request, name string) (bool, error) {
	rawValue := QueryParam(r, name)
	if rawValue == "" {
		return false, nil
	}

	value, err := strconv.ParseBool(rawValue)
	if err != nil {
		return false, api.StatusErrorf(http.StatusBadRequest, "Invalid %s value %q", name, rawValue)
	}

	return value, nil
}
