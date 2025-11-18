package config

import (
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/canonical/lxd/shared"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database"
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
	TestMode               bool
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
	OIDCClientID string
	OIDCIssuer   string
	OIDCAudience string
	// cos configs
	GrafanaBaseURL    string
	PrometheusBaseURL string
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
	oidcIssuer := os.Getenv("OIDC_ISSUER")
	oidcAudience := os.Getenv("OIDC_AUDIENCE")
	if oidcClientID == "" || oidcIssuer == "" {
		return nil, fmt.Errorf("OIDC_CLIENT_ID and OIDC_ISSUER are required")
	}

	grafanaBaseURL := os.Getenv("GRAFANA_BASE_URL")
	prometheusBaseURL := os.Getenv("PROMETHEUS_BASE_URL")

	return &Config{
		Version:                getEnvOrDefault("VERSION", "development"),
		APIVersion:             getEnvOrDefault("API_VERSION", "1.0"),
		ServerHost:             getEnvOrDefault("SERVER_HOST", "0.0.0.0"),
		ServerPort:             getEnvOrDefault("SERVER_PORT", "9000"),
		StatusPort:             getEnvOrDefault("STATUS_PORT", "10000"),
		TestMode:               getEnvOrDefault("TEST_MODE", "false") == "true",
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
		OIDCIssuer:        oidcIssuer,
		OIDCAudience:      oidcAudience,
		GrafanaBaseURL:    grafanaBaseURL,
		PrometheusBaseURL: prometheusBaseURL,
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
