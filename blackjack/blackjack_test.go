// ABOUTME: Tests for the blackjack game engine.
// ABOUTME: Covers deck creation, drawing, hand evaluation, payout logic, and full game flow.
package blackjack

import "testing"

func TestNewDeck(t *testing.T) {
	deck := newDeck()
	if len(deck) != 52 {
		t.Errorf("Expected 52 cards in deck, got %d.", len(deck))
	}
}

func TestDraw(t *testing.T) {
	deck := newDeck()
	card, remaining := draw(deck)
	if len(remaining) != 51 {
		t.Errorf("Expected 51 cards after draw, got %d.", len(remaining))
	}
	if card < 1 || card > 13 {
		t.Errorf("Expected valid card (1-13), got %d.", card)
	}
}

func TestEvaluateEmptyHand(t *testing.T) {
	score, _ := Evaluate([]int{})
	if score != 0 {
		t.Errorf("Expected 0 for empty hand, got %d.", score)
	}
}

func TestEvaluateNumericHand(t *testing.T) {
	score, soft := Evaluate([]int{6, 4})
	if score != 10 {
		t.Errorf("Expected 10 for 6+4, got %d.", score)
	}
	if soft {
		t.Errorf("Expected hard hand for 6+4.")
	}
}

func TestEvaluateHardAce(t *testing.T) {
	score, soft := Evaluate([]int{10, 2, 1})
	if score != 13 {
		t.Errorf("Expected 13 for 10+2+A, got %d.", score)
	}
	if soft {
		t.Errorf("Expected hard hand for 10+2+A.")
	}
}

func TestEvaluateSoftAce(t *testing.T) {
	score, soft := Evaluate([]int{10, 1})
	if score != 21 {
		t.Errorf("Expected 21 for 10+A, got %d.", score)
	}
	if !soft {
		t.Errorf("Expected soft hand for 10+A.")
	}
}

func TestEvaluateFaceCards(t *testing.T) {
	score, _ := Evaluate([]int{11, 12, 1})
	if score != 21 {
		t.Errorf("Expected 21 for J+Q+A, got %d.", score)
	}
}

func TestPayoutWin(t *testing.T) {
	g := &Game{Player: []int{10, 11}, Dealer: []int{10, 8}, Complete: true}
	if p := calculatePayout(g); p != 10 {
		t.Errorf("Expected payout 10 for 20 vs 18, got %d.", p)
	}
}

func TestPayoutLoss(t *testing.T) {
	g := &Game{Player: []int{10, 7}, Dealer: []int{10, 4, 6}, Complete: true}
	if p := calculatePayout(g); p != -10 {
		t.Errorf("Expected payout -10 for 17 vs 20, got %d.", p)
	}
}

func TestPayoutDoubleWin(t *testing.T) {
	g := &Game{Player: []int{9, 2, 10}, Dealer: []int{10, 4, 4}, Doubled: true, Complete: true}
	if p := calculatePayout(g); p != 20 {
		t.Errorf("Expected payout 20 for doubled 21 vs 18, got %d.", p)
	}
}

func TestPayoutBlackjack(t *testing.T) {
	g := &Game{Player: []int{10, 1}, Dealer: []int{10, 4, 4}, Complete: true}
	if p := calculatePayout(g); p != 15 {
		t.Errorf("Expected payout 15 for blackjack, got %d.", p)
	}
}

func TestPayoutPlayerBust(t *testing.T) {
	g := &Game{Player: []int{10, 6, 8}, Dealer: []int{10, 8}, Complete: true}
	if p := calculatePayout(g); p != -10 {
		t.Errorf("Expected payout -10 for player bust, got %d.", p)
	}
}

func TestPayoutDealerBust(t *testing.T) {
	g := &Game{Player: []int{10, 2}, Dealer: []int{10, 3, 9}, Complete: true}
	if p := calculatePayout(g); p != 10 {
		t.Errorf("Expected payout 10 for dealer bust, got %d.", p)
	}
}

func TestPayoutPush(t *testing.T) {
	g := &Game{Player: []int{10, 8}, Dealer: []int{10, 8}, Complete: true}
	if p := calculatePayout(g); p != 0 {
		t.Errorf("Expected payout 0 for push, got %d.", p)
	}
}

func TestHitThenStand(t *testing.T) {
	for range 100 {
		g := NewGame()
		g.Hit()
		g.Stand()

		if len(g.Player) < 3 {
			t.Errorf("Expected at least 3 player cards after hit, got %d.", len(g.Player))
		}
		if !g.Complete {
			t.Errorf("Expected game to be complete after stand.")
		}
		d, _ := Evaluate(g.Dealer)
		if d < 17 {
			t.Errorf("Expected dealer score >= 17, got %d.", d)
		}
	}
}

func TestDoubleDown(t *testing.T) {
	for range 100 {
		g := NewGame()
		g.Double()

		if len(g.Player) != 3 {
			t.Errorf("Expected 3 player cards after double, got %d.", len(g.Player))
		}
		if !g.Complete {
			t.Errorf("Expected game to be complete after double.")
		}
		if !g.Doubled {
			t.Errorf("Expected doubled flag to be set.")
		}
		d, _ := Evaluate(g.Dealer)
		if d < 17 {
			t.Errorf("Expected dealer score >= 17, got %d.", d)
		}
	}
}

func TestStand(t *testing.T) {
	for range 100 {
		g := NewGame()
		g.Stand()

		if !g.Complete {
			t.Errorf("Expected game to be complete after stand.")
		}
		d, _ := Evaluate(g.Dealer)
		if d < 17 {
			t.Errorf("Expected dealer score >= 17, got %d.", d)
		}
	}
}

func TestHitOnCompletedGameIsNoop(t *testing.T) {
	g := NewGame()
	g.Stand()
	cards := len(g.Player)
	g.Hit()
	if len(g.Player) != cards {
		t.Errorf("Expected hit on completed game to be a no-op.")
	}
}

func TestNewGameDealsCorrectly(t *testing.T) {
	g := NewGame()
	if len(g.Player) != 2 {
		t.Errorf("Expected 2 player cards after deal, got %d.", len(g.Player))
	}
	if len(g.Dealer) != 2 {
		t.Errorf("Expected 2 dealer cards after deal, got %d.", len(g.Dealer))
	}
	if len(g.Deck) != 48 {
		t.Errorf("Expected 48 cards in deck after deal, got %d.", len(g.Deck))
	}
	if g.Complete {
		t.Errorf("Expected new game to not be complete.")
	}
}
