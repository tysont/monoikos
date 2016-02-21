package main

import (
	//"fmt"
	//"math/rand"
	"strconv"
	"testing"
)

/*
func TestZeroShakePolicyDeterminism(t *testing.T) {

	environment := new(CountEnvironment)
	policy := environment.CreateRandomPolicy()
	policy.SetShakeRate(0)

	deterministic := true
	id := ""
	n := rand.Intn(max)
	for i := 0; i < 10; i++ {

		experiment := NewCountExperiment()
		experiment.Context[countContextKey] = n
		state := experiment.ObserveState()
		action := policy.GetAction(state)

		if i > 0 && id != action.GetId() {

			deterministic = false
			break
		}

		id = action.GetId()
	}

	if !deterministic {

		t.Errorf("Expected policy with zero shake to return a deterministic action, but got different actions.")
	}
}

func TestAddStateToPolicy(t *testing.T) {

	environment := new(CountEnvironment)
	policy := environment.CreateRandomPolicy()
	policy.SetShakeRate(0)

	experiment := environment.CreateExperiment()
	state := experiment.ObserveState()
	action := policy.GetAction(state)

	if policy.GetPreferredAction(state).GetId() != action.GetId() {

		t.Errorf("Expected policy to return the preferred action for a new state.")
	}
}

func TestActionResults(t *testing.T) {

	environment := new(CountEnvironment)
	policy := environment.CreateRandomPolicy()

	i := 0
	s := 0

	ia := new(IncrementAction)
	sa := new(StopAction)

	for j := 0; j < 100000; j++ {

		ie := new(CountExperiment)
		ie.Context = make(map[string]interface{})
		ie.Context[countContextKey] = 1
		ie.Context[doneContextKey] = false

		se := new(CountExperiment)
		se.Context = make(map[string]interface{})
		se.Context[countContextKey] = 1
		se.Context[doneContextKey] = false

		io := ie.ForceRun(ia, policy)[0]
		so := se.ForceRun(sa, policy)[0]

		i += io.GetReward()
		s += so.GetReward()
	}

	if i <= s {

		t.Errorf("Expected incrementing a 1 to be better than stopping on 1 over a large number of attempts.")
	}
}

*/

func TestCreatePolicyFromOutcomes(t *testing.T) {

	environment := new(CountEnvironment)
	policy := environment.CreateRandomPolicy()

	ia := new(IncrementAction)
	sa := new(StopAction)

	s1 := NewBasicState()
	s1.GetContext()[countContextKey] = strconv.Itoa(1)
	s1.GetContext()[doneContextKey] = strconv.FormatBool(false)
	s1.Terminal = false
	SetReward(s1)

	s2 := NewBasicState()
	s2.GetContext()[countContextKey] = strconv.Itoa(2)
	s2.GetContext()[doneContextKey] = strconv.FormatBool(true)
	s2.Terminal = true
	SetReward(s2)

	s3 := NewBasicState()
	s3.GetContext()[countContextKey] = strconv.Itoa(1)
	s3.GetContext()[doneContextKey] = strconv.FormatBool(true)
	s3.Terminal = true
	SetReward(s3)

	o1 := new(BasicOutcome)
	o1.InitialState = s1
	o1.ActionTaken = ia
	o1.FinalState = s2

	o2 := new(BasicOutcome)
	o2.InitialState = s1
	o2.ActionTaken = sa
	o2.FinalState = s3

	outcomes := make([]Outcome, 2)
	outcomes[0] = o1
	outcomes[1] = o2

	correct := true
	for j := 0; j < 100; j++ {

		policy = environment.CreatePolicy(outcomes)
		if policy.GetPreferredAction(s1).GetId() != ia.GetId() {

			correct = false
			break
		}
	}

	if !correct {

		t.Errorf("Expected policy to pick correct preferred action based on outcomes, and it didn't.")
	}

}
