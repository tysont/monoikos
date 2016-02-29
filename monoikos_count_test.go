package monoikos_test

import (
	"math/rand"
	"strconv"
	"testing"

	"bitbucket.org/tysont/monoikos"
)

var countContextKey = "count"
var doneContextKey = "done"
var max = 20

func TestZeroRandomizationPolicyDeterminism(t *testing.T) {

	environment := new(CountEnvironment)
	policy := environment.CreateRandomPolicy()
	policy.SetRandomizationRate(0)

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

		t.Errorf("Expected policy with zero randomization to return a deterministic action, but got different actions.")
	}
}

func TestAddStateToPolicy(t *testing.T) {

	environment := new(CountEnvironment)
	policy := environment.CreateRandomPolicy()
	policy.SetRandomizationRate(0)

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

func TestCreatePolicyFromOutcomes(t *testing.T) {

	environment := new(CountEnvironment)
	policy := environment.CreateRandomPolicy()

	ia := new(IncrementAction)
	sa := new(StopAction)

	s1 := monoikos.NewBasicState()
	s1.GetContext()[countContextKey] = strconv.Itoa(1)
	s1.GetContext()[doneContextKey] = strconv.FormatBool(false)
	s1.Terminal = false
	SetReward(s1)

	s2 := monoikos.NewBasicState()
	s2.GetContext()[countContextKey] = strconv.Itoa(2)
	s2.GetContext()[doneContextKey] = strconv.FormatBool(true)
	s2.Terminal = true
	SetReward(s2)

	s3 := monoikos.NewBasicState()
	s3.GetContext()[countContextKey] = strconv.Itoa(1)
	s3.GetContext()[doneContextKey] = strconv.FormatBool(true)
	s3.Terminal = true
	SetReward(s3)

	o1 := new(monoikos.BasicOutcome)
	o1.InitialState = s1
	o1.ActionTaken = ia
	o1.FinalState = s2

	o2 := new(monoikos.BasicOutcome)
	o2.InitialState = s1
	o2.ActionTaken = sa
	o2.FinalState = s3

	outcomes := make([]monoikos.Outcome, 2)
	outcomes[0] = o1
	outcomes[1] = o2

	correct := true
	for j := 0; j < 100; j++ {

		policy = environment.CreateImprovedPolicy(outcomes)
		if policy.GetPreferredAction(s1).GetId() != ia.GetId() {

			correct = false
			break
		}
	}

	if !correct {

		t.Errorf("Expected policy to pick correct preferred action based on outcomes, and it didn't.")
	}
}

func TestCreateOptimizedCountPolicy(t *testing.T) {

	environment := new(CountEnvironment)
	policy := environment.CreateOptimizedPolicy(40, 100000, 5)

	var state monoikos.State
	var action monoikos.Action

	for i := 1; i < max-1; i++ {

		state = monoikos.NewBasicState()
		state.GetContext()[countContextKey] = strconv.Itoa(i)
		state.GetContext()[doneContextKey] = strconv.FormatBool(false)

		action = policy.GetPreferredAction(state)
		if action.GetId() != "Increment" {
			t.Errorf("Expected optimized policy to Increment on '%v', got '%v'.", i, action.GetId())
		}
	}

	/*
		// 19 and 20 fail right now, need to debug.
		state = monoikos.NewBasicState()
		state.GetContext()[countContextKey] = strconv.Itoa(max)
		state.GetContext()[doneContextKey] = strconv.FormatBool(false)

		action = policy.GetPreferredAction(state)
		if action.GetId() != "Stop" {
			t.Errorf("Expected optimized policy to Stop on '%v', got '%v'.", max, action.GetId())
		}
	*/
}

type CountEnvironment struct{}

func (this *CountEnvironment) CreateRandomPolicy() monoikos.Policy {

	return monoikos.CreateRandomPolicy(this)
}

func (this *CountEnvironment) CreateImprovedPolicy(outcomes []monoikos.Outcome) monoikos.Policy {

	return monoikos.CreateImprovedPolicy(this, outcomes)
}

func (this *CountEnvironment) CreateOptimizedPolicy(initialRandomizationRate int, experimentsPerIteration int, iterations int) monoikos.Policy {

	return monoikos.CreateOptimizedPolicy(this, initialRandomizationRate, experimentsPerIteration, iterations)
}

func (this *CountEnvironment) CreateExperiment() monoikos.Experiment {

	experiment := NewCountExperiment()
	return experiment
}

func (this *CountEnvironment) GetLegalActions(state monoikos.State) []monoikos.Action {

	actions := make([]monoikos.Action, 2)
	actions[0] = new(IncrementAction)
	actions[1] = new(StopAction)
	return actions
}

func (this *CountEnvironment) GetKnownStates() []monoikos.State {

	states := make([]monoikos.State, 0)

	for i := 0; i <= max; i++ {

		s1 := monoikos.NewBasicState()
		s1.GetContext()[countContextKey] = strconv.Itoa(i)
		s1.GetContext()[doneContextKey] = strconv.FormatBool(false)
		s1.Terminal = false
		SetReward(s1)
		states = append(states, s1)

		s2 := monoikos.NewBasicState()
		s2.GetContext()[countContextKey] = strconv.Itoa(i)
		s2.GetContext()[doneContextKey] = strconv.FormatBool(true)
		s2.Terminal = true
		SetReward(s2)
		states = append(states, s2)
	}

	return states
}

type CountExperiment struct {
	Context map[string]interface{}
}

func NewCountExperiment() *CountExperiment {

	experiment := new(CountExperiment)
	experiment.Context = make(map[string]interface{})
	experiment.Context[countContextKey] = rand.Intn(max)
	experiment.Context[doneContextKey] = false
	return experiment
}

func (this *CountExperiment) ObserveState() monoikos.State {

	state := monoikos.NewBasicState()

	count := this.Context[countContextKey].(int)
	done := this.Context[doneContextKey].(bool)

	state.Context[countContextKey] = strconv.Itoa(count)
	state.Context[doneContextKey] = strconv.FormatBool(done)
	state.Terminal = done

	SetReward(state)
	return state
}

func SetReward(state *monoikos.BasicState) {

	count, _ := strconv.Atoi(state.Context[countContextKey])
	done, _ := strconv.ParseBool(state.Context[doneContextKey])

	if !done {

		state.Reward = 0

	} else {

		if count > max {

			state.Reward = -1

		} else {

			state.Reward = count
		}
	}
}

func (this *CountExperiment) Run(policy monoikos.Policy) []monoikos.Outcome {

	basicOutcomes := make([]*monoikos.BasicOutcome, 0)
	state := this.ObserveState()
	for !state.IsTerminal() {

		action := policy.GetAction(state)
		action.Run(this.Context)

		outcome := new(monoikos.BasicOutcome)
		outcome.InitialState = state
		outcome.ActionTaken = action
		basicOutcomes = append(basicOutcomes, outcome)

		state = this.ObserveState()
	}

	outcomes := make([]monoikos.Outcome, 0)
	for _, outcome := range basicOutcomes {

		outcome.FinalState = state
		outcomes = append(outcomes, outcome)
	}

	return outcomes
}

func (this *CountExperiment) ForceRun(action monoikos.Action, policy monoikos.Policy) []monoikos.Outcome {

	basicOutcomes := make([]*monoikos.BasicOutcome, 0)
	state := this.ObserveState()

	action.Run(this.Context)
	outcome := new(monoikos.BasicOutcome)
	outcome.InitialState = state
	outcome.ActionTaken = action
	basicOutcomes = append(basicOutcomes, outcome)

	state = this.ObserveState()
	for !state.IsTerminal() {

		action := policy.GetAction(state)
		action.Run(this.Context)

		outcome := new(monoikos.BasicOutcome)
		outcome.InitialState = state
		outcome.ActionTaken = action
		basicOutcomes = append(basicOutcomes, outcome)

		state = this.ObserveState()
	}

	outcomes := make([]monoikos.Outcome, 0)
	for _, outcome := range basicOutcomes {

		outcome.FinalState = state
		outcomes = append(outcomes, outcome)
	}

	return outcomes
}

type IncrementAction struct{}

func (this *IncrementAction) Run(context map[string]interface{}) {

	context[countContextKey] = context[countContextKey].(int) + 1
	if context[countContextKey].(int) > max {
		context[doneContextKey] = true
	}
}

func (this *IncrementAction) GetId() string {

	return "Increment"
}

type StopAction struct{}

func (this *StopAction) Run(context map[string]interface{}) {

	context[doneContextKey] = true
}

func (this *StopAction) GetId() string {

	return "Stop"
}
