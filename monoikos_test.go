package monoikos_test

import (
	"testing"

	"bitbucket.org/tysont/monoikos"
)

func TestSetShakeRate(t *testing.T) {

	n := 72

	policy := monoikos.NewBasicPolicy()
	policy.SetShakeRate(n)

	if policy.GetShakeRate() != n {

		t.Errorf("Expected shake rate to be set, and it wasn't.")
	}
}
