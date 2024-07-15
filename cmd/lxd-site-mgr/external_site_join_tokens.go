package main

import (
	"sort"
	"time"

	lxdCmd "github.com/canonical/lxd/shared/cmd"
	"github.com/canonical/microcluster/microcluster"
	"github.com/spf13/cobra"

	"github.com/canonical/lxd-site-manager/internal/api/types"
	"github.com/canonical/lxd-site-manager/internal/client"
	"github.com/canonical/lxd-site-manager/version"
)

type cmdExternalSiteJoinToken struct {
	common *CmdControl
}

func (c *cmdExternalSiteJoinToken) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "external-site-join-token",
		Short: "manage external join tokens for LXD sites.",
		RunE:  c.run,
	}

	var cmdAdd = cmdExternalSiteJoinTokenAdd{common: c.common}
	cmd.AddCommand(cmdAdd.command())

	var cmdShow = cmdExternalSiteJoinTokenShow{common: c.common}
	cmd.AddCommand(cmdShow.command())

	return cmd
}

func (c *cmdExternalSiteJoinToken) run(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

type cmdExternalSiteJoinTokenAdd struct {
	common *CmdControl

	flagExpiry time.Duration
}

func (c *cmdExternalSiteJoinTokenAdd) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <site_name>",
		Short: "Add a new external join token for a LXD site.",
		Example: `
			lxd-site-mgr external-site-join-token add <site_name>
			Will created an external site join token for a LXD site.
		`,
		RunE: c.run,
	}

	cmd.Flags().DurationVar(&c.flagExpiry, "expiry", 0, "Specify the duration (i.e. 5m or 48h) for the token to be valid")

	return cmd
}

func (c *cmdExternalSiteJoinTokenAdd) run(cmd *cobra.Command, args []string) error {
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

	siteName := args[0]
	payload := types.ExternalSiteTokenPost{
		SiteName: siteName,
		Expiry:   expiry,
	}

	token, err := client.ExternalSiteJoinTokenPostCmd(cmd.Context(), cli, &payload)
	if err != nil {
		return err
	}

	cmd.Println("The token can not be retrieved at a later stage, please save it now.")
	cmd.Println(token)
	return nil
}

type cmdExternalSiteJoinTokenShow struct {
	common *CmdControl
}

func (c *cmdExternalSiteJoinTokenShow) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "List all external join tokens for LXD sites.",
		Example: `
			lxd-site-mgr external-site-join-token show
			Will list all external join tokens.
		`,
		RunE: c.run,
	}

	return cmd
}

func (c *cmdExternalSiteJoinTokenShow) run(cmd *cobra.Command, args []string) error {
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

	tokens, err := client.ExternalSiteJoinTokenGetCmd(cmd.Context(), cli)
	if err != nil {
		return err
	}

	headers := []string{
		"SITE_NAME",
		"EXPIRY",
		"CREATED_AT",
	}

	var rows [][]string
	for _, token := range tokens {
		rows = append(rows, []string{
			token.SiteName,
			token.Expiry.String(),
			token.CreateAt.String(),
		})
	}

	sort.Sort(lxdCmd.SortColumnsNaturally(rows))
	return lxdCmd.RenderTable(lxdCmd.TableFormatTable, headers, rows, tokens)
}
