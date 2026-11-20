package main

import (
	"path"
	"strings"
	"testing"

	routes "github.com/canonical/microcloud-cluster-manager/internal/app/management-api/api"
	"github.com/canonical/microcloud-cluster-manager/internal/pkg/config"
	"github.com/getkin/kin-openapi/openapi3"
)

const specFile = "../../doc/api/cluster-manager-api.yaml"

// TestManagementAPIRoutesDocumented ensures every registered management API endpoint
// (excluding root-level groups such as UI and OIDC) is documented in the OpenAPI spec.
// Paths, including path parameter names, and HTTP methods must match exactly.
func TestManagementAPIRoutesDocumented(t *testing.T) {
	doc, err := openapi3.NewLoader().LoadFromFile(specFile)
	if err != nil {
		t.Fatalf("Failed to load OpenAPI spec %q: %v", specFile, err)
	}

	apiVersion := config.APIVersion()
	var missing []string

	for _, rg := range routes.APIRoutes {
		if rg.IsRoot {
			continue
		}

		for _, e := range rg.Endpoints {
			// Mirrors path construction in internal/pkg/api/rest.go.
			endpointPath := path.Join("/", apiVersion, rg.Prefix, e.Path)
			method := strings.ToUpper(e.Method)

			pathItem := doc.Paths.Value(endpointPath)
			if pathItem == nil || pathItem.GetOperation(method) == nil {
				missing = append(missing, method+" "+endpointPath)
			}
		}
	}

	if len(missing) > 0 {
		t.Errorf("%d registered management API endpoint(s) not documented in %s:\n  %s", len(missing), specFile, strings.Join(missing, "\n  "))
	}
}
