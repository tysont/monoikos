package main

import (
	"fmt"
	"math/rand"
	"strconv"
	//"reflect"
)

var countContextKey = "count"
var doneContextKey = "done"
var max = 10

func main() {

	environment := new(CountEnvironment)
	policy := environment.CreateRandomPolicy()
	policy.SetShakeRate(0)

	for i := 40; i >= 1; i -= 10 {

		n := 0
		t := 0
		policy.SetShakeRate(i)
		outcomes := []Outcome{}
		for j := 0; j < 100000; j++ {

			r := 0
			experiment := environment.CreateExperiment()
			for _, outcome := range experiment.Run(policy) {

				outcomes = append(outcomes, outcome)
				r = outcome.GetReward()
			}

			n++
			t += r
		}

		for i := 0; i < 10; i++ {

			state := NewBasicState()
			state.GetContext()[countContextKey] = strconv.Itoa(i)
			state.GetContext()[doneContextKey] = strconv.FormatBool(false)
			action := policy.GetPreferredAction(state)
			fmt.Printf("'%v'->'%v'\n", state.GetId(), action.GetId())
		}

		fmt.Printf("========================\n")
		policy = environment.CreatePolicy(outcomes)
	}

}

type CountEnvironment struct{}

func (this *CountEnvironment) CreateRandomPolicy() Policy {

	policy := NewBasicPolicy()
	policy.Environment = this
	return policy
}

func (this *CountEnvironment) CreatePolicy(outcomes []Outcome) Policy {

	policy := NewBasicPolicy()
	policy.Environment = this

	occurences := make(map[string]int)
	rewards := make(map[string]int)
	for _, outcome := range outcomes {

		id := outcome.GetId()
		if _, ok := occurences[id]; !ok {

			occurences[id] = 0
			rewards[id] = 0
		}

		occurences[id] = occurences[id] + 1
		rewards[id] = rewards[id] + outcome.GetReward()
	}

	for _, state := range this.GetKnownStates() {

		set := false
		max := 0.0
		var preferredAction Action
		var otherActions []Action
		for _, action := range this.GetLegalActions(state) {

			outcome := BasicOutcome{InitialState: state, ActionTaken: action}
			id := outcome.GetId()
			if _, ok := occurences[id]; ok {

				reward := float64(rewards[id]) / float64(occurences[id])
				//fmt.Printf("%v: %v\n", id, strconv.FormatFloat(reward, 'f', 3, 64))
				if !set {

					set = true
					max = reward
					preferredAction = action

				} else if reward > max {

					max = reward
					otherActions = append(otherActions, preferredAction)
					preferredAction = action

				} else {

					otherActions = append(otherActions, action)
				}
			}
		}

		if set {

			policy.AddState(state, preferredAction, otherActions)
			//fmt.Printf("*%v <- %v\n", state.GetId(), preferredAction.GetId())

		} else {

			policy.AddRandomState(state)
			//fmt.Printf("*%v <- Random\n", state.GetId())

		}
	}

	return policy
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
