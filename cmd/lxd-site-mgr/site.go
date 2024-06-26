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

	"github.com/canonical/lxd-site-manager/internal/client"
)

type cmdSite struct {
	common *CmdControl
}

func (c *cmdSite) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "site",
		Short: "Manage sites",
		RunE:  c.run,
	}

	siteListCmd := &cmdSiteList{
		common: c.common,
		site:   c,
	}

	cmd.AddCommand(siteListCmd.command())

	siteShowCmd := &cmdSiteShow{
		common: c.common,
		site:   c,
	}

	cmd.AddCommand(siteShowCmd.command())

	return cmd
}

func (c *cmdSite) run(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

type cmdSiteList struct {
	common *CmdControl
	site   *cmdSite
}

func (c *cmdSiteList) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List sites",
		RunE:  c.run,
	}

	return cmd
}

func (c *cmdSiteList) run(cmd *cobra.Command, args []string) error {
	m, err := microcluster.App(microcluster.Args{StateDir: c.common.FlagStateDir, Verbose: c.common.FlagLogVerbose, Debug: c.common.FlagLogDebug})
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

	sites, err := client.GetSites(cmd.Context(), microclusterClient)
	if err != nil {
		return err
	}

	headers := []string{
		"NAME",
		"SITE_CERTIFICATE",
		"CREATED_AT",
		"UPDATED_AT",
	}

	var rows [][]string
	for _, site := range sites {
		rows = append(rows, []string{
			site.Name,
			site.SiteCertificate,
			site.CreatedAt.String(),
			site.LastStatusUpdateAt.String(),
		})
	}

	sort.Sort(cli.SortColumnsNaturally(rows))
	return cli.RenderTable(cli.TableFormatTable, headers, rows, sites)
}

type cmdSiteShow struct {
	common *CmdControl
	site   *cmdSite
}

func (c *cmdSiteShow) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show a site",
		RunE:  c.run,
	}

	return cmd
}

func (c *cmdSiteShow) run(cmd *cobra.Command, args []string) error {
	m, err := microcluster.App(microcluster.Args{StateDir: c.common.FlagStateDir, Verbose: c.common.FlagLogVerbose, Debug: c.common.FlagLogDebug})
	if err != nil {
		return err
	}

	var siteName string
	var microclusterClient *microClient.Client
	if len(args) == 1 {
		siteName = args[0]
		microclusterClient, err = m.LocalClient()
	} else if len(args) == 2 {
		microclusterClient, err = m.RemoteClient(args[0])
		siteName = args[1]
	} else {
		return errors.New("Invalid number of arguments")
	}

	if err != nil {
		return err
	}

	site, err := client.GetSite(cmd.Context(), microclusterClient, siteName)
	if err != nil {
		return err
	}

	return yaml.NewEncoder(os.Stdout).Encode(site)
}
