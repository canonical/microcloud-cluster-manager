package main

import (
	"errors"
	"os"
	"sort"

	cli "github.com/canonical/lxd/shared/cmd"
	microClient "github.com/canonical/microcluster/client"
	"github.com/canonical/microcluster/microcluster"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/canonical/lxd-cluster-manager/internal/api/types"
	"github.com/canonical/lxd-cluster-manager/internal/client"
	"github.com/canonical/lxd-cluster-manager/version"
)

type cmdRemoteCluster struct {
	common *CmdControl
}

func (c *cmdRemoteCluster) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remote-cluster",
		Short: "Manage LXD clusters",
		RunE:  c.run,
	}

	remoteClusterListCmd := &cmdRemoteClusterList{
		common:  c.common,
		cluster: c,
	}

	cmd.AddCommand(remoteClusterListCmd.command())

	remoteClusterShowCmd := &cmdRemoteClusterShow{
		common:  c.common,
		cluster: c,
	}

	cmd.AddCommand(remoteClusterShowCmd.command())

	remoteClusterEdit := &cmdRemoteClusterEdit{
		common: c.common,
	}

	cmd.AddCommand(remoteClusterEdit.command())

	return cmd
}

func (c *cmdRemoteCluster) run(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

type cmdRemoteClusterList struct {
	common  *CmdControl
	cluster *cmdRemoteCluster
}

func (c *cmdRemoteClusterList) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List LXD clusters",
		RunE:  c.run,
	}

	return cmd
}

func (c *cmdRemoteClusterList) run(cmd *cobra.Command, args []string) error {
	m, err := microcluster.App(microcluster.Args{
		StateDir: c.common.FlagStateDir,
		Verbose:  c.common.FlagLogVerbose,
		Debug:    c.common.FlagLogDebug,
		Version:  version.Version(),
	})

	if err != nil {
		return err
	}

	var microclusterClient *microClient.Client
	if len(args) == 1 {
		microclusterClient, err = m.RemoteClient(args[0])
	} else if len(args) == 0 {
		microclusterClient, err = m.LocalClient()
	} else {
		return errors.New("Invalid number of arguments")
	}

	if err != nil {
		return err
	}

	remoteClusters, err := client.GetRemoteClusters(cmd.Context(), microclusterClient)
	if err != nil {
		return err
	}

	headers := []string{
		"NAME",
		"CLUSTER_CERTIFICATE",
		"CREATED_AT",
		"UPDATED_AT",
	}

	var rows [][]string
	for _, cluster := range remoteClusters {
		rows = append(rows, []string{
			cluster.Name,
			cluster.ClusterCertificate,
			cluster.CreatedAt.String(),
			cluster.LastStatusUpdateAt.String(),
		})
	}

	sort.Sort(cli.SortColumnsNaturally(rows))
	return cli.RenderTable(cli.TableFormatTable, headers, rows, remoteClusters)
}

type cmdRemoteClusterShow struct {
	common  *CmdControl
	cluster *cmdRemoteCluster
}

func (c *cmdRemoteClusterShow) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show a LXD cluster",
		RunE:  c.run,
	}

	return cmd
}

func (c *cmdRemoteClusterShow) run(cmd *cobra.Command, args []string) error {
	m, err := microcluster.App(microcluster.Args{
		StateDir: c.common.FlagStateDir,
		Verbose:  c.common.FlagLogVerbose,
		Debug:    c.common.FlagLogDebug,
		Version:  version.Version(),
	})

	if err != nil {
		return err
	}

	var remoteClusterName string
	var microclusterClient *microClient.Client
	if len(args) == 1 {
		remoteClusterName = args[0]
		microclusterClient, err = m.LocalClient()
	} else if len(args) == 2 {
		microclusterClient, err = m.RemoteClient(args[0])
		remoteClusterName = args[1]
	} else {
		return errors.New("Invalid number of arguments")
	}

	if err != nil {
		return err
	}

	cluster, err := client.GetRemoteCluster(cmd.Context(), microclusterClient, remoteClusterName)
	if err != nil {
		return err
	}

	return yaml.NewEncoder(os.Stdout).Encode(cluster)
}

type cmdRemoteClusterEdit struct {
	common *CmdControl

	flagStatus string
}

func (c *cmdRemoteClusterEdit) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit <cluster_name>",
		Short: "Edit the data of a cluster",
		Example: `
			lxd-cluster-mgr cluster edit <cluster_name> --status=ACTIVE
			Will modify the status data attribute for the specified cluster.
		`,
		RunE: c.run,
	}

	cmd.Flags().StringVar(&c.flagStatus, "status", "", "Set the status of the cluster, can be ACTIVE or PENDING_APPROVAL")

	return cmd
}

func (c *cmdRemoteClusterEdit) run(cmd *cobra.Command, args []string) error {
	m, err := microcluster.App(microcluster.Args{
		StateDir: c.common.FlagStateDir,
		Verbose:  c.common.FlagLogVerbose,
		Debug:    c.common.FlagLogDebug,
		Version:  version.Version(),
	})

	if err != nil {
		return err
	}

	if len(args) < 1 {
		return cmd.Help()
	}

	cli, err := m.LocalClient()
	if err != nil {
		return err
	}

	payload := &types.RemoteClusterPatch{}
	dataFlagsCount := 0

	if c.flagStatus != "" {
		payload.Status = types.RemoteClusterStatus(c.flagStatus)
		dataFlagsCount++
	}

	if dataFlagsCount == 0 {
		return errors.New("at least one data flag must be provided")
	}

	return client.RemoteClusterPatchCmd(cmd.Context(), cli, args[0], payload)
}
