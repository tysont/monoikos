package monoikos_test

import (
	"testing"

	"bitbucket.org/tysont/monoikos"
)

func TestSetRandomizationRate(t *testing.T) {

	n := 72

	policy := monoikos.NewBasicPolicy()
	policy.SetRandomizationRate(n)

	if policy.GetRandomizationRate() != n {

		t.Errorf("Expected randomization rate to be set, and it wasn't.")
	}
}
