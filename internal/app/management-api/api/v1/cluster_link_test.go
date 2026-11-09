package v1

import (
	"testing"
	"time"

	"github.com/canonical/microcloud-cluster-manager/internal/pkg/config"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
)

func TestClusterLinksDeadline(t *testing.T) {
	for _, tc := range []struct {
		name         string
		writeTimeout int
		want         time.Duration
	}{
		{"default", 10, 9 * time.Second},
		{"below cluster timeout", 4, 3 * time.Second},
		{"short timeout", 1, 500 * time.Millisecond},
		{"long timeout", 30, 15 * time.Second},
		{"no write timeout", 0, 15 * time.Second},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rc := types.RouteConfig{Env: &config.Config{WriteTimeout: tc.writeTimeout}}
			if got := clusterLinksDeadline(rc); got != tc.want {
				t.Fatalf("clusterLinksDeadline() = %v, want %v", got, tc.want)
			}
		})
	}
}
