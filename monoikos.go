package monoikos

import (
	"math/rand"
	"sort"
	"strconv"
)

// Environment contains all aspects of a domain in which reinforcement learning may be applied.
// It handles things like creating and iterating on policies, and keeping tabs on the sets of
// known states and actions that can be taken in a given state.
type Environment interface {
	CreateRandomPolicy() Policy
	CreateImprovedPolicy([]Outcome) Policy
	CreateOptimizedPolicy(initialRandomizationRate int, experimentsPerIteration int, iterations int) Policy
	CreateExperiment() Experiment
	GetLegalActions(State) []Action
	GetKnownStates() []State
}

// Experiment is a single instance of a walk thru a domain that generates outcomes.
// It can be run, and it provides the ability to observe the current state or force the
// execution of a particular action.
type Experiment interface {
	ObserveState() State
	Run(Policy) []Outcome
	ForceRun(Action, Policy) []Outcome
}

// Action is a step that can be executed with context at any point in time.
// It includes a way to get an identifier that could be unique to the type or the action instance,
// and it can be executed by passing in context.
type Action interface {
	GetId() string
	Run(map[string]interface{})
}

// State is a snapshot of an experiment at a point in time.
// It includes a unique identifier of the state, a context, and information about whether
// the state is terminal and what reward has been paid out (typically zero unless terminal).
type State interface {
	GetId() string
	IsTerminal() bool
	GetContext() map[string]string
	GetReward() int
}

// BasicState is a super simple and fairly generic implementation of a state with broad applicability.
type BasicState struct {
	Context  map[string]string
	Terminal bool
	Reward   int
}

// NewBasicState should be used to create a BasicState; it handles instantiating members appropriately.
func NewBasicState() *BasicState {

	state := new(BasicState)
	state.Context = make(map[string]string)
	return state
}

// GetId returns an identifier that uniquely identifies the state by concatenating the contextual
// values and whether the state is terminal into a string.
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

// IsTerminal returns whether the state is terminal.
func (this *BasicState) IsTerminal() bool {

	return this.Terminal
}

// GetContext returns the state's context.
func (this *BasicState) GetContext() map[string]string {

	return this.Context
}

// GetReward returns the state's reward, which is typically zero if the state isn't terminal.
func (this *BasicState) GetReward() int {

	return this.Reward
}

// Policy is a decision making process for which actions should be taken in which states.
// Creating a Policy generally involves iterating on a previous policy to generate outcomes, and then
// crafting a new policy from those outcomes in succession.
type Policy interface {
	GetAction(State) Action
	GetPreferredAction(State) Action
	AddRandomState(State)
	AddState(State, Action, []Action)
	SetRandomizationRate(int)
	GetRandomizationRate() int
}

// BasicPolicy is a straightforward and fairly generic implementation of a policy with broad applicability.
type BasicPolicy struct {
	RandomizationRate int
	Environment       Environment
	KnownStates       map[string]State
	PreferredAction   map[string]Action
	OtherActions      map[string][]Action
}

// NewBasicPolicy should be used to create a BasicPolicy; it handles instantiating members appropriately.
func NewBasicPolicy() *BasicPolicy {

	policy := new(BasicPolicy)
	policy.RandomizationRate = 40
	policy.KnownStates = make(map[string]State)
	policy.PreferredAction = make(map[string]Action)
	policy.OtherActions = make(map[string][]Action)

	return policy
}

// GetAction returns an action for a given state that could either be the preferred action, or
// another random action if randomization is triggered based on the randomization rate.
func (this *BasicPolicy) GetAction(state State) Action {

	id := state.GetId()
	if _, ok := this.KnownStates[id]; !ok {

		this.AddRandomState(state)
	}

	k := rand.Intn(100)
	l := len(this.OtherActions[id])
	if l > 0 && k < this.RandomizationRate {

		m := rand.Intn(l)
		return this.OtherActions[id][m]
	}

	return this.PreferredAction[id]
}

