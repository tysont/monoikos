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

	// Get the list of context keys.
	keys := make([]string, len(this.Context))
	i := 0
	for k, _ := range this.Context {
		keys[i] = k
		i++
	}

	// Sort them, important to make the identifier deterministic.
	sort.Strings(keys)

	// Concatenate them together in a somewhat legible format.
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

	// If the state hasn't been seen before, add it with a random action associated to it.
	id := state.GetId()
	if _, ok := this.KnownStates[id]; !ok {

		this.AddRandomState(state)
	}

	// Pick a random number to see whether we should randomize.
	k := rand.Intn(100)
	l := len(this.OtherActions[id])

	// If we know of other actions and should randomize, return a random other action.
	if l > 0 && k < this.RandomizationRate {

		m := rand.Intn(l)
		return this.OtherActions[id][m]
	}

	// Otherwise return the preferred action.
	return this.PreferredAction[id]
}

// GetPreferredAction returns the preferred action, and never uses any randomization.
func (this *BasicPolicy) GetPreferredAction(state State) Action {

	return this.PreferredAction[state.GetId()]
}

// AddRandomState adds a state to the policy and picks a random action as the state preferred action.
func (this *BasicPolicy) AddRandomState(state State) {

	actions := this.Environment.GetLegalActions(state)

	// Select a random action from the list, and remove it from the other actions list.
	k := rand.Intn(len(actions))
	action := actions[k]
	actions = append(actions[:k], actions[k+1:]...)

	// Add the state with the randomly selected preferred action plus other actions.
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

// GetAverageRewards returns the average reward for each state represented in a set of outcomes.
// In the future it may make sense to check for things like the least dense states in the list of
// outcomes since those averages may not mean much.
func GetAverageRewards(outcomes []Outcome) map[string]float64 {

	// Get the raw occurences and total rewards for each state and action pair.
	occurences := make(map[string]int)
	totalRewards := make(map[string]int)
	for _, outcome := range outcomes {

		id := outcome.GetId()
		if _, ok := occurences[id]; !ok {

			occurences[id] = 0
			totalRewards[id] = 0
		}

		occurences[id] = occurences[id] + 1
		totalRewards[id] = totalRewards[id] + outcome.GetReward()
	}

	// Go back thru and calculate the average rewards for each state and action pair.
	averageRewards := make(map[string]float64)
	for id, _ := range occurences {

		averageRewards[id] = float64(totalRewards[id]) / float64(occurences[id])
	}

	return averageRewards
}

// GetOptimalAction returns the optimal preferred action for a state based on a set of rewards for
// outcomes, along with the other possible actions for the state.
func GetOptimalAction(environment Environment, state State, rewards map[string]float64) (Action, []Action) {

	set := false
	max := 0.0

	var preferredAction Action
	var otherActions []Action

	// Iterate over actions to find the one with the highest reward.
	for _, action := range environment.GetLegalActions(state) {

		outcome := BasicOutcome{InitialState: state, ActionTaken: action}
		id := outcome.GetId()
		if _, ok := rewards[id]; ok {

			// If this is the first reward that we've seen, use it.
			reward := rewards[id]
			if !set {

				max = reward
				preferredAction = action
				set = true

				// Otherwise if this reward is better, use it.
			} else if reward > max {

				max = reward
				otherActions = append(otherActions, preferredAction)
				preferredAction = action

				// Or if the old reward was better, stick with it.
			} else {

				otherActions = append(otherActions, action)
			}
		}
	}

	// If we didn't find any rewards for this state, return nil.
	if !set {

		return nil, nil
	}

	// Otherwise return the selected preferred action and other actions.
	return preferredAction, otherActions
}

// CreateImprovedPolicy is a utility function for creating an improved policy from an existing
// policy and a set of outcomes.
func CreateImprovedPolicy(environment Environment, outcomes []Outcome) Policy {

	rewards := GetAverageRewards(outcomes)
	policy := NewBasicPolicy()
	policy.Environment = environment

	// For each state, add it to the policy with a preferred or randomized action.
	for _, state := range environment.GetKnownStates() {

		preferredAction, otherActions := GetOptimalAction(environment, state, rewards)
		if preferredAction == nil {

			policy.AddRandomState(state)

		} else {

			policy.AddState(state, preferredAction, otherActions)
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

	// Loop for the number of desired iterations.
	// The -1 is because we always want an extra iteration at 0 randomization.
	for i := (iterations - 1); i >= 0; i-- {

		n := 0
		t := 0

		// Set a randomization rate that decreases with each iteration.
		randomizationRate := int(float64(initialRandomizationRate) * (float64(i) / float64(iterations-1)))
		policy.SetRandomizationRate(randomizationRate)

		// Run experiments to generate sets of outcomes to improve the policy.
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

		// Create the improved policy and use it moving forward.
		policy = environment.CreateImprovedPolicy(outcomes)
	}

	// Set the final randomization rate to zero and return the policy.
	policy.SetRandomizationRate(0)
	return policy
}
