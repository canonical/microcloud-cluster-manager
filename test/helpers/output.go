package helpers

import "testing"

// LogTestOutcome logs the test outcome for success and error cases.
func LogTestOutcome(t *testing.T, condition string, err error) {
	// ensure calling function's file and line number is logged
	t.Helper()

	if err == nil {
		t.Logf("\t%s\t%s", Succeed, condition)
	} else {
		t.Fatalf("\t%s\t%s: %v", Failed, condition, err)
	}
}
