package main

import (
	"fmt"
	"strings"

	"github.com/canonical/microcluster/microcluster"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/canonical/lxd-site-manager/internal/api/types"
	"github.com/canonical/lxd-site-manager/internal/client"
	"github.com/canonical/lxd-site-manager/version"
)

type cmdConfig struct {
	common *CmdControl
}

func (c *cmdConfig) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "manage LXD Cluster Manager configurations.",
		RunE:  c.run,
	}

	var cmdSet = cmdConfigSet{common: c.common}
	cmd.AddCommand(cmdSet.command())

	var cmdShow = cmdConfigShow{common: c.common}
	cmd.AddCommand(cmdShow.command())

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
			lxd-cluster-mgr config set [<member>] https_address=192.0.2.1:9001
			Will set the https_address configuration key for the local member.

			lxd-cluster-mgr config set global.address=203.0.113.1
			Will set the global.address configuration key for the cluster.
		`,
		RunE: c.run,
	}

	return cmd
}

func (c *cmdConfigSet) run(cmd *cobra.Command, args []string) error {
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
	configs := types.MemberConfigPatch{
		Config: make(map[types.MemberConfigKey]string),
	}

	for _, arg := range args {
		if !isConfig(arg) {
			return types.MemberConfigPatch{}, fmt.Errorf("Invalid argument: %s", arg)
		}

		key, val := parseConfig(arg)
		memberConfigKey := types.MemberConfigKey(key)

		if _, ok := validMemberConfigKeys[memberConfigKey]; !ok {
			return types.MemberConfigPatch{}, fmt.Errorf("Invalid member config key: %s", key)
		}

		switch memberConfigKey {
		case types.HTTPSAddress:
			configs.Config[types.HTTPSAddress] = val
		case types.ExternalAddress:
			configs.Config[types.ExternalAddress] = val
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

type cmdConfigShow struct {
	common *CmdControl
}

func (c *cmdConfigShow) command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [<member>]",
		Short: "Show member or Cluster Manager configurations.",
		Example: `
			lxd-cluster-mgr config show [<member>]
			Will show member specific configurations.

			lxd-cluster-mgr config show
			Will show all LXd Cluster Manager configurations.
		`,
		RunE: c.run,
	}

	return cmd
}

func (c *cmdConfigShow) run(cmd *cobra.Command, args []string) error {
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

	showMemberConfigs := len(args) > 0
	var configs any

	if showMemberConfigs {
		configs, err = client.MemberConfigGetCmd(cmd.Context(), cli, args[0])
	} else {
		configs, err = client.ManagerConfigsGetCmd(cmd.Context(), cli)
	}

	if err != nil {
		return err
	}

	data, err := yaml.Marshal(configs)
	if err != nil {
		return err
	}

	fmt.Printf("%s", data)

	return nil
}
