// mcm-dev is a development and testing client for management-api.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/canonical/microcloud-cluster-manager/dev-cli/internal/client"
	"github.com/canonical/microcloud-cluster-manager/dev-cli/internal/command"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := command.New(command.Options{}).ExecuteContext(ctx)
	cancel()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(client.ExitCode(err))
	}
}
