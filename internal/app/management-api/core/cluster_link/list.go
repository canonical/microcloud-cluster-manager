package cluster_link

import (
	"time"
)

const (
	// ListClusterTimeout bounds the time spent on one cluster so that a hung tunnel cannot stall the list.
	ListClusterTimeout = 10 * time.Second
	// ListTotalTimeout bounds the time spent on all clusters together.
	ListTotalTimeout = 15 * time.Second
	// ListConcurrencyLimit is the max number of clusters queried at the same time.
	ListConcurrencyLimit = 40
	// MaxResponseBytes limits the size of a cluster connector response for one cluster.
	MaxResponseBytes = 4 << 20 // 4 MiB

	// listResponseMargin is the time kept free for writing the response before the server write timeout.
	listResponseMargin = time.Second
)

// ListOptions bounds the time and concurrency of listing the cluster links of all clusters.
type ListOptions struct {
	// Deadline bounds the time spent on all clusters together. Zero means no deadline.
	Deadline time.Duration
	// ClusterTimeout bounds the time spent on one cluster. Zero means no timeout.
	ClusterTimeout time.Duration
	// Concurrency is the max number of clusters queried at the same time. Zero or less means no limit.
	Concurrency int
}

// DefaultListOptions returns the list options for a server with the given write timeout.
func DefaultListOptions(writeTimeout time.Duration) ListOptions {
	return ListOptions{
		Deadline:       ListDeadline(writeTimeout),
		ClusterTimeout: ListClusterTimeout,
		Concurrency:    ListConcurrencyLimit,
	}
}

// ListDeadline returns how long listing the cluster links of all clusters may take together. Clusters are queried in
// batches, so without it enough slow clusters would make the request outlast the server and the client would get no
// response at all instead of the links that could be loaded.
func ListDeadline(writeTimeout time.Duration) time.Duration {
	deadline := ListTotalTimeout

	// The server closes the connection once its write timeout passed, so the response must be ready before that.
	if writeTimeout > 0 {
		// For very short write timeouts, reserve half the budget so the deadline stays positive.
		margin := min(listResponseMargin, writeTimeout/2)
		deadline = min(deadline, writeTimeout-margin)
	}

	return deadline
}
