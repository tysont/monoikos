// ABOUTME: Blackjack game engine for use as a reinforcement learning environment.
// ABOUTME: Implements standard blackjack rules including hit, stand, double, soft 17, and natural blackjack payouts.
package blackjack

import (
	"fmt"
	"math/rand/v2"
)

// Cards maps numeric card values to their display representation.
var Cards = map[int]string{
	1: "A", 2: "2", 3: "3", 4: "4", 5: "5", 6: "6", 7: "7",
	8: "8", 9: "9", 10: "10", 11: "J", 12: "Q", 13: "K",
}

// Game represents a single hand of blackjack.
type Game struct {
	Player   []int
	Dealer   []int
	Deck     []int
	Doubled  bool
	Complete bool
	Payout   int
}

// NewGame creates a new blackjack game, shuffles a deck, and deals two cards each
// to the player and dealer.
func NewGame() *Game {
	deck := newDeck()
	g := &Game{}

	var card int
	card, deck = draw(deck)
	g.Player = append(g.Player, card)
	card, deck = draw(deck)
	g.Dealer = append(g.Dealer, card)
	card, deck = draw(deck)
	g.Player = append(g.Player, card)
	card, deck = draw(deck)
	g.Dealer = append(g.Dealer, card)

	g.Deck = deck
	return g
}

// Hit draws one card for the player. If the player busts, the game
// automatically stands.
func (g *Game) Hit() {
	if g.Complete {
		return
	}

	var card int
	card, g.Deck = draw(g.Deck)
	g.Player = append(g.Player, card)

	score, _ := Evaluate(g.Player)
	if score > 21 {
		g.Stand()
	}
}

// Stand ends the player's turn. The dealer draws until reaching 17 or higher
// (hitting on soft 17), then the payout is calculated.
func (g *Game) Stand() {
	if g.Complete {
		return
	}

	score, soft := Evaluate(g.Dealer)
	for score < 17 || (score == 17 && soft) {
		var card int
		card, g.Deck = draw(g.Deck)
		g.Dealer = append(g.Dealer, card)
		score, soft = Evaluate(g.Dealer)
	}

	g.Complete = true
	g.Payout = calculatePayout(g)
}

// Double doubles the bet, draws exactly one card, then stands.
func (g *Game) Double() {
	if g.Complete {
		return
	}
	g.Doubled = true
	g.Hit()
	g.Stand()
}

// Evaluate returns the score of a hand and whether it is soft (contains an ace
// counted as 11).
func Evaluate(hand []int) (int, bool) {
	score := 0
	soft := false

	for _, card := range hand {
		if card >= 2 && card <= 10 {
			score += card
		} else if card >= 11 && card <= 13 {
			score += 10
		}
	}

	for _, card := range hand {
		if card == 1 {
			if score <= 10 {
				score += 11
				soft = true
			} else {
				score += 1
			}
		}
	}

	return score, soft
}

// String returns a human-readable representation of the game state.
// The dealer's hole card is hidden until the game is complete.
func (g *Game) String() string {
	return fmt.Sprintf("%v:%v", readable(g.Player, false), readable(g.Dealer, !g.Complete))
}

// newDeck creates a standard 52-card deck (4 suits of each rank 1-13).
func newDeck() []int {
	deck := make([]int, 52)
	for rank := 1; rank <= 13; rank++ {
		for suit := 0; suit < 4; suit++ {
			deck[(rank-1)*4+suit] = rank
		}
	}
	return deck
}

// draw removes a random card from the deck and returns it along with the
// remaining deck.
func draw(deck []int) (int, []int) {
	i := rand.IntN(len(deck))
	card := deck[i]
	deck = append(deck[:i], deck[i+1:]...)
	return card, deck
}

// readable converts a hand to display strings. If hide is true, all cards
// after the first are shown as "X".
func readable(hand []int, hide bool) []string {
	s := make([]string, len(hand))
	for i, card := range hand {
		if i == 0 || !hide {
			s[i] = Cards[card]
		} else {
			s[i] = "X"
		}
	}
	return s
}

// calculatePayout determines the payout for a completed game.
// Assumes a base bet of 10. Natural blackjack pays 15 (1.5x).
func calculatePayout(g *Game) int {
	player, _ := Evaluate(g.Player)
	dealer, _ := Evaluate(g.Dealer)

	// Natural blackjack (21 with 2 cards) pays 1.5x regardless of double.
	if player == 21 && len(g.Player) == 2 {
		return 15
	}

	bet := 10
	if g.Doubled {
		bet = 20
	}

	if player > 21 || (dealer <= 21 && dealer > player) {
		return -bet
	}
	if dealer > 21 || player > dealer {
		return bet
	}
	return 0
}
