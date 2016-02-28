package main

import (
	//"fmt"
	"testing"
)

func TestGetThreeLegalActions(t *testing.T) {

	state := NewBasicState()
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

	state := NewBasicState()
	state.Context[playerContextKey] = "14"
	state.Context[pairContextKey] = "false"
	state.Context[softContextKey] = "false"
	state.Context[dealerContextKey] = "15"

	environment := BlackjackEnvironment{}
	actions := environment.GetLegalActions(state)
	l := len(actions)

	if l != 2 {
		t.Errorf("Expected 2 legal actions for a non-pair of cards, got '%v'.", l)
	}
}
