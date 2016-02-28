package main

import (
	"fmt"
	"math/rand"
	"strconv"
)

var countContextKey = "count"
var doneContextKey = "done"
var max = 20

func PlayCount() {

	environment := new(CountEnvironment)
	policy := environment.CreateOptimizedPolicy(40, 100000, 5)

	for i := 0; i < max; i++ {

		state := NewBasicState()
		state.GetContext()[countContextKey] = strconv.Itoa(i)
		state.GetContext()[doneContextKey] = strconv.FormatBool(false)
		action := policy.GetPreferredAction(state)
		fmt.Printf("'%v'->'%v'\n", state.GetId(), action.GetId())
	}
}

type CountEnvironment struct{}

func (this *CountEnvironment) CreateRandomPolicy() Policy {

	return CreateRandomPolicy(this)
}

func (this *CountEnvironment) CreateImprovedPolicy(outcomes []Outcome) Policy {

	return CreateImprovedPolicy(this, outcomes)
}

func (this *CountEnvironment) CreateOptimizedPolicy(initialShakeRate int, experimentsPerIteration int, iterations int) Policy {

	return CreateOptimizedPolicy(this, initialShakeRate, experimentsPerIteration, iterations)
}

func (this *CountEnvironment) CreateExperiment() Experiment {

	experiment := NewCountExperiment()
	return experiment
}

func (this *CountEnvironment) GetLegalActions(state State) []Action {

	actions := make([]Action, 2)
	actions[0] = new(IncrementAction)
	actions[1] = new(StopAction)
	return actions
}

func (this *CountEnvironment) GetKnownStates() []State {

	states := make([]State, 0)

	for i := 0; i <= max; i++ {

		s1 := NewBasicState()
		s1.GetContext()[countContextKey] = strconv.Itoa(i)
		s1.GetContext()[doneContextKey] = strconv.FormatBool(false)
		s1.Terminal = false
		SetReward(s1)
		states = append(states, s1)

		s2 := NewBasicState()
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

func (this *CountExperiment) ObserveState() State {

	state := NewBasicState()

	count := this.Context[countContextKey].(int)
	done := this.Context[doneContextKey].(bool)

	state.Context[countContextKey] = strconv.Itoa(count)
	state.Context[doneContextKey] = strconv.FormatBool(done)
	state.Terminal = done

	SetReward(state)
	return state
}

func SetReward(state *BasicState) {

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

func (this *CountExperiment) Run(policy Policy) []Outcome {

	basicOutcomes := make([]*BasicOutcome, 0)
	state := this.ObserveState()
	for !state.IsTerminal() {

		action := policy.GetAction(state)
		action.Run(this.Context)

		outcome := new(BasicOutcome)
		outcome.InitialState = state
		outcome.ActionTaken = action
		basicOutcomes = append(basicOutcomes, outcome)

		state = this.ObserveState()
	}

	outcomes := make([]Outcome, 0)
	for _, outcome := range basicOutcomes {

		outcome.FinalState = state
		outcomes = append(outcomes, outcome)
	}

	return outcomes
}

func (this *CountExperiment) ForceRun(action Action, policy Policy) []Outcome {

	basicOutcomes := make([]*BasicOutcome, 0)
	state := this.ObserveState()

	action.Run(this.Context)
	outcome := new(BasicOutcome)
	outcome.InitialState = state
	outcome.ActionTaken = action
	basicOutcomes = append(basicOutcomes, outcome)

	state = this.ObserveState()
	for !state.IsTerminal() {

		action := policy.GetAction(state)
		action.Run(this.Context)

		outcome := new(BasicOutcome)
		outcome.InitialState = state
		outcome.ActionTaken = action
		basicOutcomes = append(basicOutcomes, outcome)

		state = this.ObserveState()
	}

	outcomes := make([]Outcome, 0)
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
