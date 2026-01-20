package types

import (
	"testing"

	"github.com/canonical/microcloud-cluster-manager/test/helpers"
)

// Test represents a test function.
type Test = func(e *helpers.Environment) (string, func(t *testing.T))
type UnitTest = func() (string, func(t *testing.T))
