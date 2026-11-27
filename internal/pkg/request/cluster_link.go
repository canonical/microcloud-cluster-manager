package request

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/canonical/lxd/shared/api"
)

const (
	// ClusterQueryParam selects the cluster that owns the cluster links.
	ClusterQueryParam = "cluster"
	// AllClustersQueryParam selects the cluster links of all managed clusters.
	AllClustersQueryParam = "all-clusters"
)

// ListScope is the set of clusters whose cluster links are listed: either one cluster or all managed clusters.
type ListScope struct {
	Cluster     string
	AllClusters bool
}

// ParseListScope parses the scope of a cluster link list request from the raw query. Exactly one of the cluster or
// all-clusters parameters must be given, once. The cluster must not be empty and all-clusters must be true. Any other
// combination is a bad request.
func ParseListScope(rawQuery string) (ListScope, error) {
	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		return ListScope{}, api.StatusErrorf(http.StatusBadRequest, "Invalid query: %v", err)
	}

	clusterValues, hasCluster := query[ClusterQueryParam]
	allClustersValues, hasAllClusters := query[AllClustersQueryParam]

	switch {
	case hasCluster && hasAllClusters:
		return ListScope{}, api.StatusErrorf(http.StatusBadRequest, "Query parameters %q and %q cannot be combined", ClusterQueryParam, AllClustersQueryParam)
	case hasCluster:
		cluster, err := singleClusterName(clusterValues)
		if err != nil {
			return ListScope{}, err
		}

		return ListScope{Cluster: cluster}, nil
	case hasAllClusters:
		if len(allClustersValues) != 1 {
			return ListScope{}, api.StatusErrorf(http.StatusBadRequest, "Query parameter %q must be given once", AllClustersQueryParam)
		}

		allClusters, err := strconv.ParseBool(allClustersValues[0])
		if err != nil || !allClusters {
			return ListScope{}, api.StatusErrorf(http.StatusBadRequest, "Invalid %s value %q, expected true", AllClustersQueryParam, allClustersValues[0])
		}

		return ListScope{AllClusters: true}, nil
	default:
		return ListScope{}, api.StatusErrorf(http.StatusBadRequest, "Missing query parameter %q or %q", ClusterQueryParam, AllClustersQueryParam)
	}
}

// ParseSourceCluster parses the required cluster parameter for a cluster-link request scoped to one cluster.
func ParseSourceCluster(rawQuery string) (string, error) {
	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		return "", api.StatusErrorf(http.StatusBadRequest, "Invalid query: %v", err)
	}

	_, hasAllClusters := query[AllClustersQueryParam]
	if hasAllClusters {
		return "", api.StatusErrorf(http.StatusBadRequest, "Query parameter %q is only supported for fleet-wide listing through the management API", AllClustersQueryParam)
	}

	clusterValues, hasCluster := query[ClusterQueryParam]
	if !hasCluster {
		return "", api.StatusErrorf(http.StatusBadRequest, "Missing query parameter %q", ClusterQueryParam)
	}

	return singleClusterName(clusterValues)
}

func singleClusterName(values []string) (string, error) {
	if len(values) != 1 {
		return "", api.StatusErrorf(http.StatusBadRequest, "Query parameter %q must be given once", ClusterQueryParam)
	}

	if strings.TrimSpace(values[0]) == "" {
		return "", api.StatusErrorf(http.StatusBadRequest, "Query parameter %q must not be empty", ClusterQueryParam)
	}

	return values[0], nil
}

// ParseCleanupTarget parses the optional cleanupTarget switch on a delete request.
func ParseCleanupTarget(query url.Values) (bool, error) {
	for name := range query {
		if name != ClusterQueryParam && name != "cleanupTarget" {
			return false, api.StatusErrorf(http.StatusBadRequest, "Unknown delete query parameter %q", name)
		}
	}

	values, exists := query["cleanupTarget"]
	if !exists {
		return false, nil
	}
	if len(values) != 1 || (values[0] != "true" && values[0] != "false") {
		return false, api.StatusErrorf(http.StatusBadRequest, "Query parameter cleanupTarget must be given once as true or false")
	}
	return values[0] == "true", nil
}
