package main

import (
	//"fmt"
	"sort"
	"strconv"
	"math/rand"
	//"reflect"
)

type Environment interface {

	CreateRandomPolicy() Policy
	CreateExperiment() Experiment
	GetLegalActions(State) []Action
}

type Experiment interface {

	ObserveState() State
	Run(Policy) []Outcome
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
	
	Context map[string]string
	Terminal bool
	Reward int
}

func NewBasicState() *BasicState {

	state := BasicState { }
	state.Context = make(map[string]string)
	return &state
}

func (this BasicState) GetId() string {

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
    	
    	if (i > 0) {
    		id += " "
    	}

    	id += k
    	id += ":"
    	id += this.Context[k]

    	i++
    }

    id += "terminal:"
    id += strconv.FormatBool(this.Terminal)
    id += "]"
    

    return id
}

func (this BasicState) IsTerminal() bool {

	return this.Terminal
}

func (this BasicState) GetContext() map[string]string {

	return this.Context
}

func (this BasicState) GetReward() int {

	return this.Reward
}

type Policy interface {

	GetAction(state State) Action
	AddState(state State)
}

type BasicPolicy struct {

	ShakeRate int
	Environment Environment
	KnownStates map[string]State
	PreferredAction map[string]Action
	OtherActions map[string][]Action
}

func NewBasicPolicy() *BasicPolicy {

	policy := BasicPolicy {}
	policy.ShakeRate = 40
	policy.KnownStates = make(map[string]State)
	policy.PreferredAction = make(map[string]Action)
	policy.OtherActions = make(map[string][]Action)

	return &policy
}

func (this BasicPolicy) GetAction(state State) Action {

	id := state.GetId()
	if _, ok := this.KnownStates[id]; !ok {

		this.AddState(state)
	}
	
	k := rand.Intn(100)
	if k < this.ShakeRate {

		l := rand.Intn(len(this.OtherActions[id]))
		return this.OtherActions[id][l]
	}

	return this.PreferredAction[id]
}

func (this BasicPolicy) AddState(state State) {

	actions := this.Environment.GetLegalActions(state)

	k := rand.Intn(len(actions))
	action := actions[k]
	actions = append(actions[:k], actions[k + 1:]...)

	id := state.GetId()
	this.KnownStates[id] = state
	this.PreferredAction[id] = action
	this.OtherActions[id] = actions
}

type Outcome interface {

	GetId() string
	GetReward() int
}

type BasicOutcome struct {

	InitialState State
	ActionTaken Action
	FinalState State
}

func (this BasicOutcome) GetId() string {

	s := "["
	s += this.InitialState.GetId()
	s += " => "
	s += this.ActionTaken.GetId()
	s += "]"

	return s
}

func (this BasicOutcome) GetReward() int {

	return this.FinalState.GetReward()
}