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
	policy := environment.CreateRandomPolicy()
	oldPolicy := policy

	for i := 40; i >= 0; i -= 2 {

		n := 0
		t := 0
		policy.SetShakeRate(i)
		outcomes := make([]Outcome, 0)
		for j := 0; j < 200000; j++ {

			r := 0
			experiment := environment.CreateExperiment()
			for _, outcome := range experiment.Run(policy) {

				outcomes = append(outcomes, outcome)
				r = outcome.GetReward()
			}

			n++
			t += r
		}

		//a := float64(t) / float64(n)
		//fmt.Println(strconv.FormatFloat(a, 'f', 3, 64))
		oldPolicy = policy
		policy = environment.CreatePolicy(outcomes)
	}

	for _, state := range environment.GetKnownStates() {
		if state.GetContext()[playerContextKey] != "21" || state.GetContext()[pairContextKey] != "true" {
			fmt.Printf("'%v'='%v'->'%v'\n", state.GetId(), oldPolicy.GetPreferredAction(state).GetId(), policy.GetPreferredAction(state).GetId())
		}
	}
}

type BlackjackEnvironment struct{}

func (this *BlackjackEnvironment) CreateRandomPolicy() Policy {

	policy := NewBasicPolicy()
	policy.Environment = this

	return policy
}

func (this *BlackjackEnvironment) CreatePolicy(outcomes []Outcome) Policy {

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
