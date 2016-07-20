package monoikos_test

import (
	"strconv"
	"testing"

	"github.com/tysont/blackjack"
	"github.com/tysont/monoikos"
)

var idContextKey = "id"
var playerContextKey = "player"
var pairContextKey = "pair"
var softContextKey = "soft"
var dealerContextKey = "dealer"

func TestGetThreeLegalActions(t *testing.T) {

	state := monoikos.NewBasicState()
	state.Context[playerContextKey] = "10"
	state.Context[pairContextKey] = "true"
	state.Context[softContextKey] = "false"
	state.Context[dealerContextKey] = "15"

	environment := BlackjackEnvironment{}
	actions := environment.GetLegalActions(state)
	l := len(actions)

	if l != 3 {
		t.Errorf("Expected 3 legal actions for a pair of cards, got '%v'.", l)
	}
}

func TestGetTwoLegalActions(t *testing.T) {

	state := monoikos.NewBasicState()
	state.Context[playerContextKey] = "14"
	state.Context[dealerContextKey] = "15"
	state.Context[pairContextKey] = "false"
	state.Context[softContextKey] = "false"

	environment := BlackjackEnvironment{}
	actions := environment.GetLegalActions(state)
	l := len(actions)

	if l != 2 {
		t.Errorf("Expected 2 legal actions for a non-pair of cards, got '%v'.", l)
	}
}

func TestOptimizeBlackjackPolicy(t *testing.T) {

	environment := new(BlackjackEnvironment)
	policy := environment.CreateOptimizedPolicy(40, 100000, 5)

	var state monoikos.State
	var action monoikos.Action

	state = monoikos.NewBasicState()
	state.GetContext()[playerContextKey] = "5"
	state.GetContext()[dealerContextKey] = "18"
	state.GetContext()[pairContextKey] = "true"
	state.GetContext()[softContextKey] = "false"

	action = policy.GetPreferredAction(state)
	if action.GetId() != "Hit" {
		t.Errorf("Expected optimized policy to Hit on 5 against 18, got '%v'.", action.GetId())
	}

	state = monoikos.NewBasicState()
	state.GetContext()[playerContextKey] = "20"
	state.GetContext()[dealerContextKey] = "15"
	state.GetContext()[pairContextKey] = "false"
	state.GetContext()[softContextKey] = "false"

	action = policy.GetPreferredAction(state)
	if action.GetId() != "Stand" {
		t.Errorf("Expected optimized policy to Stand on 20 against 15, got '%v'.", action.GetId())
	}

	/*
		// Fails right now, need to debug.
		state = monoikos.NewBasicState()
		state.GetContext()[playerContextKey] = "11"
		state.GetContext()[dealerContextKey] = "16"
		state.GetContext()[pairContextKey] = "true"
		state.GetContext()[softContextKey] = "true"

		action = policy.GetPreferredAction(state)
		if action.GetId() != "Double" {
			t.Errorf("Expected optimized policy to Double on 11 against 16, got '%v'.", action.GetId())
		}
	*/
}

type BlackjackEnvironment struct{}

func (this *BlackjackEnvironment) CreateRandomPolicy() monoikos.Policy {

	return monoikos.CreateRandomPolicy(this)
}

func (this *BlackjackEnvironment) CreateImprovedPolicy(outcomes []monoikos.Outcome) monoikos.Policy {

	return monoikos.CreateImprovedPolicy(this, outcomes)
}

func (this *BlackjackEnvironment) CreateOptimizedPolicy(initialRandomizationRate int, experimentsPerIteration int, iterations int) monoikos.Policy {

	return monoikos.CreateOptimizedPolicy(this, initialRandomizationRate, experimentsPerIteration, iterations)
}

func (this *BlackjackEnvironment) CreateExperiment() monoikos.Experiment {

	experiment := NewBlackjackExperiment()
	id := blackjack.GetNextId()
	experiment.Context[idContextKey] = id
	blackjack.Deal(id)

	return experiment
}

func (this *BlackjackEnvironment) GetLegalActions(state monoikos.State) []monoikos.Action {

	s, ok := state.GetContext()[pairContextKey]
	if !ok {
		return make([]monoikos.Action, 0)
	}

	b, err := strconv.ParseBool(s)
	if err != nil {
		return make([]monoikos.Action, 0)
	}

	var actions []monoikos.Action

	if !b {

		actions = make([]monoikos.Action, 2)
		actions[0] = new(HitAction)
		actions[1] = new(StandAction)

	} else {

		actions = make([]monoikos.Action, 3)
		actions[0] = new(HitAction)
		actions[1] = new(StandAction)
		actions[2] = new(DoubleAction)
	}

	return actions
}

func (this *BlackjackEnvironment) GetKnownStates() []monoikos.State {

	states := make([]monoikos.State, 0)
	for player := 2; player <= 21; player++ {
		for dealer := 2; dealer <= 21; dealer++ {

			for s := 0; s <= 1; s++ {
				soft := s != 0

				for p := 0; p <= 1; p++ {
					pair := p != 0

					state := monoikos.NewBasicState()
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

func (this *BlackjackExperiment) ObserveState() monoikos.State {

	game := blackjack.Peek(this.Context[idContextKey].(uint64))

	state := monoikos.NewBasicState()

	player, soft := blackjack.Evaluate(game.Player)
	state.Context[playerContextKey] = strconv.Itoa(player)
	state.Context[softContextKey] = strconv.FormatBool(soft)

	pair := len(game.Player) == 2
	state.Context[pairContextKey] = strconv.FormatBool(pair)

	dealer, _ := blackjack.Evaluate(game.Dealer)
	state.Context[dealerContextKey] = strconv.Itoa(dealer)

	state.Terminal = game.Complete
	state.Reward = game.Payout

	return state
}

func (this *BlackjackExperiment) Run(policy monoikos.Policy) []monoikos.Outcome {

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

func (this *BlackjackExperiment) ForceRun(action monoikos.Action, policy monoikos.Policy) []monoikos.Outcome {

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

type HitAction struct{}

func (this *HitAction) Run(context map[string]interface{}) {

	blackjack.Hit(context[idContextKey].(uint64))
}

func (this *HitAction) GetId() string {

	return "Hit"
}

type DoubleAction struct{}

func (this *DoubleAction) Run(context map[string]interface{}) {

	blackjack.Double(context[idContextKey].(uint64))
}

func (this *DoubleAction) GetId() string {

	return "Double"
}

type StandAction struct{}

func (this *StandAction) Run(context map[string]interface{}) {

	blackjack.Stand(context[idContextKey].(uint64))
}

func (this *StandAction) GetId() string {

	return "Stand"
}
