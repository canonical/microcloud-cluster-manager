package command

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/canonical/microcloud-cluster-manager/dev-cli/internal/client"
	"github.com/canonical/microcloud-cluster-manager/dev-cli/internal/session"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func (a *app) addAuth(root *cobra.Command) {
	g := group(root, "auth", "Import and manage browser session credentials")
	var stdin bool
	c := leaf(g, "login", "Import cookies from a browser login", 0, func(cmd *cobra.Command, _ []string) error {
		fmt.Fprintf(a.options.Err, "Log in at %s/oidc/login in your browser. In DevTools, copy the oidc_session and user_secret cookie values.\n", a.api.Origin)
		values := make([]string, 0, 2)
		if stdin {
			scanner := bufio.NewScanner(a.options.In)
			scanner.Buffer(make([]byte, 4096), 1<<20)
			for scanner.Scan() {
				values = append(values, strings.TrimSuffix(scanner.Text(), "\r"))
				if len(values) > 2 {
					break
				}
			}
			if err := scanner.Err(); err != nil {
				return err
			}
			if len(values) != 2 {
				return usage("--stdin expects exactly two lines: oidc_session, then user_secret")
			}
		} else {
			read := a.options.ReadSecret
			if read == nil {
				read = func(name string) (string, error) {
					in, ok := a.options.In.(*os.File)
					if !ok || !term.IsTerminal(int(in.Fd())) {
						return "", usage("hidden prompts require a terminal; use auth login --stdin for piped input")
					}
					fmt.Fprintf(a.options.Err, "%s: ", name)
					value, err := term.ReadPassword(int(in.Fd()))
					fmt.Fprintln(a.options.Err)
					return string(value), err
				}
			}
			for _, name := range session.Names {
				value, err := read(name)
				if err != nil {
					return err
				}
				values = append(values, value)
			}
		}
		if err := a.api.Session.Set(values); err != nil {
			return usage(err.Error())
		}
		a.api.Persist = false
		r, err := a.api.Do(cmd.Context(), http.MethodGet, client.Path(), nil, nil, true)
		if err != nil {
			return err
		}
		if err := a.api.Session.Save(); err != nil {
			return fmt.Errorf("save verified session: %w", err)
		}
		return a.printQuery(r, false)
	})
	c.Flags().BoolVar(&stdin, "stdin", false, "Read exactly two cookie values, one per line, from stdin")
	leaf(g, "status", "Verify the session and show identity/groups", 0, func(cmd *cobra.Command, _ []string) error {
		r, e := a.api.Do(cmd.Context(), http.MethodGet, client.Path(), nil, nil, true)
		if e != nil {
			return e
		}
		return a.printQuery(r, false)
	})
	leaf(g, "logout", "Revoke the server session and remove local credentials", 0, func(cmd *cobra.Command, _ []string) error {
		err := a.api.Logout(cmd.Context())
		removeErr := a.api.Session.Remove()
		if err != nil || removeErr != nil {
			return errors.Join(err, removeErr)
		}
		return a.printQuery(&client.Response{Metadata: []byte("null")}, false)
	})
}
