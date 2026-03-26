// ABOUTME: Unit tests for core monoikos framework types.
// ABOUTME: Validates basic policy behavior like randomization rate configuration.
package monoikos_test

import (
	"testing"

	"github.com/tysont/monoikos"
)

func TestSetRandomizationRate(t *testing.T) {
	n := 72
	policy := monoikos.NewBasicPolicy()
	policy.RandomizationRate = n

	if policy.RandomizationRate != n {
		t.Errorf("Expected randomization rate to be %d, got %d.", n, policy.RandomizationRate)
	}
}
