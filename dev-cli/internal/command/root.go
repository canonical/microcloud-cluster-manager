// Package command provides endpoint queries and browser session management.
package command

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/canonical/microcloud-cluster-manager/dev-cli/internal/client"
	"github.com/canonical/microcloud-cluster-manager/dev-cli/internal/session"
	"github.com/spf13/cobra"
)

// Options allows commands to run with test streams and an injected TLS transport.
type Options struct {
	In         io.Reader
	Out        io.Writer
	Err        io.Writer
	HTTPClient *http.Client
	// ReadSecret overrides the terminal prompt in tests. Production never echoes cookie values.
	ReadSecret func(string) (string, error)
}

type app struct {
	options     Options
	transport   client.Options
	sessionFile string
	api         *client.Client
}

// New builds a fresh command tree without reading credentials or making network requests.
func New(options Options) *cobra.Command {
	if options.In == nil {
		options.In = os.Stdin
	}
	if options.Out == nil {
		options.Out = os.Stdout
	}
	if options.Err == nil {
		options.Err = os.Stderr
	}
	a := &app{options: options}
	root := &cobra.Command{Use: "mcm-dev", Short: "Development and testing client for management-api (not for production)", Args: cobra.ArbitraryArgs, SilenceUsage: true, SilenceErrors: true}
	root.SetIn(options.In)
	root.SetOut(options.Out)
	root.SetErr(options.Err)
	root.CompletionOptions.DisableDefaultCmd = true
	f := root.PersistentFlags()
	f.StringVar(&a.transport.Server, "server", "https://ma.lxd-cm.local:30000", "Management API HTTPS origin")
	f.StringVar(&a.sessionFile, "session-file", "", "Session file (default: user config directory/mcm-dev/session.json)")
	f.StringVar(&a.transport.CACert, "ca-cert", "", "PEM certificate file to add to system trust (default: MCM_DEV_CA_CERT)")
	f.BoolVar(&a.transport.Insecure, "insecure", false, "Disable TLS verification for development")
	f.DurationVar(&a.transport.Timeout, "timeout", 30*time.Second, "Deadline for each HTTP request")
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return usage(err.Error()) })
	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		if cmd.HasSubCommands() {
			return nil
		}
		var err error
		if a.sessionFile == "" {
			a.sessionFile, err = session.DefaultPath()
			if err != nil {
				return err
			}
		}
		transport := a.transport
		if !f.Changed("ca-cert") {
			transport.CACert = os.Getenv("MCM_DEV_CA_CERT")
		}
		transport.HTTPClient = options.HTTPClient
		a.api, err = client.New(transport, nil)
		if err != nil {
			return err
		}
		a.api.Session, err = session.Load(a.sessionFile, a.api.Origin)
		return err
	}
	root.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return usage("unknown command: " + args[0])
		}
		return cmd.Help()
	}
	a.addAuth(root)
	a.addQuery(root)
	return root
}

func usage(message string) error { return &client.Error{Code: 2, Message: message} }

func exact(n int) cobra.PositionalArgs {
	return func(_ *cobra.Command, args []string) error {
		if len(args) != n {
			return usage(fmt.Sprintf("expected %d positional arguments, got %d", n, len(args)))
		}
		for _, arg := range args {
			if strings.TrimSpace(arg) == "" || arg == "." || arg == ".." {
				return usage("arguments must not be empty, '.' or '..'")
			}
		}
		return nil
	}
}

func group(parent *cobra.Command, name, short string) *cobra.Command {
	c := &cobra.Command{Use: name, Short: short, Args: cobra.ArbitraryArgs}
	c.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return usage("unknown command: " + args[0])
		}
		return cmd.Help()
	}
	parent.AddCommand(c)
	return c
}

func leaf(parent *cobra.Command, use, short string, n int, run func(*cobra.Command, []string) error) *cobra.Command {
	c := &cobra.Command{Use: use, Short: short, Args: exact(n), RunE: run}
	parent.AddCommand(c)
	return c
}
