package helpers

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/canonical/lxd/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	// Succeed is a checkmark symbol.
	Succeed = "\u2713"
	// Failed is a cross symbol.
	Failed = "\u2717"
)

// Environment represents the environment for the tests.
type Environment struct {
	rootDir              string
	testDir              string
	certDir              string
	processIDs           []int
	kClient              *kubernetes.Clientset
	managementAPICert    *shared.CertInfo
	clusterConnectorCert *shared.CertInfo
	managementAPIHost    string
	clusterConnectorHost string
	remoteClusters       []string
	remoteClusterTokens  []string
}

// NewEnv creates a new environment.
func NewEnv() *Environment {
	return &Environment{
		rootDir:              getProjectRoot(),
		testDir:              "",
		certDir:              "",
		managementAPIHost:    "localhost:9000",
		clusterConnectorHost: "localhost:9001",
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

	err := e.setKubeClient()
	if err != nil {
		return err
	}

	err = e.setCertificates()
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

func (e *Environment) setCertificates() error {
	// Helper function to retrieve and validate certificate data from a secret
	getCertificateData := func(secretName string) (cert, key, ca []byte, err error) {
		secret, err := e.kClient.CoreV1().Secrets("default").Get(context.TODO(), secretName, metav1.GetOptions{})
		if err != nil {
			return nil, nil, nil, fmt.Errorf("could not get secret %q: %w", secretName, err)
		}

		cert, certExist := secret.Data["tls.crt"]
		key, keyExist := secret.Data["tls.key"]
		ca, caExist := secret.Data["ca.crt"]

		if !certExist || !keyExist || !caExist {
			return nil, nil, nil, fmt.Errorf("secret %q does not contain all required keys", secretName)
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

func (e *Environment) setKubeClient() error {
	home := homedir.HomeDir()
	if home == "" {
		return fmt.Errorf("could not find home directory for kubeconfig")
	}

	kubeconfig := filepath.Join(home, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return fmt.Errorf("could not build kubeconfig: %v", err)
	}

	kClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("could not create kubernetes client: %v", err)
	}

	e.kClient = kClient
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
