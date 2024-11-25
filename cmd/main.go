package main

import (
	"fmt"
	"os"

	"github.com/canonical/lxd-cluster-manager/cmd/admin"
	"github.com/canonical/lxd-cluster-manager/cmd/control"
	"github.com/canonical/lxd-cluster-manager/cmd/management"
	"github.com/canonical/lxd-cluster-manager/internal/pkg/logger"
)

var SERVICES = []string{
	"management",
	"control",
	"admin",
}

func main() {
	// Perform the startup and shutdown sequence.
	err := run()
	if err != nil {
		logger.Log.Errorw("startup", "ERROR", err)
		logger.Log.Sync()
		os.Exit(1)
	}

	logger.Cleanup()
}

func run() error {
	// Get the service name from environment
	service := os.Getenv("SERVICE")
	if service == "" {
		return fmt.Errorf("service name is required, it should be one of: %v", SERVICES)
	}

	// Initialize the logger for the service
	logger.SetService(service)

	if service == "management" {
		return management.Run()
	}

	if service == "control" {
		return control.Run()
	}

	if service == "admin" {
		return admin.Run()
	}

	return fmt.Errorf("service name is invalid, it should be one of: %v", SERVICES)
}
