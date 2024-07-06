package main

import (
	"fmt"
	"strings"

	"github.com/canonical/microcluster/microcluster"
	"github.com/spf13/cobra"

	"github.com/canonical/lxd-site-manager/internal/api/types"
	"github.com/canonical/lxd-site-manager/internal/client"
)

type cmdConfig struct {
	common *CmdControl
}

func (c *cmdConfig) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "manage LXD site manager configurations.",
		RunE:  c.run,
	}

	var cmdSet = cmdConfigSet{common: c.common}
	cmd.AddCommand(cmdSet.command())

	return cmd
}

func (c *cmdConfig) run(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

type cmdConfigSet struct {
	common *CmdControl
}

func (c *cmdConfigSet) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [<member>] <key>=<value>...",
		Short: "Set member or cluster configuration keys",
		Example: `
			lxd-site-mgr config set [<member>] https_address=192.0.2.1:9001
			Will set the https_address configuration key for the local member.

			lxd-site-mgr config set global.address=203.0.113.1
			Will set the global.address configuration key for the cluster.
		`,
		RunE: c.run,
	}

	return cmd
}

func (c *cmdConfigSet) run(cmd *cobra.Command, args []string) error {
	m, err := microcluster.App(microcluster.Args{StateDir: c.common.FlagStateDir, Verbose: c.common.FlagLogVerbose, Debug: c.common.FlagLogDebug})
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

	if !hasKeyValue(args) {
		return fmt.Errorf("No key=value pairs found in arguments")
	}

	if isMemberConfig(args) {
		member := args[0]
		configs, err := getMemberConfigs(args[1:])
		if err != nil {
			return err
		}

		return client.MemberConfigPatchCmd(cmd.Context(), cli, member, &configs)
	}

	configs, err := getManagerConfigs(args)
	if err != nil {
		return err
	}

	return client.ManagerConfigsPatchCmd(cmd.Context(), cli, &configs)
}

func hasKeyValue(args []string) bool {
	for _, arg := range args {
		if strings.Contains(arg, "=") {
			return true
		}
	}

	return false
}

func parseConfig(arg string) (key string, val string) {
	return strings.Split(arg, "=")[0], strings.Split(arg, "=")[1]
}

func isMemberConfig(args []string) bool {
	return !strings.Contains(args[0], "=")
}

func isConfig(arg string) bool {
	return strings.Contains(arg, "=") && len(strings.Split(arg, "=")) == 2
}

func getMemberConfigs(args []string) (types.MemberConfigPatch, error) {
	validMemberConfigKeys := types.ValidMemberConfigKeys()
	configs := types.MemberConfigPatch{}

	for _, arg := range args {
		if !isConfig(arg) {
			return types.MemberConfigPatch{}, fmt.Errorf("Invalid argument: %s", arg)
		}

		key, val := parseConfig(arg)
		if _, ok := validMemberConfigKeys[key]; !ok {
			return types.MemberConfigPatch{}, fmt.Errorf("Invalid member config key: %s", key)
		}

		switch key {
		case "https_address":
			configs.HTTPSAddress = val
		case "external_address":
			configs.ExternalAddress = val
		}
	}

	return configs, nil
}

func getManagerConfigs(args []string) (types.ManagerConfigs, error) {
	validManagerConfigs := types.ValidManagerConfigKeys()
	configs := types.ManagerConfigs{
		Config: make(map[string]string),
	}

	for _, arg := range args {
		if !isConfig(arg) {
			return types.ManagerConfigs{}, fmt.Errorf("Invalid argument: %s", arg)
		}

		key, val := parseConfig(arg)
		if _, ok := validManagerConfigs[key]; !ok {
			return types.ManagerConfigs{}, fmt.Errorf("Invalid manager config key: %s", key)
		}

		configs.Config[key] = val
	}

	return configs, nil
}
