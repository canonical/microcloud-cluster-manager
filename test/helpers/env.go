package helpers

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/canonical/lxd/shared"
	"github.com/canonical/lxd/shared/api"
)

const (
	// Succeed is a checkmark symbol.
	Succeed = "\u2713"
	// Failed is a cross symbol.
	Failed = "\u2717"
)

// Environment represents the environment for the tests.
type Environment struct {
	rootDir          string
	testDir          string
	stateDirRoot     string
	stateDir         string
	certDir          string
	controlSocketURL api.URL
	daemonPath       string
	cliPath          string
	processIDs       []int
}

// NewEnv creates a new environment.
func NewEnv() *Environment {
	return &Environment{
		rootDir:          getProjectRoot(),
		testDir:          "",
		stateDirRoot:     "",
		stateDir:         "",
		certDir:          "",
		controlSocketURL: api.URL{},
		daemonPath:       "",
		cliPath:          "",
	}
}

// RootDir returns the root directory of the project.
func (e *Environment) RootDir() string {
	return e.rootDir
}

// TestDir returns the test directory.
func (e *Environment) TestDir() string {
	return e.rootDir + "/test/e2e"
}

// StateDirRoot returns the root directory for the state.
func (e *Environment) StateDirRoot() string {
	return e.rootDir + "/test/e2e/state"
}

// StateDir returns the state directory.
func (e *Environment) StateDir(path *string) string {
	if path != nil {
		e.stateDir = filepath.Join(e.StateDirRoot(), *path)
	}

	return e.stateDir
}

// CertDir returns the certificate directory.
func (e *Environment) CertDir() string {
	return e.rootDir + "/test/e2e/certs"
}

// ControlSocketURL returns the control socket URL.
func (e *Environment) ControlSocketURL() api.URL {
	path := filepath.Join(e.StateDir(nil), "control.socket")
	e.controlSocketURL = *api.NewURL().Scheme("http").Host(path)
	return e.controlSocketURL
}

// DaemonPath returns the path to the daemon.
func (e *Environment) DaemonPath() string {
	return e.rootDir + "/cmd/lxd-site-mgrd"
}

// CLIPath returns the path to the CLI.
func (e *Environment) CLIPath() string {
	return e.rootDir + "/cmd/lxd-site-mgr"
}

// AddProcessID adds a process ID to the environment.
func (e *Environment) AddProcessID(pid int) {
	e.processIDs = append(e.processIDs, pid)
}

// ProcessIDs returns the pids for all processes started during a test run.
func (e *Environment) ProcessIDs() []int {
	return e.processIDs
}

// Init initializes the environment.
func (e *Environment) Init() error {
	// cleanup state dir in case if test didn't exit properly last time
	if err := os.RemoveAll(e.StateDirRoot()); err != nil {
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

	if err := os.RemoveAll(e.StateDirRoot()); err != nil {
		return err
	}

	if err := os.RemoveAll(filepath.Join(e.TestDir(), "certs")); err != nil {
		return err
	}

	if err := os.Remove(filepath.Join(e.RootDir(), "lxd-site-mgrd")); err != nil {
		return err
	}

	if err := os.Remove(filepath.Join(e.RootDir(), "lxd-site-mgr")); err != nil {
		return err
	}

	return nil
}

// GetClusterCert returns the cluster certificate from the state directory within the test environment.
func (e *Environment) GetClusterCert() (*shared.CertInfo, error) {
	cert, err := shared.KeyPairAndCA(e.StateDir(nil), "cluster", shared.CertServer, false)
	if err != nil {
		return nil, err
	}

	return cert, nil
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