// GetPreferredAction returns the preferred action, and never uses any randomization.
func (this *BasicPolicy) GetPreferredAction(state State) Action {

	return this.PreferredAction[state.GetId()]
}

// AddRandomState adds a state to the policy and picks a random action as the state preferred action.
func (this *BasicPolicy) AddRandomState(state State) {

	actions := this.Environment.GetLegalActions(state)

	k := rand.Intn(len(actions))
	action := actions[k]
	actions = append(actions[:k], actions[k+1:]...)

	this.AddState(state, action, actions)
}

// AddState adds a state to the policy and uses the specified action as the state preferred action.
func (this *BasicPolicy) AddState(state State, preferredAction Action, otherActions []Action) {

	id := state.GetId()
	this.KnownStates[id] = state
	this.PreferredAction[id] = preferredAction
	this.OtherActions[id] = otherActions
}

// SetRandomizationRate sets the rate where a random other action will be picked instead of using
// the preferred action.  It must be a number between 0 and 100, and should typically be less
// than 50 for most cases.
func (this *BasicPolicy) SetRandomizationRate(randomizationRate int) {

	this.RandomizationRate = randomizationRate
}

// GetRandomizationRate gets the rate where a random other action will be picked instead of using
// the preferred action.
func (this *BasicPolicy) GetRandomizationRate() int {

	return this.RandomizationRate
}

// Outcome is the result of choosing a set of actions during an experiment (which will generate
// a list of outcomes; one per each state in the experiment).  It has an identifier which can uniquely
// identify the state/reward pair, the initial and final state pair, and the reward.
type Outcome interface {
	GetId() string
	GetReward() int
	GetInitialState() State
	GetFinalState() State
}

// BasicOutcome is a simple implementation of Outcome that is broadly applicable.
type BasicOutcome struct {
	InitialState State
	ActionTaken  Action
	FinalState   State
}

// GetId returns an identifier that uniquely identifies the outcome by concatenating the identifier
// of the state and the identifier of the action taken into a string.
func (this *BasicOutcome) GetId() string {

	s := "["
	s += this.InitialState.GetId()
	s += " => "
	s += this.ActionTaken.GetId()
	s += "]"

	return s
}

// GetReward returns the reward that was attained as part of the outcome by following the action
// from the initial state.
func (this *BasicOutcome) GetReward() int {

	return this.FinalState.GetReward()
}

// GetInitialState returns the initial state for the particular outcome.  Note that an outcome is
// only a pair of initial and end states, so a unique pair will be created for each state that is
// visited until the terminal state is reached.
func (this *BasicOutcome) GetInitialState() State {

	return this.InitialState
}

// GetFinalState returns the final state for the particular outcome.
func (this *BasicOutcome) GetFinalState() State {

	return this.FinalState
}

// CreateRandomPolicy is a utility function for crafting a random policy.  The current implementation
// is light enough that it may not warrant it's own function, but it's likely that more will be
// added here in the future.
func CreateRandomPolicy(environment Environment) Policy {

	policy := NewBasicPolicy()
	policy.Environment = environment
	return policy
}

// CreateImprovedPolicy is a utility function for creating an improved policy from an existing
// policy and a set of outcomes.
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

// CreateOptimizedPolicy is a utility function for running iterations of generating a random policy,
// testing the policy and keeping track of outcomes, and then iterating again and generating a
// better policy.  The policy that is returned should be fairly optimized, assuming that the environment
// and state space was defined correctly, and the tuning parameters were reasonable.
func CreateOptimizedPolicy(environment Environment, initialRandomizationRate int, experimentsPerIteration int, iterations int) Policy {

	policy := environment.CreateRandomPolicy()

	for i := (iterations - 1); i >= 0; i-- {

		n := 0
		t := 0

		randomizationRate := int(float64(initialRandomizationRate) * (float64(i) / float64(iterations-1)))
		policy.SetRandomizationRate(randomizationRate)

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

	policy.SetRandomizationRate(0)
	return policy
}
