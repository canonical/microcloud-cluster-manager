package test

import (
	"testing"
)

func TestSample(t *testing.T) {
	result := Sample()
	expected := "sample"

	if result != expected {
		t.Errorf("Impossible, but if it failed, it would print this message.")
	}
}

func Sample() string {
	return "sample"
}
