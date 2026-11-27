package helpers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// QueryClusterLXD runs a bounded local LXD query inside a cluster container in the
// microcloud-mcm project.
func QueryClusterLXD(clusterName string, method string, path string, input any, output any) error {
	args := []string{"--project", "microcloud-mcm", "exec", "--force-noninteractive", "--disable-stdin", clusterName, "--", "lxc", "--force-local", "query", "--wait", "--request", method, path}
	if input != nil {
		body, err := json.Marshal(input)
		if err != nil {
			return fmt.Errorf("encode local LXD request: %w", err)
		}

		args = append(args, "--data", string(body))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	body, err := exec.CommandContext(ctx, "lxc", args...).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("local LXD %s %s on cluster %q failed: %w: %s", method, path, clusterName, err, strings.TrimSpace(string(exitErr.Stderr)))
		}

		return fmt.Errorf("local LXD %s %s on cluster %q failed: %w", method, path, clusterName, err)
	}

	if output != nil {
		err = json.Unmarshal(body, output)
		if err != nil {
			return fmt.Errorf("decode local LXD response from cluster %q: %w", clusterName, err)
		}
	}

	return nil
}
