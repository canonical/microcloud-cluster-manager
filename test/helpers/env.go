package helpers

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/canonical/lxd/shared"
)

const (
	// Succeed is a checkmark symbol.
	Succeed = "\u2713"
	// Failed is a cross symbol.
	Failed = "\u2717"
)

// Environment represents the environment for the tests.
type Environment struct {
	rootDir                   string
	testDir                   string
	certDir                   string
	processIDs                []int
	managementAPICert         *shared.CertInfo
	clusterConnectorCert      *shared.CertInfo
	managementAPIHost         string
	managementAPIPort         int
	managementAPILoginHeaders func(*http.Request) error
	clusterConnectorHost      string
	clusterConnectorPort      int
	remoteClusters            []string
	remoteClusterTokens       []string
}

// NewEnv creates a new environment.
func NewEnv() *Environment {
	return &Environment{
		rootDir:                   getProjectRoot(),
		testDir:                   "",
		certDir:                   "",
		managementAPIHost:         "ma.lxd-cm.local",
		managementAPIPort:         30000,
		managementAPILoginHeaders: nil,
		clusterConnectorHost:      "cc.lxd-cm.local",
		clusterConnectorPort:      getClusterConnectorPort(),
	}
}

// RootDir returns the root directory of the project.
func (e *Environment) RootDir() string {
	return e.rootDir
}

// DataDir returns the data directory for tests.
func (e *Environment) DataDir() string {
	return e.rootDir + "/test/e2e/data"
}

// CertDir returns the certificate directory.
func (e *Environment) CertDir() string {
	return e.rootDir + "/test/e2e/data/certs"
}

// AddProcessID adds a process ID to the environment.
func (e *Environment) AddProcessID(pid int) {
	e.processIDs = append(e.processIDs, pid)
}

// ProcessIDs returns the pids for all processes started during a test run.
func (e *Environment) ProcessIDs() []int {
	return e.processIDs
}

// RemoveRemoteCluster mark a remote cluster for removal during test cleanup.
func (e *Environment) RemoveRemoteCluster(name string) {
	e.remoteClusters = append(e.remoteClusters, name)
}

// RemoveRemoteClusterToken mark a remote cluster join token for removal during test cleanup.
func (e *Environment) RemoveRemoteClusterToken(token string) {
	e.remoteClusterTokens = append(e.remoteClusterTokens, token)
}

// Init initializes the environment.
func (e *Environment) Init() error {
	// cleanup data dir in case if test didn't exit properly last time
	if err := os.RemoveAll(e.DataDir()); err != nil {
		return err
	}

	err := e.setCertificates()
	if err != nil {
		return err
	}

	return nil
}

// Cleanup cleans up the environment.
func (e *Environment) Cleanup() error {
	pids := e.ProcessIDs()
	for _, pid := range pids {
		killCmd := exec.Command("kill", "-9", strconv.Itoa(pid))
		if err := killCmd.Run(); err != nil {
			continue
		}
	}

	if err := os.RemoveAll(e.DataDir()); err != nil {
		return err
	}

	// remove all new clusters and tokens created during the test
	for _, cluster := range e.remoteClusters {
		err := DeleteRemoteCluster(e, cluster)
		if err != nil {
			continue
		}
	}

	for _, token := range e.remoteClusterTokens {
		err := DeleteRemoteClusterJoinToken(e, token)
		if err != nil {
			continue
		}
	}

	return nil
}

// ManagementAPICert returns the management-api certificate.
func (e *Environment) ManagementAPICert() *shared.CertInfo {
	return e.managementAPICert
}

// ManagementAPILoginHeaders performs a cached login to the management-api and returns
// a function that adds the authentication headers to an HTTP request.
func (e *Environment) ManagementAPILoginHeaders() (func(*http.Request) error, error) {
	if e.managementAPILoginHeaders == nil {
		username := os.Getenv("OIDC_USER")
		if username == "" {
			username = "cluster-manager-e2e-tests@example.org"
		}
		password := os.Getenv("OIDC_PASSWORD")
		if password == "" {
			password = "cluster-manager-e2e-password"
		}

		cookies, err := LoginToManagementAPI(e, username, password)
		if err != nil {
			return nil, err
		}

		e.managementAPILoginHeaders = func(req *http.Request) error {
			for _, cookie := range cookies {
				req.AddCookie(cookie)
			}
			return nil
		}
	}
	return e.managementAPILoginHeaders, nil
}

