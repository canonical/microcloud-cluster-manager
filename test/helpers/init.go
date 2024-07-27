package helpers

import (
	"os/exec"
	"path/filepath"
	"testing"
)

// BuildBinaries builds the daemon and CLI binaries.
func BuildBinaries(env *Environment, t *testing.T) {
	makeCmd := exec.Command("make", "compile")
	makeCmd.Dir = env.RootDir()
	err := makeCmd.Run()
	LogTestOutcome(t, "Build binaries", err)
}

// InitSingleMemberCluster initializes a single member cluster for testing Cluster Manager.
func InitSingleMemberCluster(env *Environment, t *testing.T) {
	memberName := "member1"
	coreAddress := "0.0.0.0:9100"
	controlAddress := "0.0.0.0:9110"
	stateDir := env.StateDir(&memberName)
	daemonBinary := filepath.Join(env.RootDir(), "lxd-cluster-mgrd")
	cliBinary := filepath.Join(env.RootDir(), "lxd-cluster-mgr")

	// Start the daemon
	daemonCmd := exec.Command(daemonBinary, "--state-dir", stateDir)
	err := daemonCmd.Start()
	LogTestOutcome(t, "Start daemon", err)
	env.AddProcessID(daemonCmd.Process.Pid)

	// Wait until the daemon is ready
	cliCmd := exec.Command(cliBinary, "--state-dir", stateDir, "waitready")
	err = cliCmd.Run()
	LogTestOutcome(t, "Wait for daemon to be ready", err)

	// initialise the cluster
	cliCmd = exec.Command(cliBinary, "--state-dir", stateDir, "init", memberName, coreAddress, "--bootstrap", "--control-address", controlAddress)
	err = cliCmd.Run()
	LogTestOutcome(t, "Initialize cluster", err)
}
