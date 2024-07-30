package main

import (
	"github.com/canonical/microcluster/microcluster"
	"github.com/spf13/cobra"

	"github.com/canonical/lxd-cluster-manager/version"
)

type cmdShutdown struct {
	common *CmdControl
}

func (c *cmdShutdown) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shutdown",
		Short: "Shutdown LXD Cluster Manager daemon",
		RunE:  c.run,
	}

	return cmd
}

func (c *cmdShutdown) run(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return cmd.Help()
	}

	m, err := microcluster.App(microcluster.Args{
		StateDir: c.common.FlagStateDir,
		Verbose:  c.common.FlagLogVerbose,
		Debug:    c.common.FlagLogDebug,
		Version:  version.Version(),
	})

	if err != nil {
		return err
	}

	client, err := m.LocalClient()
	if err != nil {
		return err
	}

	chResult := make(chan error, 1)
	go func() {
		defer close(chResult)

		err := client.ShutdownDaemon(cmd.Context())
		if err != nil {
			chResult <- err
			return
		}
	}()

	return <-chResult
}
