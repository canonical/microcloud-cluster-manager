package config

import (
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/canonical/lxd/shared"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database"
	"golang.org/x/time/rate"
)

// Config represents the configurable environment variables for all services within the application.
type Config struct {
	// system configs
	Version                string
	APIVersion             string
	ManagementAPICert      *shared.CertInfo
	ClusterConnectorCert   *shared.CertInfo
	ClusterConnectorDomain string
	ClusterConnectorPort   string
	// db configs
	database.DBConfig
	// api configs
	ServerHost     string
	ServerPort     string
	StatusPort     string
	AllowedOrigins []string
	ReadTimeout    int
	WriteTimeout   int
	IdleTimeout    int
	// oidc configs
	OIDCClientID     string
	OIDCClientSecret string
	OIDCIssuer       string
	OIDCAudience     string
	// cos configs
	GrafanaBaseURL    string
	PrometheusBaseURL string
	// rate limit configs
	RateLimitRefillRate           rate.Limit
	RateLimitBucketSize           int
	RateLimitClientActiveInterval time.Duration
	RateLimitCleanupInterval      time.Duration
	RateLimitLogInterval          time.Duration
	RateLimitMaxClients           int
}

// getEnvOrDefault retrieves an environment variable or returns the default value if not set.
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt retrieves an environment variable as an integer or returns the default value.
func getEnvAsInt(key string, defaultValue int) (int, error) {
	value := os.Getenv(key)
	if value != "" {
		return strconv.Atoi(value)
	}
	return defaultValue, nil
}

// getRequiredEnv retrieves a required environment variable.
func getRequiredEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return value, nil
}

