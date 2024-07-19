package types

import (
	"testing"

	"github.com/canonical/lxd-site-manager/test/helpers"
)

// Test represents a test function.
type Test = func(e *helpers.Environment) (string, func(t *testing.T))
