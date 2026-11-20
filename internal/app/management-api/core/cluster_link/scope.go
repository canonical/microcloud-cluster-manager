package cluster_link

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

// ParseSourceCluster parses the source cluster of the cluster link of a write request from the raw query. The cluster
// parameter must be given once and must not be empty. The all-clusters parameter is not accepted, as a write always
// applies to one cluster.
func ParseSourceCluster(rawQuery string) (string, error) {
	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		return "", api.StatusErrorf(http.StatusBadRequest, "Invalid query: %v", err)
	}

	_, hasAllClusters := query[AllClustersQueryParam]
	if hasAllClusters {
		return "", api.StatusErrorf(http.StatusBadRequest, "Query parameter %q is only supported when listing cluster links", AllClustersQueryParam)
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
