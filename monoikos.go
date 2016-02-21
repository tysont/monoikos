package main

import (
	//"fmt"
	"math/rand"
	"sort"
	"strconv"
	//"reflect"
)

type Environment interface {
	CreateRandomPolicy() Policy
	CreatePolicy([]Outcome) Policy
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
