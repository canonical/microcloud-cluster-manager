package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/canonical/microcloud-cluster-manager/cmd/cli"
	clusterconnector "github.com/canonical/microcloud-cluster-manager/cmd/cluster-connector"
	managementapi "github.com/canonical/microcloud-cluster-manager/cmd/management-api"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/logger"
)

// VERSION represents the current version of the application.
var VERSION = "0.1"

// SERVICES is a list of valid service names within the cluster manager application.
var SERVICES = []string{
	"management-api",
	"cluster-connector",
	"cli",
}

func main() {
	// Perform the startup and shutdown sequence.
	err := run()
	if err != nil {
		logger.Log.Errorw("startup", "ERROR", err)
		err = logger.Log.Sync()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to sync logger: %v\n", err)
		}
		os.Exit(1)
	}

	_ = logger.Cleanup()
}

func run() error {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(VERSION)
		os.Exit(0)
	}

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

	if service == "cli" {
		return cli.Run()
	}

	return fmt.Errorf("service name is invalid, it should be one of: %v", SERVICES)
}
