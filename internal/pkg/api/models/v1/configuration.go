package models

// ConfigData is a struct that holds the configuration data.
type ConfigData struct {
	Value       string `json:"value"`
	Description string `json:"description"`
	Title       string `json:"title"`
}

// Configuration represents application environment data shown to users.
type Configuration struct {
	APIVersion                    ConfigData `json:"api_version"`
	Version                       ConfigData `json:"version"`
	ClusterConnectorDomain        ConfigData `json:"cluster_connector_domain"`
	ClusterConnectorPort          ConfigData `json:"cluster_connector_port"`
	OIDCClientID                  ConfigData `json:"oidc_client_id"`
	OIDCClientSecret              ConfigData `json:"oidc_client_secret"`
	OIDCIssuer                    ConfigData `json:"oidc_issuer"`
	OIDCAudience                  ConfigData `json:"oidc_audience"`
	DBConnectionString            ConfigData `json:"db_connection_string"`
	DBMaxIdleConns                ConfigData `json:"db_max_idle_conns"`
	DBMaxOpenConns                ConfigData `json:"db_max_open_conns"`
	GrafanaBaseURL                ConfigData `json:"grafana_base_url"`
	PrometheusBaseURL             ConfigData `json:"prometheus_base_url"`
	RateLimitRefillRate           ConfigData `json:"rate_limit_refill_rate"`
	RateLimitBucketSize           ConfigData `json:"rate_limit_bucket_size"`
	RateLimitClientActiveInterval ConfigData `json:"rate_limit_client_active_interval"`
	RateLimitCleanupInterval      ConfigData `json:"rate_limit_cleanup_interval"`
	RateLimitLogInterval          ConfigData `json:"rate_limit_log_interval"`
	RateLimitMaxClients           ConfigData `json:"rate_limit_max_clients"`
}
