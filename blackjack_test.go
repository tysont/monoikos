package main

import (
	//"fmt"
	"testing"
)

func TestNext(t *testing.T) {
	
	Played = 0
	i := next()
	j := next()
	if (j != i + 1) {
		t.Errorf("Expected consecutive identifier values and didn't get them.")
	}
}

func TestShuffle(t *testing.T) {

	deck := shuffle()
	l := len(deck)
	if (l != 52) {
		t.Errorf("Expected 52 cards in shuffled deck, got '%v'.", l)
	}
}

func TestDraw(t *testing.T) {

	deck := shuffle()
	k, deck := draw(deck)
	l := len(deck)
	if (l != 51) {
		t.Errorf("Expected 51 cards in shuffled deck after 1 draw, got '%v'", l)
	}
	if (k < 1) || (k > 13) {
		t.Errorf("Expected to draw a valid card from a shuffled deck, got an invalid one.")
	}
}

func TestEvaluateEmptyHand(t *testing.T) {

	hand := []int{}
	v, _ := evaluate(hand)
	if (v != 0) {
		t.Errorf("Expected 0 when evaluating an empty hand, got '%v'.", v)
	}
}

func TestEvaluateTenHand(t *testing.T) {

	hand := []int{6, 4}
	v, _ := evaluate(hand)
	if (v != 10) {
		t.Errorf("Expected 10 when evaluating a 6/4 hand, got '%v'.", v)
	}
}

func TestEvaluateAceHand(t *testing.T) {

	hand := []int{10, 2, 1}
	v, s := evaluate(hand)
	if (v != 13) {
		t.Errorf("Expected 13 when evaluating a 10/2/A hand, got '%v'.", v)
	} else if (s == true) {
		t.Errorf("Expected a hard hand on 10/2/A, got a soft one.")
	}
}

func TestEvaluateBlackjackHand(t *testing.T) {

	hand := []int{10, 1}
	v, s := evaluate(hand)
	if (v != 21) {
		t.Errorf("Expected 21 when evaluating a 10/A hand, got '%v'.", v)
	}  else if (s == false) {
		t.Errorf("Expected a soft hand on 10/A, got a hard one.")
	}
}

func TestPayoutWin(t *testing.T) {

	game := new(Game)
	game.Player = []int{10, 11}
	game.Dealer = []int{10, 8}
	game.Double = false
	game.Complete = true

	p := payout(game)
	if (p != 1) {
		t.Errorf("Expected a payout of 1 for a 1 bet 21 vs 18, got '%v'.", p)
	}
}

func TestPayoutLoss(t *testing.T) {

	game := new(Game)
	game.Player = []int{10, 7}
	game.Dealer = []int{10, 4, 6}
	game.Double = false
	game.Complete = true

	p := payout(game)
	if (p != -1) {
		t.Errorf("Expected a payout of -1 for a 1 bet 17 vs 20, got '%v'.", p)
	}
}

func TestPayoutDouble(t *testing.T) {

	game := new(Game)
	game.Player = []int{9, 2, 10}
	game.Dealer = []int{10, 4, 4}
	game.Double = true
	game.Complete = true

	p := payout(game)
	if (p != 2) {
		t.Errorf("Expected a payout of 2 for a 1 bet and double 21 vs 18, got '%v'.", p)
	}
}

func TestPayoutPlayerBust(t *testing.T) {

	game := new(Game)
	game.Player = []int{10, 6, 8}
	game.Dealer = []int{10, 8}
	game.Double = false
	game.Complete = true

	p := payout(game)
	if (p != -1) {
		t.Errorf("Expected a payout of -1 for a 1 bet player bust, got '%v'.", p)
	}
}

func TestPayoutDealerBust(t *testing.T) {

	game := new(Game)
	game.Player = []int{10, 2}
	game.Dealer = []int{10, 3, 9}
	game.Double = false
	game.Complete = true

	p := payout(game)
	if (p != 1) {
		t.Errorf("Expected a payout of 1 for a 1 bet dealer bust, got '%v'.", p)
	}
}

func TestHit(t *testing.T) {

	Played = 0
	Games = make(map[uint64]*Game, 0)

	for z := 0; z < 100; z++ {
		
		i := Deal(1)
		Hit(i)
		Stand(i)

		g := Games[i]
		l := len(g.Player)
		d, _ := evaluate(g.Dealer)

		if (l != 3) {
			t.Errorf("Expected 3 cards after a hit, got '%v'.", l)
		} else if (!g.Complete) {
			t.Errorf("Expected game to be complete after stand and it wasn't.")
		} else if (d < 17) {
			t.Errorf("Expected the dealer's hand to be 17 or more after stand, it was '%v'.", d)
		}
	}
}

func TestDouble(t *testing.T) {

	Played = 0
	Games = make(map[uint64]*Game, 0)

	for z := 0; z < 100; z++ {

		i := Deal(1)
		Double(i)
		Stand(i)

		g := Games[i]
		l := len(g.Player)
		d, _ := evaluate(g.Dealer)
		
		if (l != 3) {
			t.Errorf("Expected 3 cards after a hit, got '%v'.", l)
		} else if (!g.Complete) {
			t.Errorf("Expected game to be complete after stand and it wasn't.")
		} else if (d < 17) {
			t.Errorf("Expected the dealer's hand to be 17 or more after stand, it was '%v'.", d)
		} else if (!g.Double) {
			t.Errorf("Expected double to be set after a double and it wasn't.")
		}
	}
}

func TestStand(t *testing.T) {

	Played = 0
	Games = make(map[uint64]*Game, 0)

	for z := 0; z < 100; z++ {

		i := Deal(1)
		Stand(i)

		g := Games[i]
		d, _ := evaluate(g.Dealer)

		if (!g.Complete) {
			t.Errorf("Expected game to be complete after stand and it wasn't.")
		} else if (d < 17) {
			t.Errorf("Expected the dealer's hand to be 17 or more after stand, it was '%v'.", d)
		}
	}
}

