package config

import (
	"crypto/tls"
	"fmt"
	"os"
	"strconv"

	"github.com/canonical/lxd-cluster-manager/internal/pkg/database"
)

type Config struct {
	// system configs
	Version    string
	ServerCert tls.Certificate
	// db configs
	database.DBConfig
	// api configs
	ServerHost     string
	ManagementPort string
	ControlPort    string
	AllowedOrigins []string
	ReadTimeout    int
	WriteTimeout   int
	IdleTimeout    int
}

func getNumConfig(key string, defaultValue int) (int, error) {
	var val int
	var err error
	enVar := os.Getenv(key)
	if enVar == "" {
		val = defaultValue
	} else {
		val, err = strconv.Atoi(enVar)
		if err != nil {
			return 0, err
		}
	}

	return val, nil
}

func getServiceCert() (tls.Certificate, error) {
	tlsCertPath := os.Getenv("TLS_CERT_PATH")
	if tlsCertPath == "" {
		return tls.Certificate{}, fmt.Errorf("TLS_CERT_PATH is required")
	}

	tlsKeyPath := os.Getenv("TLS_KEY_PATH")
	if tlsKeyPath == "" {
		return tls.Certificate{}, fmt.Errorf("TLS_KEY_PATH is required")
	}

	cert, err := tls.LoadX509KeyPair(tlsCertPath, tlsKeyPath)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("could not load TLS key pair: %w", err)
	}

	return cert, nil
}

// LoadConfig loads the configuration from the environment variables.
func LoadConfig(requireCerts bool) (*Config, error) {
	version := os.Getenv("VERSION")
	if version == "" {
		version = "development"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432" // default port
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "admin"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "admin"
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "cm"
	}

	dbMaxIdleConns, err := getNumConfig("DB_MAX_IDLE", 10)
	if err != nil {
		return nil, err
	}

	dbMaxOpenConns, err := getNumConfig("DB_MAX_OPEN", 2)
	if err != nil {
		return nil, err
	}

	dbDisableTLS := os.Getenv("DB_DISABLE_TLS")
	if dbDisableTLS == "" {
		dbDisableTLS = "true"
	}

	serverHost := os.Getenv("SERVER_HOST")
	if serverHost == "" {
		serverHost = "localhost"
	}

	managementPort := os.Getenv("MANAGEMENT_PORT")
	if managementPort == "" {
		managementPort = "9000"
	}

	controlPort := os.Getenv("CONTROL_PORT")
	if controlPort == "" {
		controlPort = "9001"
	}

	var serverCert tls.Certificate
	if requireCerts {
		serverCert, err = getServiceCert()
		if err != nil {
			return nil, err
		}
	}

	return &Config{
		Version:        version,
		ServerHost:     serverHost,
		ManagementPort: managementPort,
		ControlPort:    controlPort,
		AllowedOrigins: []string{"*"}, // Configure as needed
		ReadTimeout:    10,            // in seconds
		WriteTimeout:   10,            // in seconds
		IdleTimeout:    60,            // in seconds
		ServerCert:     serverCert,
		DBConfig: database.DBConfig{
			DBPort:         dbPort,
			DBUser:         dbUser,
			DBPassword:     dbPassword,
			DBHost:         dbHost,
			DBName:         dbName,
			DBMaxIdleConns: dbMaxIdleConns,
			DBMaxOpenConns: dbMaxOpenConns,
			DBDisableTLS:   dbDisableTLS == "true",
		},
	}, nil
}
