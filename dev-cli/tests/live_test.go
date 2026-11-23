package tests

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/canonical/microcloud-cluster-manager/dev-cli/internal/command"
)

// TestLiveSmoke explicitly opts in to read-only queries against a development server.
func TestLiveSmoke(t *testing.T) {
	if os.Getenv("MCM_DEV_LIVE") != "1" {
		t.Skip("set MCM_DEV_LIVE=1 and provide an imported development session")
	}
	server := os.Getenv("MCM_DEV_SERVER")
	file := os.Getenv("MCM_DEV_SESSION_FILE")
	if server == "" || file == "" {
		t.Fatal("MCM_DEV_SERVER and MCM_DEV_SESSION_FILE are required")
	}
	base := []string{"--server", server, "--session-file", file}
	run := func(args ...string) (string, error) {
		var out bytes.Buffer
		c := command.New(command.Options{Out: &out, Err: io.Discard, In: strings.NewReader("")})
		c.SetArgs(append(append([]string{}, base...), args...))
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		err := c.ExecuteContext(ctx)
		return out.String(), err
	}
	for _, args := range [][]string{{"auth", "status"}, {"query", "/1.0/status"}} {
		if _, err := run(args...); err != nil {
			t.Fatalf("%s: %v", strings.Join(args, " "), err)
		}
	}
}
