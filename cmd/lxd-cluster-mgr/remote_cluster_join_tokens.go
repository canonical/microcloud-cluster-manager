package main

import (
	"sort"
	"time"

	lxdCmd "github.com/canonical/lxd/shared/cmd"
	"github.com/canonical/microcluster/microcluster"
	"github.com/spf13/cobra"

	"github.com/canonical/lxd-cluster-manager/internal/api/types"
	"github.com/canonical/lxd-cluster-manager/internal/client"
	"github.com/canonical/lxd-cluster-manager/version"
)

type cmdRemoteClusterJoinToken struct {
	common *CmdControl
}

func (c *cmdRemoteClusterJoinToken) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remote-cluster-join-token",
		Short: "manage join tokens for LXD clusters.",
		RunE:  c.run,
	}

	var cmdAdd = cmdRemoteClusterJoinTokenAdd{common: c.common}
	cmd.AddCommand(cmdAdd.command())

	var cmdShow = cmdRemoteClusterJoinTokenShow{common: c.common}
	cmd.AddCommand(cmdShow.command())

	var cmdRevoke = cmdRemoteClusterJoinTokenRevoke{common: c.common}
	cmd.AddCommand(cmdRevoke.command())

	return cmd
}

func (c *cmdRemoteClusterJoinToken) run(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

type cmdRemoteClusterJoinTokenAdd struct {
	common *CmdControl

	flagExpiry time.Duration
}

func (c *cmdRemoteClusterJoinTokenAdd) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <cluster_name>",
		Short: "Add a new remote join token for a LXD cluster.",
		Example: `
			lxd-cluster-mgr remote-cluster-join-token add <cluster_name>
			Will create a remote cluster join token for a LXD cluster.
		`,
		RunE: c.run,
	}

	cmd.Flags().DurationVar(&c.flagExpiry, "expiry", 0, "Specify the duration (i.e. 5m or 48h) for the token to be valid")

	return cmd
}

func (c *cmdRemoteClusterJoinTokenAdd) run(cmd *cobra.Command, args []string) error {
	m, err := microcluster.App(microcluster.Args{
		StateDir: c.common.FlagStateDir,
		Verbose:  c.common.FlagLogVerbose,
		Debug:    c.common.FlagLogDebug,
		Version:  version.Version(),
	})

	if err != nil {
		return err
	}

	cli, err := m.LocalClient()
	if err != nil {
		return err
	}

	if len(args) < 1 {
		return cmd.Help()
	}

	// convert expiry flag from duration to time stamp
	expiry := time.Time{}
	if c.flagExpiry != 0 {
		expiry = time.Now().Add(c.flagExpiry)
	}

	clusterName := args[0]
	payload := types.RemoteClusterTokenPost{
		ClusterName: clusterName,
		Expiry:      expiry,
	}

	token, err := client.RemoteClusterJoinTokenPostCmd(cmd.Context(), cli, &payload)
	if err != nil {
		return err
	}

	cmd.Println("The token can not be retrieved at a later stage, please save it now.")
	cmd.Println(token)
	return nil
}

type cmdRemoteClusterJoinTokenShow struct {
	common *CmdControl
}

func (c *cmdRemoteClusterJoinTokenShow) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "List all remote join tokens for LXD clusters.",
		Example: `
			lxd-cluster-mgr remote-cluster-join-token show
			Will list all remote join tokens.
		`,
		RunE: c.run,
	}

	return cmd
}

func (c *cmdRemoteClusterJoinTokenShow) run(cmd *cobra.Command, args []string) error {
	m, err := microcluster.App(microcluster.Args{
		StateDir: c.common.FlagStateDir,
		Verbose:  c.common.FlagLogVerbose,
		Debug:    c.common.FlagLogDebug,
		Version:  version.Version(),
	})

	if err != nil {
		return err
	}

	cli, err := m.LocalClient()
	if err != nil {
		return err
	}

	if len(args) > 0 {
		return cmd.Help()
	}

	tokens, err := client.RemoteClusterJoinTokenGetCmd(cmd.Context(), cli)
	if err != nil {
		return err
	}

	headers := []string{
		"CLUSTER_NAME",
		"EXPIRY",
		"CREATED_AT",
	}

	var rows [][]string
	for _, token := range tokens {
		rows = append(rows, []string{
			token.ClusterName,
			token.Expiry.String(),
			token.CreateAt.String(),
		})
	}

	sort.Sort(lxdCmd.SortColumnsNaturally(rows))
	return lxdCmd.RenderTable(lxdCmd.TableFormatTable, headers, rows, tokens)
}

type cmdRemoteClusterJoinTokenRevoke struct {
	common *CmdControl
}

func (c *cmdRemoteClusterJoinTokenRevoke) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke <cluster_name>",
		Short: "Revoke the remote cluster join token with the given name.",
		Example: `
			lxd-cluster-mgr remote-cluster-join-token revoke <cluster_name>
			Will render the specific remote cluster join token invalid.
		`,
		RunE: c.run,
	}

	return cmd
}

func (c *cmdRemoteClusterJoinTokenRevoke) run(cmd *cobra.Command, args []string) error {
	m, err := microcluster.App(microcluster.Args{
		StateDir: c.common.FlagStateDir,
		Verbose:  c.common.FlagLogVerbose,
		Debug:    c.common.FlagLogDebug,
		Version:  version.Version(),
	})

	if err != nil {
		return err
	}

	cli, err := m.LocalClient()
	if err != nil {
		return err
	}

	if len(args) < 1 {
		return cmd.Help()
	}

	return client.RemoteClusterJoinTokenDeleteCmd(cmd.Context(), cli, args[0])
}
