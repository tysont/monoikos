// ABOUTME: Blackjack environment for testing the reinforcement learning framework.
// ABOUTME: The agent learns an optimal blackjack strategy (hit, stand, or double).
package monoikos_test

import (
	"strconv"
	"testing"

	"github.com/tysont/blackjack"
	"github.com/tysont/monoikos"
)

const (
	idContextKey     = "id"
	playerContextKey = "player"
	pairContextKey   = "pair"
	softContextKey   = "soft"
	dealerContextKey = "dealer"
)

func TestGetThreeLegalActions(t *testing.T) {
	state := monoikos.NewBasicState()
	state.Context()[playerContextKey] = "10"
	state.Context()[pairContextKey] = "true"
	state.Context()[softContextKey] = "false"
	state.Context()[dealerContextKey] = "15"

	env := &BlackjackEnvironment{}
	actions := env.LegalActions(state)
	if len(actions) != 3 {
		t.Errorf("Expected 3 legal actions for a pair of cards, got '%v'.", len(actions))
	}
}

func TestGetTwoLegalActions(t *testing.T) {
	state := monoikos.NewBasicState()
	state.Context()[playerContextKey] = "14"
	state.Context()[dealerContextKey] = "15"
	state.Context()[pairContextKey] = "false"
	state.Context()[softContextKey] = "false"

	env := &BlackjackEnvironment{}
	actions := env.LegalActions(state)
	if len(actions) != 2 {
		t.Errorf("Expected 2 legal actions for a non-pair of cards, got '%v'.", len(actions))
	}
}

func TestOptimizeBlackjackPolicy(t *testing.T) {
	env := &BlackjackEnvironment{}
	policy := monoikos.CreateOptimizedPolicy(env, 40, 100000, 5)

	state := monoikos.NewBasicState()
	state.Context()[playerContextKey] = "5"
	state.Context()[dealerContextKey] = "18"
	state.Context()[pairContextKey] = "true"
	state.Context()[softContextKey] = "false"

	action := policy.PreferredAction(state)
	if action.ID() != "Hit" {
		t.Errorf("Expected optimized policy to Hit on 5 against 18, got '%v'.", action.ID())
	}

	state = monoikos.NewBasicState()
	state.Context()[playerContextKey] = "20"
	state.Context()[dealerContextKey] = "15"
	state.Context()[pairContextKey] = "false"
	state.Context()[softContextKey] = "false"

	action = policy.PreferredAction(state)
	if action.ID() != "Stand" {
		t.Errorf("Expected optimized policy to Stand on 20 against 15, got '%v'.", action.ID())
	}

	/*
		// Fails right now, need to debug.
		state = monoikos.NewBasicState()
		state.Context()[playerContextKey] = "11"
		state.Context()[dealerContextKey] = "16"
		state.Context()[pairContextKey] = "true"
		state.Context()[softContextKey] = "true"

		action = policy.PreferredAction(state)
		if action.ID() != "Double" {
			t.Errorf("Expected optimized policy to Double on 11 against 16, got '%v'.", action.ID())
		}
	*/
}

// BlackjackEnvironment is a domain where the agent learns optimal blackjack strategy.
type BlackjackEnvironment struct{}

func (e *BlackjackEnvironment) CreateExperiment() monoikos.Experiment {
	x := &BlackjackExperiment{
		state: make(map[string]any),
	}
	id := blackjack.GetNextId()
	x.state[idContextKey] = id
	blackjack.Deal(id)
	return x
}

func (e *BlackjackEnvironment) LegalActions(state monoikos.State) []monoikos.Action {
	s, ok := state.Context()[pairContextKey]
	if !ok {
		return nil
	}
	isPair, err := strconv.ParseBool(s)
	if err != nil {
		return nil
	}

	if isPair {
		return []monoikos.Action{&HitAction{}, &StandAction{}, &DoubleAction{}}
	}
	return []monoikos.Action{&HitAction{}, &StandAction{}}
}

func (e *BlackjackEnvironment) KnownStates() []monoikos.State {
	var states []monoikos.State
	for player := 2; player <= 21; player++ {
		for dealer := 2; dealer <= 21; dealer++ {
			for _, soft := range []bool{false, true} {
				for _, pair := range []bool{false, true} {
					s := monoikos.NewBasicState()
					s.Context()[playerContextKey] = strconv.Itoa(player)
					s.Context()[softContextKey] = strconv.FormatBool(soft)
					s.Context()[pairContextKey] = strconv.FormatBool(pair)
					s.Context()[dealerContextKey] = strconv.Itoa(dealer)
					s.Terminal = false
					states = append(states, s)
				}
			}
		}
	}
	return states
}

// BlackjackExperiment is a single hand of blackjack.
type BlackjackExperiment struct {
	state map[string]any
}

func (x *BlackjackExperiment) ObserveState() monoikos.State {
	game := blackjack.Peek(x.state[idContextKey].(uint64))

	s := monoikos.NewBasicState()
	player, soft := blackjack.Evaluate(game.Player)
	s.Context()[playerContextKey] = strconv.Itoa(player)
	s.Context()[softContextKey] = strconv.FormatBool(soft)
	s.Context()[pairContextKey] = strconv.FormatBool(len(game.Player) == 2)

	dealer, _ := blackjack.Evaluate(game.Dealer)
	s.Context()[dealerContextKey] = strconv.Itoa(dealer)

	s.Terminal = game.Complete
	s.RewardVal = float64(game.Payout)
	return s
}

func (x *BlackjackExperiment) Context() map[string]any {
	return x.state
}

// HitAction draws another card.
type HitAction struct{}

func (a *HitAction) Run(ctx map[string]any) {
	blackjack.Hit(ctx[idContextKey].(uint64))
}

func (a *HitAction) ID() string { return "Hit" }

// DoubleAction doubles down (draw one card, double the bet).
type DoubleAction struct{}

func (a *DoubleAction) Run(ctx map[string]any) {
	blackjack.Double(ctx[idContextKey].(uint64))
}

func (a *DoubleAction) ID() string { return "Double" }

// StandAction keeps the current hand.
type StandAction struct{}

func (a *StandAction) Run(ctx map[string]any) {
	blackjack.Stand(ctx[idContextKey].(uint64))
}

func (a *StandAction) ID() string { return "Stand" }
