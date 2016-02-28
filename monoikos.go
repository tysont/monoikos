package main

import (
	"math/rand"
	"sort"
	"strconv"
)

func main() {

}

type Environment interface {
	CreateRandomPolicy() Policy
	CreateImprovedPolicy([]Outcome) Policy
	CreateOptimizedPolicy(initialShakeRate int, experimentsPerIteration int, iterations int) Policy
	CreateExperiment() Experiment
	GetLegalActions(State) []Action
	GetKnownStates() []State
}

type Experiment interface {
	ObserveState() State
	Run(Policy) []Outcome
	ForceRun(Action, Policy) []Outcome
}

type Action interface {
	GetId() string
	Run(map[string]interface{})
}

type State interface {
	GetId() string
	IsTerminal() bool
	GetContext() map[string]string
	GetReward() int
}

type BasicState struct {
	Context  map[string]string
	Terminal bool
	Reward   int
}

func NewBasicState() *BasicState {

	state := BasicState{}
	state.Context = make(map[string]string)
	return &state
}

func (this *BasicState) GetId() string {

	keys := make([]string, len(this.Context))
	i := 0
	for k, _ := range this.Context {
		keys[i] = k
		i++
	}

	sort.Strings(keys)

	id := "["
	i = 0
	for _, k := range keys {

		if i > 0 {
			id += " "
		}

		id += k
		id += ":"
		id += this.Context[k]

		i++
	}

	id += " terminal:"
	id += strconv.FormatBool(this.Terminal)
	id += "]"

	return id
}

func (this *BasicState) IsTerminal() bool {

	return this.Terminal
}

func (this *BasicState) GetContext() map[string]string {

	return this.Context
}

func (this *BasicState) GetReward() int {

	return this.Reward
}

type Policy interface {
	GetAction(State) Action
	GetPreferredAction(State) Action
	AddRandomState(State)
	AddState(State, Action, []Action)
	SetShakeRate(int)
	GetShakeRate() int
}

type BasicPolicy struct {
	ShakeRate       int
	Environment     Environment
	KnownStates     map[string]State
	PreferredAction map[string]Action
	OtherActions    map[string][]Action
}

func NewBasicPolicy() *BasicPolicy {

	policy := BasicPolicy{}
	policy.ShakeRate = 40
	policy.KnownStates = make(map[string]State)
	policy.PreferredAction = make(map[string]Action)
	policy.OtherActions = make(map[string][]Action)

	return &policy
}

func (this *BasicPolicy) GetPreferredAction(state State) Action {

	return this.PreferredAction[state.GetId()]
}

func (this *BasicPolicy) GetAction(state State) Action {

	id := state.GetId()
	if _, ok := this.KnownStates[id]; !ok {

		this.AddRandomState(state)
	}

	k := rand.Intn(100)
	l := len(this.OtherActions[id])
	if l > 0 && k < this.ShakeRate {

		m := rand.Intn(l)
		return this.OtherActions[id][m]
	}

	return this.PreferredAction[id]
}

func (this *BasicPolicy) AddRandomState(state State) {

	actions := this.Environment.GetLegalActions(state)

	k := rand.Intn(len(actions))
	action := actions[k]
	actions = append(actions[:k], actions[k+1:]...)

	this.AddState(state, action, actions)
}

func (this *BasicPolicy) AddState(state State, preferredAction Action, otherActions []Action) {

	id := state.GetId()
	this.KnownStates[id] = state
	this.PreferredAction[id] = preferredAction
	this.OtherActions[id] = otherActions
}

func (this *BasicPolicy) SetShakeRate(shakeRate int) {

	this.ShakeRate = shakeRate
}

func (this *BasicPolicy) GetShakeRate() int {

	return this.ShakeRate
}

type Outcome interface {
	GetId() string
	GetReward() int
	GetInitialState() State
	GetFinalState() State
}

type BasicOutcome struct {
	InitialState State
	ActionTaken  Action
	FinalState   State
}

func (this *BasicOutcome) GetId() string {

	s := "["
	s += this.InitialState.GetId()
	s += " => "
	s += this.ActionTaken.GetId()
	s += "]"

	return s
}

func (this *BasicOutcome) GetReward() int {

	return this.FinalState.GetReward()
}

func (this *BasicOutcome) GetInitialState() State {

	return this.InitialState
}

func (this *BasicOutcome) GetFinalState() State {

	return this.FinalState
}

func CreateRandomPolicy(environment Environment) Policy {

	policy := NewBasicPolicy()
	policy.Environment = environment
	return policy
}

func CreateImprovedPolicy(environment Environment, outcomes []Outcome) Policy {

	policy := NewBasicPolicy()
	policy.Environment = environment

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

	for _, state := range environment.GetKnownStates() {

		set := false
		max := 0.0
		var preferredAction Action
		var otherActions []Action
		for _, action := range environment.GetLegalActions(state) {

			outcome := BasicOutcome{InitialState: state, ActionTaken: action}
			id := outcome.GetId()
			if _, ok := occurences[id]; ok {

				reward := float64(rewards[id]) / float64(occurences[id])
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

		} else {

			policy.AddRandomState(state)

		}
	}

	return policy
}

func CreateOptimizedPolicy(environment Environment, initialShakeRate int, experimentsPerIteration int, iterations int) Policy {

	policy := environment.CreateRandomPolicy()

	for i := (iterations - 1); i >= 0; i-- {

		n := 0
		t := 0

		shakeRate := int(float64(initialShakeRate) * (float64(i) / float64(iterations-1)))
		policy.SetShakeRate(shakeRate)

		outcomes := []Outcome{}
		for j := 0; j < experimentsPerIteration; j++ {

			r := 0
			experiment := environment.CreateExperiment()
			for _, outcome := range experiment.Run(policy) {

				outcomes = append(outcomes, outcome)
				r = outcome.GetReward()
			}

			n++
			t += r
		}

		policy = environment.CreateImprovedPolicy(outcomes)
	}

	policy.SetShakeRate(0)
	return policy
}
