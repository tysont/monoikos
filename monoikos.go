// ABOUTME: Core reinforcement learning framework implementing Monte Carlo policy optimization.
// ABOUTME: Provides interfaces and default implementations for environments, policies, states, actions, and outcomes.
package monoikos

import (
	"fmt"
	"math/rand/v2"
	"sort"
	"strings"
)

// Environment defines a domain in which reinforcement learning can be applied.
// Implementations provide domain-specific experiment creation, legal actions per state,
// and the set of known states for policy construction.
type Environment interface {
	CreateExperiment() Experiment
	LegalActions(State) []Action
	KnownStates() []State
}

// Experiment is a single episode through an environment. ObserveState returns
// the current snapshot; Context returns the mutable state that actions operate on.
// The framework drives the run loop via RunExperiment.
type Experiment interface {
	ObserveState() State
	Context() map[string]any
}

// Action is a step that can be executed within an experiment context.
type Action interface {
	ID() string
	Run(map[string]any)
}

// State is a snapshot of an experiment at a point in time.
type State interface {
	ID() string
	IsTerminal() bool
	Context() map[string]string
	Reward() float64
}

// Policy decides which action to take in a given state.
type Policy interface {
	Action(State) Action
	PreferredAction(State) Action
}

// Outcome is the result of taking an action in a state during an experiment.
type Outcome interface {
	Reward() float64
	InitialState() State
	FinalState() State
	Action() Action
}

// BasicState is a generic implementation of State backed by a string context map.
type BasicState struct {
	ContextMap map[string]string
	Terminal   bool
	RewardVal  float64
}

// NewBasicState creates a BasicState with an initialized context map.
func NewBasicState() *BasicState {
	return &BasicState{
		ContextMap: make(map[string]string),
	}
}

// ID returns a deterministic identifier built from sorted context key-value pairs.
func (s *BasicState) ID() string {
	keys := make([]string, 0, len(s.ContextMap))
	for k := range s.ContextMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	b.WriteString("[")
	for i, k := range keys {
		if i > 0 {
			b.WriteString(" ")
		}
		fmt.Fprintf(&b, "%s:%s", k, s.ContextMap[k])
	}
	fmt.Fprintf(&b, " terminal:%t]", s.Terminal)
	return b.String()
}

func (s *BasicState) IsTerminal() bool           { return s.Terminal }
func (s *BasicState) Context() map[string]string { return s.ContextMap }
func (s *BasicState) Reward() float64            { return s.RewardVal }

// BasicPolicy is a generic policy that tracks a preferred action and alternatives per state,
// with configurable randomization for exploration.
type BasicPolicy struct {
	RandomizationRate int
	Env               Environment
	KnownStates       map[string]State
	PreferredActions  map[string]Action
	OtherActions      map[string][]Action
}

// NewBasicPolicy creates a BasicPolicy with initialized maps and a default randomization rate.
func NewBasicPolicy() *BasicPolicy {
	return &BasicPolicy{
		RandomizationRate: 40,
		KnownStates:       make(map[string]State),
		PreferredActions:  make(map[string]Action),
		OtherActions:      make(map[string][]Action),
	}
}

// Action returns the preferred action for a state, or a random alternative
// based on the randomization rate.
func (p *BasicPolicy) Action(state State) Action {
	id := state.ID()
	if _, ok := p.KnownStates[id]; !ok {
		p.AddRandomState(state)
	}

	k := rand.IntN(100)
	l := len(p.OtherActions[id])
	if l > 0 && k < p.RandomizationRate {
		return p.OtherActions[id][rand.IntN(l)]
	}
	return p.PreferredActions[id]
}

// PreferredAction returns the preferred action without any randomization.
func (p *BasicPolicy) PreferredAction(state State) Action {
	return p.PreferredActions[state.ID()]
}

// AddRandomState adds a state with a randomly selected preferred action.
func (p *BasicPolicy) AddRandomState(state State) {
	actions := p.Env.LegalActions(state)
	k := rand.IntN(len(actions))
	preferred := actions[k]
	others := append(actions[:k], actions[k+1:]...)
	p.AddState(state, preferred, others)
}

// AddState adds a state with a specific preferred action and alternatives.
func (p *BasicPolicy) AddState(state State, preferred Action, others []Action) {
	id := state.ID()
	p.KnownStates[id] = state
	p.PreferredActions[id] = preferred
	p.OtherActions[id] = others
}

// BasicOutcome is a generic implementation of Outcome.
type BasicOutcome struct {
	Initial     State
	ActionTaken Action
	Final       State
}