// ClusterConnectorCert returns the cluster-connector certificate.
func (e *Environment) ClusterConnectorCert() *shared.CertInfo {
	return e.clusterConnectorCert
}

// ManagementAPIHost returns the management-api host.
func (e *Environment) ManagementAPIHost() string {
	return e.managementAPIHost
}

// ClusterConnectorHost returns the cluster-connector host.
func (e *Environment) ClusterConnectorHost() string {
	return e.clusterConnectorHost
}

// ManagementAPIHostPort returns the management-api host and port.
func (e *Environment) ManagementAPIHostPort() string {
	return fmt.Sprintf("%s:%d", e.managementAPIHost, e.managementAPIPort)
}

// ClusterConnectorHostPort returns the cluster-connector host and port.
func (e *Environment) ClusterConnectorHostPort() string {
	return fmt.Sprintf("%s:%d", e.clusterConnectorHost, e.clusterConnectorPort)
}

func (e *Environment) setCertificates() error {
	// Helper function to retrieve and validate certificate data from a secret
	getCertificateData := func(secretName string) (cert, key, ca []byte, err error) {
		secretPathEnvVar := strings.ReplaceAll(secretName, "-", "_")
		secretPath := os.Getenv(secretPathEnvVar)
		cert, certErr := os.ReadFile(filepath.Join(secretPath, "tls.crt"))
		key, keyErr := os.ReadFile(filepath.Join(secretPath, "tls.key"))
		ca, caErr := os.ReadFile(filepath.Join(secretPath, "ca.crt"))

		if certErr != nil || keyErr != nil || caErr != nil {
			return nil, nil, nil, fmt.Errorf("secret path %q for secret name %q does not contain all required keys", secretPath, secretName)
		}

		return cert, key, ca, nil
	}

	// Helper function to write a file and handle errors
	getCertInfo := func(certPem []byte, keyPem []byte, caPem []byte) (*shared.CertInfo, error) {
		cert, err := tls.X509KeyPair(certPem, keyPem)
		if err != nil {
			return nil, fmt.Errorf("could not load key pair: %w", err)
		}

		ca, err := shared.ParseCert(caPem)
		if err != nil {
			return nil, fmt.Errorf("could not parse CA certificate: %w", err)
		}

		return shared.NewCertInfo(cert, ca, nil), nil
	}

	// Fetch management-api and cluster-connector certificate data
	managementAPICert, managementAPIKey, managementAPICA, err := getCertificateData("management-api-cert-secret")
	if err != nil {
		return err
	}

	certInfo, err := getCertInfo(managementAPICert, managementAPIKey, managementAPICA)
	if err != nil {
		return err
	}
	e.managementAPICert = certInfo

	clusterConnectorCert, clusterConnectorKey, clusterConnectorCA, err := getCertificateData("cluster-connector-cert-secret")
	if err != nil {
		return err
	}
	certInfo, err = getCertInfo(clusterConnectorCert, clusterConnectorKey, clusterConnectorCA)
	if err != nil {
		return err
	}
	e.clusterConnectorCert = certInfo

	return nil
}

func getProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}

		// Move to the parent directory, Dir excludes the last file path element
		parent := filepath.Dir(wd)
		if parent == wd {
			// Reached the root of the filesystem
			break
		}

		wd = parent
	}

	panic("could not find go.mod")
}

func getClusterConnectorPort() int {
	value := os.Getenv("CLUSTER_CONNECTOR_PORT")
	if value != "" {
		i, err := strconv.Atoi(value)
		if err == nil {
			return i
		}
	}
	return 30000
}
