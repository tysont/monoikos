package main

import (
    "fmt"
    "log"
    "net/http"
    "math/rand"
    "strconv"

    "github.com/gorilla/mux"
)

var Cards = map[int]string{
	  1: "A",
	  2: "2",
	  3: "3",
	  4: "4",
	  5: "5",
	  6: "6",
	  7: "7",
	  8: "8",
	  9: "9",
	  10: "10",
	  11: "J",
	  12: "Q",
	  13: "K",
}

type Game struct {

	Player []int `json:"player"`
	Dealer []int `json:"dealer"`
	Deck []int `json:"deck"`

	Double bool `json:"double"`
	Complete bool `json:"complete"`
	Payout int `json:"payout"`
}

func (game *Game) String() string {

    return fmt.Sprintf("%v:%v", readable(game.Player, false), readable(game.Dealer, !game.Complete))
}

var Games map[uint64]*Game

var Played uint64

func Initiatlize() {

	Games = make(map[uint64]*Game)
}

func StartService() {

	router := mux.NewRouter().StrictSlash(true)
    router.HandleFunc("/blackjack", ShowGames)
    router.HandleFunc("/blackjack/{id}", ShowGame)
    log.Fatal(http.ListenAndServe(":8080", router))
}

func GetNextId() uint64 {
	
	Played++
	return Played
}

func getGameIds() []uint64 {

	u := make([]uint64, len(Games))
	i := 0
	for k, _ := range Games {
		u[i] = k
		i++
	}

	return u
}

func ShowGames(w http.ResponseWriter, r *http.Request) {
 
 	fmt.Fprintf(w, "<a href=\"http://localhost:8080/blackjack/%v\">Deal</a><br />", GetNextId())
 	for _, k := range getGameIds() {
 		fmt.Fprintf(w, "<a href=\"http://localhost:8080/blackjack/%v\">Play Hand %v</a><br />", k, k)
 	}
}

func ShowGame(w http.ResponseWriter, r *http.Request) {
    
    vars := mux.Vars(r)
    i, _ := strconv.ParseUint(vars["id"], 10, 64)

    if _, x := Games[i]; !x {    	
    	Deal(i)
	}

	a := r.URL.Query().Get("action")
	if (a == "hit") {
    	Hit(i)
    } else if (a == "double") {
    	Double(i)
    } else if (a == "stand") {
    	Stand(i)
    }

    g := Games[i]
    
    if (!g.Complete) {
	    
	    fmt.Fprintf(w, "<a href=\"http://localhost:8080/blackjack/%v?action=hit\">Hit</a><br />", i)
	    fmt.Fprintf(w, "<a href=\"http://localhost:8080/blackjack/%v?action=double\">Double</a><br />", i)
	    fmt.Fprintf(w, "<a href=\"http://localhost:8080/blackjack/%v?action=stand\">Stand</a><br />", i)
	
	} else {

		fmt.Fprint(w, "<a href=\"http://localhost:8080/blackjack\">Back</a><br />")
		fmt.Fprintf(w, "Payout was %v.<br />", g.Payout)

	}

    fmt.Fprintf(w, "%v", g)
}

func Deal(id uint64) uint64 {

	deck := Shuffle()
	game := new(Game)
	game.Double = false

	player := make([]int, 0)
	dealer := make([]int, 0)
	var k int

	k, deck = Draw(deck)
	player = append(player, k)

	k, deck = Draw(deck)
	dealer = append(dealer, k)

	k, deck = Draw(deck)
	player = append(player, k)

	k, deck = Draw(deck)
	dealer = append(dealer, k)

	game.Deck = deck
	game.Player = player
	game.Dealer = dealer

	game.Complete = false
	game.Payout = 0

	Games[id] = game
	return id
}

func Hit(id uint64) *Game {
	
	game := Peek(id)
	k, deck := Draw(game.Deck)
	if (game.Complete) {
		return game
	}

	game.Player = append(game.Player, k)
	game.Deck = deck
	
	p, _ := Evaluate(game.Player)
	if (p > 21) {
		game = Stand(id)
	}

	return game
}

func Stand(id uint64) *Game {

	game := Peek(id)
	if (game.Complete) {
		return game
	}

	d, s := Evaluate(game.Dealer)
	for (d < 18) || ((d == 17) && (s == true)) {
		
		k, deck := Draw(game.Deck)
		game.Dealer = append(game.Dealer, k)
		game.Deck = deck
		d, s = Evaluate(game.Dealer)
	}

	game.Complete = true
	game.Payout = payout(game)
	return game
}

func Double(id uint64) *Game {

	game := Peek(id)
	if (game.Complete) {
		return game
	}

	game.Double = true
	game = Hit(id)
	game = Stand(id)
	return game
}

func Peek(id uint64) *Game {

	return Games[id]
}

func Shuffle() []int {

	hand := make([]int, 52)
	for k := 1; k <= 13; k++ {
		for i := (k - 1) * 4; i < k *4; i++ {
			hand[i] = k
		}
	}

	return hand
}

func Draw(deck []int) (int, []int) {
	
	k := rand.Intn(len(deck))
	n := deck[k]
	deck = append(deck[:k], deck[k + 1:]...)
	return n, deck
}

func Evaluate(hand []int) (int, bool) {
	
	e := 0
	s := false

	for _, k := range hand {
		if (k >= 2) && (k <= 10) {
			e += k
		}
		if (k >= 11) && (k <= 13) {
			e += 10
		}
	}

	for _, k := range hand {
		if k == 1 {
			if e <= 10 {
				e += 11
				s = true
			} else {
				e += 1
			}
		}
	}

	return e, s
}

func readable(hand []int, hide bool) []string {

	s := make([]string, len(hand))
	for i, k := range hand {
		if (i == 0) || (!hide) {
			s[i] = Cards[k]
		} else {
			s[i] = "X"
		}
	}

	return s
}

func payout (game *Game) int {

	//fmt.Println(game)
	p, _ := Evaluate(game.Player)
	d, _ := Evaluate(game.Dealer)

	if (p == 21) && (len(game.Player) == 2) {
		
		return 15
	}

	if (p > 21) || ((d <= 21) && (d > p)) {
		
		if (game.Double) {
			return -20
		}

		return -10

	} else if (d > 21) || (p > d) {
		
		if (game.Double) {
			
			return 20
		}

		return 10
	}
	
	return 0
}

