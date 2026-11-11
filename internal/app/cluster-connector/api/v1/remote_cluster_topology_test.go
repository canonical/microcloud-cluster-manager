package v1

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/database/store"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/types"
)

func TestHasOnwardClusterLinks(t *testing.T) {
	source := store.RemoteClusterWithDetail{
		Name:        "source",
		ClusterUUID: "source-uuid",
		LXDURL:      "https://source.example.test:8443",
	}

	for _, tt := range []struct {
		name    string
		body    string
		want    bool
		wantErr bool
	}{
		{
			name: "empty target",
			body: `{"metadata":[]}`,
		},
		{
			name: "only reverse link by UUID",
			body: `{"metadata":[{"name":"arbitrary-name","config":{"volatile.uuid":"source-uuid"}}]}`,
		},
		{
			name: "only reverse link by address",
			body: `{"metadata":[{"config":{"volatile.addresses":"other.example.test:8443, source.example.test:8443"}}]}`,
		},
		{
			name: "multiple reverse links",
			body: `{"metadata":[{"config":{"volatile.uuid":"source-uuid"}},{"config":{"volatile.addresses":"source.example.test:8443"}}]}`,
		},
		{
			name: "onward link alongside reverse link",
			body: `{"metadata":[{"config":{"volatile.uuid":"source-uuid"}},{"config":{"volatile.uuid":"external-uuid"}}]}`,
			want: true,
		},
		{
			name: "unenrolled onward target",
			body: `{"metadata":[{"name":"external","type":"public","config":{"volatile.uuid":"external-uuid"}}]}`,
			want: true,
		},
		{
			name: "onward link without target identity",
			body: `{"metadata":[{"name":"pending","type":"public"}]}`,
			want: true,
		},
		{
			name: "onward name is not an identity",
			body: `{"metadata":[{"name":"source","config":{"volatile.uuid":"external-uuid"}}]}`,
			want: true,
		},
		{
			name: "empty addresses do not identify a reverse link",
			body: `{"metadata":[{"config":{"volatile.addresses":", "}}]}`,
			want: true,
		},
		{
			name:    "missing metadata",
			body:    `{}`,
			wantErr: true,
		},
		{
			name:    "null metadata",
			body:    `{"metadata":null}`,
			wantErr: true,
		},
		{
			name:    "malformed metadata",
			body:    `{"metadata":{}}`,
			wantErr: true,
		},
		{
			name:    "non-recursive links",
			body:    `{"metadata":["/1.0/cluster/links/external"]}`,
			wantErr: true,
		},
		{
			name:    "invalid response",
			body:    `{`,
			wantErr: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hasOnwardClusterLinks([]byte(tt.body), source)
			if (err != nil) != tt.wantErr {
				t.Fatalf("hasOnwardClusterLinks() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("hasOnwardClusterLinks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveTopologyTargetCluster(t *testing.T) {
	clusters := []store.RemoteClusterWithDetail{
		{Name: "source", ClusterUUID: "source-uuid", LXDURL: "https://source.example.test:8443"},
		{Name: "target", ClusterUUID: "target-uuid", LXDURL: "https://target.example.test:8443/"},
		{Name: "ipv6", LXDURL: "https://[2001:db8::1]:8443"},
	}

	for _, tt := range []struct {
		name   string
		config map[string]string
		want   string
	}{
		{name: "unknown target"},
		{
			name:   "unknown UUID",
			config: map[string]string{"volatile.uuid": "external-uuid"},
		},
		{
			name:   "UUID takes priority",
			config: map[string]string{"volatile.uuid": "target-uuid", "volatile.addresses": "source.example.test:8443"},
			want:   "target",
		},
		{
			name:   "address fallback",
			config: map[string]string{"volatile.uuid": "old-uuid", "volatile.addresses": "other.example.test:8443, target.example.test:8443"},
			want:   "target",
		},
		{
			name:   "IPv6 address",
			config: map[string]string{"volatile.addresses": "[2001:db8::1]:8443"},
			want:   "ipv6",
		},
		{
			name:   "empty addresses",
			config: map[string]string{"volatile.addresses": ", "},
		},
		{
			name:   "partial address must not match",
			config: map[string]string{"volatile.addresses": "example.test:8443"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveTargetClusterName(api.ClusterLink{Name: "arbitrary", Config: tt.config}, clusters)
			if got != tt.want {
				t.Fatalf("resolveTargetClusterName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEnrichedClusterLinkJSON(t *testing.T) {
	for _, state := range []string{"unknown", "empty", "onward"} {
		t.Run(state, func(t *testing.T) {
			link := api.ClusterLink{
				Name:        "external",
				Description: "Original description",
				Type:        "public",
				Config:      map[string]string{"volatile.uuid": "external-uuid"},
				UsedBy:      []string{"/1.0/replicators/example"},
			}
			enriched := enrichedClusterLink{ClusterLink: link}
			if state != "unknown" {
				hasOnward := state == "onward"
				enriched.HasOnwardClusterLinks = &hasOnward
			}
			body, err := json.Marshal(enriched)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(body, &fields); err != nil {
				t.Fatal(err)
			}
			flag, present := fields["has_onward_cluster_links"]
			if present != (state != "unknown") {
				t.Fatalf("unexpected onward field presence: %s", body)
			}
			if present && string(flag) != strconv.FormatBool(state == "onward") {
				t.Fatalf("unexpected onward field: %s", flag)
			}
			var original api.ClusterLink
			if err := json.Unmarshal(body, &original); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(original, link) {
				t.Fatalf("original link fields changed: got %+v, want %+v", original, link)
			}
		})
	}
}

func TestEnrichClusterLinksLookupCancelled(t *testing.T) {
	db, err := database.NewDB(database.DBConfig{DBHost: "127.0.0.1:1", DBDisableTLS: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	links := []api.ClusterLink{{Name: "target", Type: "bidirectional"}}

	for _, state := range []string{"cancelled", "deadline exceeded"} {
		t.Run(state, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			if state == "deadline exceeded" {
				cancel()
				ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
			}
			cancel()
			got := enrichClusterLinksWithDownstreamInfo(ctx, types.RouteConfig{DB: db}, "source", links, httptest.NewRequest("GET", "/", nil))
			if len(got) != 1 || got[0].HasOnwardClusterLinks != nil || !reflect.DeepEqual(got[0].ClusterLink, links[0]) {
				t.Fatalf("failed lookup must preserve the link with unknown onward status: %+v", got)
			}
		})
	}
}

func TestEnrichClusterLinksEmpty(t *testing.T) {
	got := enrichClusterLinksWithDownstreamInfo(context.Background(), types.RouteConfig{}, "source", nil, httptest.NewRequest("GET", "/", nil))
	body, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "[]" {
		t.Fatalf("expected empty array, got %s", body)
	}
}

func TestHasDownstreamLinksRequestFailure(t *testing.T) {
	got, err := hasDownstreamLinks(context.Background(), types.RouteConfig{}, "target", store.RemoteClusterWithDetail{Name: "source"}, httptest.NewRequest("GET", "/", nil))
	if err == nil || got {
		t.Fatalf("failed request must return an error, not confirmed empty topology: got %v, err %v", got, err)
	}
}
