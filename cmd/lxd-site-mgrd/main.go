// Package microd provides the daemon.
package main

import (
	"context"
	"database/sql"
	"os"

	"github.com/canonical/lxd/shared/logger"
	"github.com/canonical/microcluster/config"
	"github.com/canonical/microcluster/microcluster"
	"github.com/canonical/microcluster/state"
	"github.com/spf13/cobra"

	"github.com/canonical/lxd-site-manager/internal/api"
	"github.com/canonical/lxd-site-manager/internal/database"
	"github.com/canonical/lxd-site-manager/version"
)

// Debug indicates whether to log debug messages or not.
var Debug bool

// Verbose indicates verbosity.
var Verbose bool

type cmdGlobal struct {
	cmd *cobra.Command //nolint:structcheck,unused // FIXME: Remove the nolint flag when this is in use.

	flagHelp    bool
	flagVersion bool

	flagLogDebug   bool
	flagLogVerbose bool
}

func (c *cmdGlobal) run(cmd *cobra.Command, args []string) error {
	Debug = c.flagLogDebug
	Verbose = c.flagLogVerbose

	return logger.InitLogger("", "", c.flagLogVerbose, c.flagLogDebug, nil)
}

type cmdDaemon struct {
	global *cmdGlobal

	flagStateDir    string
	flagSocketGroup string
}

func (c *cmdDaemon) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "microd",
		Short:   "Example daemon for MicroCluster - This will start a daemon with a running control socket and no database",
		Version: version.Version,
	}

	cmd.RunE = c.run

	return cmd
}

func (c *cmdDaemon) run(cmd *cobra.Command, args []string) error {
	err := c.global.run(cmd, args)
	if err != nil {
		return err
	}

	m, err := microcluster.App(microcluster.Args{
		StateDir:         c.flagStateDir,
		SocketGroup:      c.flagSocketGroup,
		Verbose:          c.global.flagLogVerbose,
		Debug:            c.global.flagLogDebug,
		ExtensionServers: api.Servers,
	})

	if err != nil {
		return err
	}

	hooks := &config.Hooks{
		PostBootstrap: func(s *state.State, initConfig map[string]string) error {
			return s.Database.Transaction(context.TODO(), func(ctx context.Context, tx *sql.Tx) error {
				testStmt := `
					INSERT INTO core_sites (name, site_certificate) VALUES ('site1', 'a');
					INSERT INTO core_sites (name, site_certificate) VALUES ('site2', 'as');
					INSERT INTO core_sites (name, site_certificate) VALUES ('site3', 'asd');
					INSERT INTO site_details (core_site_id, status, instance_statuses) VALUES (1, 'PENDING_APPROVAL', '{}');
					INSERT INTO site_details (core_site_id, status, instance_statuses) VALUES (2, 'ACTIVE', '{}');
					INSERT INTO site_details (core_site_id, status, instance_statuses) VALUES (3, 'ACTIVE', '{}');
					INSERT INTO site_member_statuses (core_site_id, member_name, address, architecture, role, usage_cpu, usage_memory, usage_disk, status) 
						VALUES (1, 'member1', '127.0.0.1:9001', 'x86_64', 'controller', 0.0, 0.0, 0.0, 'ACTIVE');
					INSERT INTO site_member_statuses (core_site_id, member_name, address, architecture, role, usage_cpu, usage_memory, usage_disk, status) 
						VALUES (1, 'member2', '127.0.0.1:9002', 'x86_64', 'controller', 0.0, 0.0, 0.0, 'ACTIVE');
					INSERT INTO site_member_statuses (core_site_id, member_name, address, architecture, role, usage_cpu, usage_memory, usage_disk, status) 
						VALUES (2, 'member1', '127.0.0.1:9001', 'x86_64', 'controller', 0.0, 0.0, 0.0, 'ACTIVE');
					INSERT INTO site_member_statuses (core_site_id, member_name, address, architecture, role, usage_cpu, usage_memory, usage_disk, status) 
						VALUES (2, 'member2', '127.0.0.1:9002', 'x86_64', 'controller', 0.0, 0.0, 0.0, 'ACTIVE');
					INSERT INTO site_member_statuses (core_site_id, member_name, address, architecture, role, usage_cpu, usage_memory, usage_disk, status) 
						VALUES (3, 'member1', '127.0.0.1:9001', 'x86_64', 'controller', 0.0, 0.0, 0.0, 'ACTIVE');
					INSERT INTO site_member_statuses (core_site_id, member_name, address, architecture, role, usage_cpu, usage_memory, usage_disk, status) 
						VALUES (3, 'member2', '127.0.0.1:9002', 'x86_64', 'controller', 0.0, 0.0, 0.0, 'ACTIVE');
					
				`
				_, err := tx.ExecContext(ctx, testStmt)
				return err
			})
		},
	}

	return m.Start(cmd.Context(), database.SchemaExtensions, api.Extensions(), hooks)
}

func main() {
	daemonCmd := cmdDaemon{global: &cmdGlobal{}}
	app := daemonCmd.command()
	app.SilenceUsage = true
	app.CompletionOptions = cobra.CompletionOptions{DisableDefaultCmd: true}

	app.PersistentFlags().BoolVarP(&daemonCmd.global.flagHelp, "help", "h", false, "Print help")
	app.PersistentFlags().BoolVar(&daemonCmd.global.flagVersion, "version", false, "Print version number")
	app.PersistentFlags().BoolVarP(&daemonCmd.global.flagLogDebug, "debug", "d", false, "Show all debug messages")
	app.PersistentFlags().BoolVarP(&daemonCmd.global.flagLogVerbose, "verbose", "v", false, "Show all information messages")

	app.PersistentFlags().StringVar(&daemonCmd.flagStateDir, "state-dir", "", "Path to store state information"+"``")
	app.PersistentFlags().StringVar(&daemonCmd.flagSocketGroup, "socket-group", "", "Group to set socket's group ownership to")

	app.SetVersionTemplate("{{.Version}}\n")

	err := app.Execute()
	if err != nil {
		os.Exit(1)
	}
}
