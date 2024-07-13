package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/canonical/microcluster/microcluster"
	"github.com/spf13/cobra"

	"github.com/canonical/lxd-site-manager/internal/api"
	"github.com/canonical/lxd-site-manager/version"
)

type cmdInit struct {
	common *CmdControl

	flagBootstrap      bool
	flagToken          string
	flagConfig         []string
	flagControlAddress string
}

func (c *cmdInit) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init <name> <address>",
		Short: "Initialize the network endpoint and create or join a new cluster",
		RunE:  c.run,
		Example: `
			lxd-site-mgr init member1 0.0.0.0:9001 --bootstrap
    		lxd-site-mgr init member1 0.0.0.0:9001 --token <token>
		`,
	}

	cmd.Flags().BoolVar(&c.flagBootstrap, "bootstrap", false, "Configure a new cluster with this daemon")
	cmd.Flags().StringVar(&c.flagToken, "token", "", "Join a cluster with a join token")
	cmd.Flags().StringVar(&c.flagControlAddress, "control-address", "", "Specify the address of the control network listener")
	cmd.Flags().StringSliceVar(&c.flagConfig, "config", nil, "Extra configuration to be applied during bootstrap")
	cmd.MarkFlagsMutuallyExclusive("bootstrap", "token")

	return cmd
}

func (c *cmdInit) run(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return cmd.Help()
	}

	m, err := microcluster.App(microcluster.Args{
		StateDir: c.common.FlagStateDir,
		Verbose:  c.common.FlagLogVerbose,
		Debug:    c.common.FlagLogDebug,
		Version:  version.Version(),
	})

	if err != nil {
		return fmt.Errorf("Unable to configure LXD site manager cluster: %w", err)
	}

	conf := make(map[string]string, len(c.flagConfig))
	for _, setting := range c.flagConfig {
		key, value, ok := strings.Cut(setting, "=")
		if !ok {
			return fmt.Errorf("Malformed additional configuration value %s", setting)
		}

		conf[key] = value
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
	defer cancel()

	if c.flagControlAddress == "" {
		return fmt.Errorf("Control address must be specified")
	}

	conf[string(api.ControlListener)] = c.flagControlAddress

	if c.flagBootstrap {
		return m.NewCluster(ctx, args[0], args[1], conf)
	}

	if c.flagToken != "" {
		return m.JoinCluster(ctx, args[0], args[1], c.flagToken, conf)
	}

	return fmt.Errorf("Option must be one of bootstrap or token")
}
