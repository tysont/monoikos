# Monoikos

A reinforcement learning framework for Go implementing Monte Carlo policy optimization. Monoikos provides a small set of interfaces for defining environments, and handles the learning loop: running experiments, collecting outcomes, computing discounted rewards, and iterating toward an optimal policy.

## Installation

```
go get github.com/tysont/monoikos
```

## How It Works

Monoikos uses Monte Carlo methods to learn optimal policies through experimentation. The process:

1. Start with a random policy
2. Run many episodes through the environment, collecting state-action-reward outcomes
3. Compute average discounted rewards for each state-action pair
4. Build an improved policy that prefers higher-reward actions
5. Repeat with decreasing exploration until the policy converges

The framework supports configurable discount factors (gamma) for weighting near-term vs. future rewards, epsilon-greedy exploration with linear decay, and first-visit or every-visit Monte Carlo variants.

## Defining an Environment

Implement three interfaces to define a new domain:

```go
// Environment describes the domain.
type Environment interface {
    CreateExperiment() Experiment
    LegalActions(State) []Action
    KnownStates() []State
}

// Experiment is a single episode. The framework calls ObserveState
// to read the current snapshot and Context to get the mutable state
// that actions operate on.
type Experiment interface {
    ObserveState() State
    Context() map[string]any
}

// Action is a step the agent can take.
type Action interface {
    ID() string
    Run(map[string]any)
}
```

`State` and `Policy` are also interfaces, but `BasicState` and `BasicPolicy` cover most use cases out of the box.

## Training a Policy

```go
env := &MyEnvironment{}

policy, stats := monoikos.CreateOptimizedPolicy(env, monoikos.TrainingConfig{
    InitialExplorationRate:  40,       // 40% random actions initially
    ExperimentsPerIteration: 100000,   // episodes per training round
    Iterations:              5,        // number of training rounds
    DiscountFactor:          0.99,     // gamma: weight near-term rewards higher
    FirstVisitOnly:          true,     // first-visit Monte Carlo
})

// Use the trained policy.
action := policy.PreferredAction(someState)

// Inspect training progress.
for _, s := range stats {
    fmt.Printf("Iteration %d: exploration=%d%%, avg reward=%.2f\n",
        s.Iteration, s.ExplorationRate, s.AverageReward)
}
```

## Included Packages

### blackjack

The `blackjack` subpackage (`github.com/tysont/monoikos/blackjack`) is a standalone blackjack game engine implementing standard rules: hit, stand, double down, soft 17 dealer behavior, and natural blackjack payouts. It can be used independently or as an RL environment.

```go
g := blackjack.NewGame()
g.Hit()
g.Stand()
fmt.Println(g.Payout)
```

## Examples

The test suite includes two complete environment implementations:

- **Counting game** (`monoikos_count_test.go`) — The agent learns to count as high as possible without exceeding a maximum. Demonstrates basic environment setup and how the discount factor helps the agent learn boundary behavior.

- **Blackjack** (`monoikos_blackjack_test.go`) — The agent learns when to hit, stand, or double down using the built-in blackjack engine. A more realistic domain showing how Monoikos handles larger state spaces with multiple actions.

## Running Tests

```
make test
```

Or directly:

```
go test -v -count=1 ./...
```

## License

MIT License. See [LICENSE](LICENSE) for details.