// getRequiredInt retrieves a required environment variable as an integer.
func getRequiredInt(key string) (int, error) {
	value, err := getRequiredEnv(key)
	if err != nil {
		return 0, err
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	return parsed, nil
}

// getRequiredFloat retrieves a required environment variable as a float.
func getRequiredFloat(key string) (float64, error) {
	value, err := getRequiredEnv(key)
	if err != nil {
		return 0, err
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	return parsed, nil
}

// getServiceCert loads the TLS certificate and key from the environment based on the service name.
func getServiceCert(service string) (*shared.CertInfo, error) {
	if service != "management-api" && service != "cluster-connector" {
		return nil, fmt.Errorf("invalid service name: %s", service)
	}

	// convert service to underscore
	service = strings.ReplaceAll(service, "-", "_")
	key := strings.ToUpper(service) + "_TLS_PATH"
	tlsPath := os.Getenv(key)

	if tlsPath == "" {
		return nil, fmt.Errorf("missing config %s", key)
	}

	certPath := filepath.Join(tlsPath, "tls.crt")
	keyPath := filepath.Join(tlsPath, "tls.key")
	caPath := filepath.Join(tlsPath, "ca.crt")

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificate: %w", err)
	}

	ca, err := shared.ReadCert(caPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load CA certificate: %w", err)
	}

	return shared.NewCertInfo(cert, ca, nil), nil
}

// LoadConfig loads the configuration from the environment variables.
func LoadConfig() (*Config, error) {
	// DB Config
	dbMaxIdleConns, err := getEnvAsInt("DB_MAX_IDLE", 10)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_IDLE: %w", err)
	}
	dbMaxOpenConns, err := getEnvAsInt("DB_MAX_OPEN", 2)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_OPEN: %w", err)
	}

	// OIDC Config
	oidcClientID := os.Getenv("OIDC_CLIENT_ID")
	oidcClientSecret := os.Getenv("OIDC_CLIENT_SECRET")
	oidcIssuer := os.Getenv("OIDC_ISSUER")
	oidcAudience := os.Getenv("OIDC_AUDIENCE")
	if oidcClientID == "" || oidcIssuer == "" {
		return nil, fmt.Errorf("OIDC_CLIENT_ID and OIDC_ISSUER are required")
	}

	// cos Config
	grafanaBaseURL := os.Getenv("GRAFANA_BASE_URL")
	prometheusBaseURL := os.Getenv("PROMETHEUS_BASE_URL")

	// Rate Limit Config
	rateLimitRefillRate, err := getRequiredFloat("RATE_LIMIT_REFILL_RATE")
	if err != nil {
		return nil, err
	}

	rateLimitBucketSize, err := getRequiredInt("RATE_LIMIT_BUCKET_SIZE")
	if err != nil {
		return nil, err
	}

	rateLimitClientActiveIntervalSeconds, err := getRequiredInt("RATE_LIMIT_CLIENT_ACTIVE_INTERVAL_SECONDS")
	if err != nil {
		return nil, err
	}

	rateLimitCleanupIntervalSeconds, err := getRequiredInt("RATE_LIMIT_CLEANUP_INTERVAL_SECONDS")
	if err != nil {
		return nil, err
	}

	rateLimitLogIntervalSeconds, err := getRequiredInt("RATE_LIMIT_LOG_INTERVAL_SECONDS")
	if err != nil {
		return nil, err
	}

	rateLimitMaxClients, err := getRequiredInt("RATE_LIMIT_MAX_CLIENTS")
	if err != nil {
		return nil, err
	}

	return &Config{
		Version:                getEnvOrDefault("VERSION", "development"),
		APIVersion:             getEnvOrDefault("API_VERSION", "1.0"),
		ServerHost:             getEnvOrDefault("SERVER_HOST", "0.0.0.0"),
		ServerPort:             getEnvOrDefault("SERVER_PORT", "9000"),
		StatusPort:             getEnvOrDefault("STATUS_PORT", "10000"),
		AllowedOrigins:         []string{"*"},
		ReadTimeout:            10,
		WriteTimeout:           10,
		IdleTimeout:            60,
		ClusterConnectorDomain: getEnvOrDefault("CLUSTER_CONNECTOR_DOMAIN", "cc.lxd-cm.local"),
		ClusterConnectorPort:   getEnvOrDefault("CLUSTER_CONNECTOR_PORT", "30000"),
		DBConfig: database.DBConfig{
			DBPort:         getEnvOrDefault("DB_PORT", "5432"),
			DBUser:         getEnvOrDefault("DB_USER", "admin"),
			DBPassword:     os.Getenv("DB_PASSWORD"),
			DBHost:         getEnvOrDefault("DB_HOST", "db-svc"),
			DBName:         getEnvOrDefault("DB_NAME", "cm"),
			DBMaxIdleConns: dbMaxIdleConns,
			DBMaxOpenConns: dbMaxOpenConns,
			DBDisableTLS:   getEnvOrDefault("DB_DISABLE_TLS", "true") == "true",
		},
		OIDCClientID:      oidcClientID,
		OIDCClientSecret:  oidcClientSecret,
		OIDCIssuer:        oidcIssuer,
		OIDCAudience:      oidcAudience,
		GrafanaBaseURL:    grafanaBaseURL,
		PrometheusBaseURL: prometheusBaseURL,

		RateLimitRefillRate:           rate.Limit(rateLimitRefillRate),
		RateLimitBucketSize:           rateLimitBucketSize,
		RateLimitClientActiveInterval: time.Duration(rateLimitClientActiveIntervalSeconds) * time.Second,
		RateLimitCleanupInterval:      time.Duration(rateLimitCleanupIntervalSeconds) * time.Second,
		RateLimitLogInterval:          time.Duration(rateLimitLogIntervalSeconds) * time.Second,
		RateLimitMaxClients:           rateLimitMaxClients,
	}, nil
}

// LoadCertificates loads the TLS certificates from the environment.
func (c *Config) LoadCertificates() error {
	// service certificates
	managementAPICert, err := getServiceCert("management-api")
	if err != nil {
		return fmt.Errorf("failed to load management-api certificate: %w", err)
	}

	clusterConnectorCert, err := getServiceCert("cluster-connector")
	if err != nil {
		return fmt.Errorf("failed to load cluster-connector certificate: %w", err)
	}

	c.ManagementAPICert = managementAPICert
	c.ClusterConnectorCert = clusterConnectorCert

	return nil
}
