package types

// StatusDistribution is a status and count pair used by the remote cluster status endpoint.
type StatusDistribution struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}
