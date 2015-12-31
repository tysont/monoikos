package main

import (
	"fmt"
	"strconv"
	//"reflect"
)

var idContextKey = "id"
var playerContextKey = "player"
var pairContextKey = "pair"
var softContextKey = "soft"
var dealerContextKey = "dealer"

func main() {

	Initiatlize()
	
	environment := new(BlackjackEnvironment)
	experiment := environment.CreateExperiment()
	policy := environment.CreateRandomPolicy()
	experiment.Run(policy)
}


type BlackjackEnvironment struct { }

func (this BlackjackEnvironment) CreateRandomPolicy() Policy {

	policy := NewBasicPolicy()
	policy.Environment = this

	return *policy
}

func (this BlackjackEnvironment) CreateExperiment() Experiment {

	experiment := NewBlackjackExperiment()

	id := GetNextId()
	experiment.Context[idContextKey] = id
	Deal(id)

	return experiment
}

func (this BlackjackEnvironment) GetLegalActions(state State) []Action {

	s, ok := state.GetContext()[pairContextKey]
	if !ok {
		return make([]Action, 0)
	}

	b, err := strconv.ParseBool(s)
	if err != nil {
		return make([]Action, 0)
	}
	
	var actions []Action

	if (!b) {
		
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

type BlackjackExperiment struct {

	Context map[string]interface{}
}

func NewBlackjackExperiment() *BlackjackExperiment {

	experiment := BlackjackExperiment {}
	experiment.Context = make(map[string]interface{})

	return &experiment
}

func (this BlackjackExperiment) ObserveState() State {
	
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

func (this BlackjackExperiment) Run(policy Policy) []Outcome {

	basicOutcomes := make([]BasicOutcome, 0)
	state := this.ObserveState()	
	for !state.IsTerminal() {

		action := policy.GetAction(state)
		action.Run(this.Context)

		outcome := BasicOutcome {}
		outcome.InitialState = state
		outcome.ActionTaken = action
		basicOutcomes = append(basicOutcomes, outcome)

		state = this.ObserveState()		
	}

	outcomes := make([]Outcome, 0)
	for _, outcome := range(basicOutcomes) {

		outcome.FinalState = state
		outcomes = append(outcomes, outcome)
		fmt.Printf("%+v", outcome.GetId())
	}

	return outcomes
}

type HitAction struct {}

func (this HitAction) Run(context map[string]interface{}) {

	Hit(context[idContextKey].(uint64))
}

func (this HitAction) GetId() string {

	return "Hit"
}

type DoubleAction struct {}

func (this DoubleAction) Run(context map[string]interface{}) {

	Double(context[idContextKey].(uint64))
}

func (this DoubleAction) GetId() string {

	return "Double"
}

type StandAction struct {}

func (this StandAction) Run(context map[string]interface{}) {

	Stand(context[idContextKey].(uint64))
}

func (this StandAction) GetId() string {

	return "Stand"
}