package command

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/canonical/microcloud-cluster-manager/dev-cli/internal/client"
	"github.com/spf13/cobra"
)

func (a *app) addQuery(root *cobra.Command) {
	var method, data string
	var raw, wait bool
	var timeout time.Duration
	c := leaf(root, "query ENDPOINT", "Call a management API endpoint directly and print JSON", 1, func(cmd *cobra.Command, args []string) error {
		method = strings.ToUpper(method)
		switch method {
		case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		default:
			return usage("--request must be GET, POST, PUT, PATCH, or DELETE")
		}
		endpoint, err := url.Parse(args[0])
		if err != nil || !strings.HasPrefix(args[0], "/") || strings.HasPrefix(args[0], "//") || endpoint.IsAbs() || endpoint.Host != "" || strings.Contains(args[0], "#") {
			return usage("ENDPOINT must be an API path starting with /, optionally with a query string; URLs and fragments are not accepted")
		}
		query, err := url.ParseQuery(endpoint.RawQuery)
		if err != nil {
			return usage("invalid endpoint query string")
		}
		if timeout <= 0 {
			return usage("--wait-timeout must be positive")
		}
		var cluster string
		if wait {
			parts := strings.Split(endpoint.EscapedPath(), "/")
			if len(parts) < 4 || parts[1] != "1.0" || parts[2] != "remote-cluster" || parts[3] == "" {
				return usage("--wait requires a /1.0/remote-cluster/CLUSTER endpoint to identify the operation's cluster")
			}
			cluster, err = url.PathUnescape(parts[3])
			if err != nil {
				return usage("invalid cluster name in endpoint")
			}
		}
		var body any
		if cmd.Flags().Changed("data") {
			input := []byte(data)
			if data == "-" {
				input, err = io.ReadAll(a.options.In)
				if err != nil {
					return fmt.Errorf("read request data: %w", err)
				}
			}
			if !json.Valid(input) {
				return usage("--data must contain valid JSON (use -d - to read JSON from stdin)")
			}
			body = json.RawMessage(input)
		}
		r, err := a.api.Do(cmd.Context(), method, endpoint.EscapedPath(), query, body, true)
		if err != nil {
			return err
		}
		if wait && r.Type == "async" {
			id, err := client.OperationID(r)
			if err != nil {
				return err
			}
			fmt.Fprintf(a.options.Err, "Waiting for operation %s on %s\n", id, cluster)
			ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()
			r, err = a.api.WaitOperation(ctx, cluster, id)
			if r != nil {
				if outErr := a.printQuery(r, raw); outErr != nil {
					return outErr
				}
			}
			return err
		}
		return a.printQuery(r, raw)
	})
	c.Flags().StringVarP(&method, "request", "X", "GET", "HTTP method: GET, POST, PUT, PATCH, DELETE")
	c.Flags().StringVarP(&data, "data", "d", "", "JSON request body, or - to read JSON from stdin")
	c.Flags().BoolVar(&raw, "raw", false, "Print the full API response envelope instead of metadata")
	c.Flags().BoolVar(&wait, "wait", false, "Wait for the returned operation")
	c.Flags().DurationVar(&timeout, "wait-timeout", 5*time.Minute, "Maximum time to wait for operation completion")
	c.Long = "Call an exact API path on --server using the imported session. Include /1.0 explicitly. Output is JSON; response fields and array order are preserved."
	c.Example = `  mcm-dev query /1.0/remote-cluster
  mcm-dev query '/1.0/remote-cluster/lab-a/replicators?project=demo'
  mcm-dev query /1.0/remote-cluster/lab-a -X PATCH -d '{"description":"Development"}'
  mcm-dev query /1.0/remote-cluster/lab-a/replicators/backup/state?project=demo -X PUT -d '{"action":"start"}' --wait`
}

func (a *app) printQuery(r *client.Response, raw bool) error {
	data := []byte(r.Metadata)
	if raw {
		var err error
		data, err = json.Marshal(r)
		if err != nil {
			return err
		}
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, data, "", "  "); err != nil {
		return fmt.Errorf("format query response: %w", err)
	}
	_, err := fmt.Fprintln(a.options.Out, pretty.String())
	return err
}
