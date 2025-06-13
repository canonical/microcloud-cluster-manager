package main

import (
	"fmt"
	"os"

	clusterconnector "github.com/canonical/microcloud-cluster-manager/cmd/cluster-connector"
	managementapi "github.com/canonical/microcloud-cluster-manager/cmd/management-api"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/logger"
)

// SERVICES is a list of valid service names within the cluster manager application.
var SERVICES = []string{
	"management-api",
	"cluster-connector",
}

func main() {
	// Perform the startup and shutdown sequence.
	err := run()
	if err != nil {
		logger.Log.Errorw("startup", "ERROR", err)
		_ = logger.Log.Sync()
		os.Exit(1)
	}

	_ = logger.Cleanup()
}

func run() error {
	// Get the service name from environment
	service := os.Getenv("SERVICE")
	if service == "" {
		return fmt.Errorf("service name is required, it should be one of: %v", SERVICES)
	}

	// Initialize the logger for the service
	logger.SetService(service)

	if service == "management-api" {
		return managementapi.Run()
	}

	if service == "cluster-connector" {
		return clusterconnector.Run()
	}

	return fmt.Errorf("service name is invalid, it should be one of: %v", SERVICES)
}
