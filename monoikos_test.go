package main

import (
	"testing"
)

func TestSetShakeRate(t *testing.T) {

	n := 72

	policy := NewBasicPolicy()
	policy.SetShakeRate(n)

	if policy.GetShakeRate() != n {

		t.Errorf("Expected shake rate to be set, and it wasn't.")
	}
}