func (o *BasicOutcome) Reward() float64     { return o.Final.Reward() }
func (o *BasicOutcome) InitialState() State { return o.Initial }
func (o *BasicOutcome) FinalState() State   { return o.Final }
func (o *BasicOutcome) Action() Action      { return o.ActionTaken }

// outcomeKey returns a composite key identifying a state-action pair for reward aggregation.
func outcomeKey(initial State, action Action) string {
	return "[" + initial.ID() + " => " + action.ID() + "]"
}

// RunExperiment drives an experiment to completion using the given policy,
// collecting outcomes for each state transition.
func RunExperiment(experiment Experiment, policy Policy) []Outcome {
	var pending []*BasicOutcome
	state := experiment.ObserveState()
	for !state.IsTerminal() {
		action := policy.Action(state)
		action.Run(experiment.Context())
		outcome := &BasicOutcome{Initial: state, ActionTaken: action}
		pending = append(pending, outcome)
		state = experiment.ObserveState()
	}

	outcomes := make([]Outcome, len(pending))
	for i, o := range pending {
		o.Final = state
		outcomes[i] = o
	}
	return outcomes
}

// ForceRunExperiment runs an experiment, forcing a specific first action
// then following the policy for subsequent steps.
func ForceRunExperiment(experiment Experiment, firstAction Action, policy Policy) []Outcome {
	var pending []*BasicOutcome
	state := experiment.ObserveState()

	firstAction.Run(experiment.Context())
	pending = append(pending, &BasicOutcome{Initial: state, ActionTaken: firstAction})
	state = experiment.ObserveState()

	for !state.IsTerminal() {
		action := policy.Action(state)
		action.Run(experiment.Context())
		pending = append(pending, &BasicOutcome{Initial: state, ActionTaken: action})
		state = experiment.ObserveState()
	}

	outcomes := make([]Outcome, len(pending))
	for i, o := range pending {
		o.Final = state
		outcomes[i] = o
	}
	return outcomes
}

// CreateRandomPolicy creates a new policy with random action selection for an environment.
func CreateRandomPolicy(env Environment) *BasicPolicy {
	p := NewBasicPolicy()
	p.Env = env
	return p
}

// GetAverageRewards computes the average reward for each state-action pair in a set of outcomes.
func GetAverageRewards(outcomes []Outcome) map[string]float64 {
	counts := make(map[string]int)
	totals := make(map[string]float64)
	for _, o := range outcomes {
		key := outcomeKey(o.InitialState(), o.Action())
		counts[key]++
		totals[key] += o.Reward()
	}

	averages := make(map[string]float64, len(counts))
	for key := range counts {
		averages[key] = totals[key] / float64(counts[key])
	}
	return averages
}

// GetOptimalAction finds the action with the highest average reward for a given state.
func GetOptimalAction(env Environment, state State, rewards map[string]float64) (Action, []Action) {
	set := false
	max := 0.0

	var preferred Action
	var others []Action

	for _, action := range env.LegalActions(state) {
		key := outcomeKey(state, action)
		reward, ok := rewards[key]
		if !ok {
			continue
		}
		if !set {
			max = reward
			preferred = action
			set = true
		} else if reward > max {
			max = reward
			others = append(others, preferred)
			preferred = action
		} else {
			others = append(others, action)
		}
	}

	if !set {
		return nil, nil
	}
	return preferred, others
}

// CreateImprovedPolicy builds a new policy by selecting the best action per state
// based on observed outcomes.
func CreateImprovedPolicy(env Environment, outcomes []Outcome) *BasicPolicy {
	rewards := GetAverageRewards(outcomes)
	p := NewBasicPolicy()
	p.Env = env

	for _, state := range env.KnownStates() {
		preferred, others := GetOptimalAction(env, state, rewards)
		if preferred == nil {
			p.AddRandomState(state)
		} else {
			p.AddState(state, preferred, others)
		}
	}
	return p
}

// CreateOptimizedPolicy iterates over rounds of experimentation and improvement,
// decreasing the randomization rate each round to converge on an optimal policy.
func CreateOptimizedPolicy(env Environment, initialRandomizationRate int, experimentsPerIteration int, iterations int) *BasicPolicy {
	policy := CreateRandomPolicy(env)

	for i := iterations - 1; i >= 0; i-- {
		randomizationRate := int(float64(initialRandomizationRate) * (float64(i) / float64(iterations-1)))
		policy.RandomizationRate = randomizationRate

		var outcomes []Outcome
		for j := 0; j < experimentsPerIteration; j++ {
			experiment := env.CreateExperiment()
			outcomes = append(outcomes, RunExperiment(experiment, policy)...)
		}

		policy = CreateImprovedPolicy(env, outcomes)
	}

	policy.RandomizationRate = 0
	return policy
}
