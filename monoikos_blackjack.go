package main

import (
	"fmt"
	//"reflect"
	"strconv"
)

var idContextKey = "id"
var playerContextKey = "player"
var pairContextKey = "pair"
var softContextKey = "soft"
var dealerContextKey = "dealer"

func PlayBlackjack() {

	Initiatlize()

	environment := new(BlackjackEnvironment)
	policy := environment.CreateOptimizedPolicy(40, 100000, 5)

	for _, state := range environment.GetKnownStates() {
		if state.GetContext()[playerContextKey] != "21" && state.GetContext()[dealerContextKey] != "21" {
			fmt.Printf("'%v'='%v'\n", state.GetId(), policy.GetPreferredAction(state).GetId())
		}
	}
}

type BlackjackEnvironment struct{}

func (this *BlackjackEnvironment) CreateRandomPolicy() Policy {

	return CreateRandomPolicy(this)
}

func (this *BlackjackEnvironment) CreateImprovedPolicy(outcomes []Outcome) Policy {

	return CreateImprovedPolicy(this, outcomes)
}

func (this *BlackjackEnvironment) CreateOptimizedPolicy(initialShakeRate int, experimentsPerIteration int, iterations int) Policy {

	return CreateOptimizedPolicy(this, initialShakeRate, experimentsPerIteration, iterations)
}

func (this *BlackjackEnvironment) CreateExperiment() Experiment {

	experiment := NewBlackjackExperiment()

	// In a less deterministic world, it may make sense for the experiment to keep a reference
	// to the environment and register new states as they are observed.

	id := GetNextId()
	experiment.Context[idContextKey] = id
	Deal(id)

	return experiment
}

func (this *BlackjackEnvironment) GetLegalActions(state State) []Action {

	s, ok := state.GetContext()[pairContextKey]
	if !ok {
		return make([]Action, 0)
	}

	b, err := strconv.ParseBool(s)
	if err != nil {
		return make([]Action, 0)
	}

	var actions []Action

	if !b {

		actions = make([]Action, 2)
		actions[0] = new(HitAction)
		actions[1] = new(StandAction)

	} else {

		actions = make([]Action, 3)
		actions[0] = new(HitAction)
		actions[1] = new(StandAction)
		actions[2] = new(DoubleAction)
	}

	return actions
}

func (this *BlackjackEnvironment) GetKnownStates() []State {

	states := make([]State, 0)
	for player := 2; player <= 21; player++ {
		for dealer := 2; dealer <= 21; dealer++ {

			for s := 0; s <= 1; s++ {
				soft := s != 0

				for p := 0; p <= 1; p++ {
					pair := p != 0

					state := NewBasicState()
					state.Context[playerContextKey] = strconv.Itoa(player)
					state.Context[softContextKey] = strconv.FormatBool(soft)
					state.Context[pairContextKey] = strconv.FormatBool(pair)
					state.Context[dealerContextKey] = strconv.Itoa(dealer)
					state.Terminal = false

					states = append(states, state)
				}
			}
		}
	}

	return states
}

type BlackjackExperiment struct {
	Context map[string]interface{}
}

func NewBlackjackExperiment() *BlackjackExperiment {

	experiment := new(BlackjackExperiment)
	experiment.Context = make(map[string]interface{})

	return experiment
}

func (this *BlackjackExperiment) ObserveState() State {

	game := Peek(this.Context[idContextKey].(uint64))

	state := NewBasicState()

	player, soft := Evaluate(game.Player)
	state.Context[playerContextKey] = strconv.Itoa(player)
	state.Context[softContextKey] = strconv.FormatBool(soft)

	pair := len(game.Player) == 2
	state.Context[pairContextKey] = strconv.FormatBool(pair)

	dealer, _ := Evaluate(game.Dealer)
	state.Context[dealerContextKey] = strconv.Itoa(dealer)

	state.Terminal = game.Complete
	state.Reward = game.Payout

	return state
}

func (this *BlackjackExperiment) Run(policy Policy) []Outcome {

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

func (this *BlackjackExperiment) ForceRun(action Action, policy Policy) []Outcome {

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

type HitAction struct{}

func (this *HitAction) Run(context map[string]interface{}) {

	Hit(context[idContextKey].(uint64))
}

func (this *HitAction) GetId() string {

	return "Hit"
}

type DoubleAction struct{}

func (this *DoubleAction) Run(context map[string]interface{}) {

	Double(context[idContextKey].(uint64))
}

func (this *DoubleAction) GetId() string {

	return "Double"
}

type StandAction struct{}

func (this *StandAction) Run(context map[string]interface{}) {

	Stand(context[idContextKey].(uint64))
}

func (this *StandAction) GetId() string {

	return "Stand"
}
