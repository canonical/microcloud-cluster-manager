package types

import (
	"testing"

	"github.com/canonical/microcloud-cluster-manager/test/helpers"
	"github.com/getkin/kin-openapi/routers"
)

// Test represents a test function.
type Test = func(e *helpers.Environment) (string, func(t *testing.T))
type UnitTest = func() (string, func(t *testing.T))
type APISchemaTest = func(e *helpers.Environment, router routers.Router) (string, func(t *testing.T))
