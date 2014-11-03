package joker

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// A Ranking is one of the ten possible hand rankings that determine the
// value of a hand.  Hand rankings are composed of different arrangments of
// pairs, straights, and flushes.
type Ranking int

const (
	// HighCard represents a hand composed of no pairs, straights, or flushes.
	// Ex: A♠ K♠ J♣ 7♥ 5♦
	HighCard Ranking = iota

	// Pair represents a hand composed of a single pair.
	// Ex: A♠ A♣ K♣ J♥ 5♦
	Pair

	// TwoPair represents a hand composed of two pairs.
	// Ex: A♠ A♣ J♣ J♦ 5♦
	TwoPair

	// ThreeOfAKind represents a hand composed of three of the same rank.
	// Ex: A♠ A♣ A♦ J♥ 5♦
	ThreeOfAKind

	// Straight represents a hand composed of five cards of consecutive rank.
	// Ex: A♠ K♣ Q♦ J♥ T♦
	Straight

	// Flush represents a hand composed of five cards that share the same suit.
	// Ex: T♠ 7♠ 4♠ 3♠ 2♠
	Flush

	// FullHouse represents a hand composed of three of a kind and a pair.
	// Ex: 4♠ 4♣ 4♦ 2♠ 2♥
	FullHouse

	// FourOfAKind represents a hand composed of four cards of the same rank.
	// Ex: A♠ A♣ A♦ A♥ 5♥
	FourOfAKind

	// StraightFlush represents a hand composed of five cards of consecutive
	// rank that share the same suit.
	// Ex: 5♥ 4♥ 3♥ 2♥ A♥
	StraightFlush

	// RoyalFlush represents a hand composed of ace, king, queen, jack, and ten
	// of the same suit.
	// Ex: A♥ K♥ Q♥ J♥ T♥
	RoyalFlush
)

var rankingNames = map[Ranking]string{
	HighCard:      "high card",
	Pair:          "pair",
	TwoPair:       "two pair",
	ThreeOfAKind:  "three of a kind",
	Straight:      "straight",
	Flush:         "flush",
	FullHouse:     "full house",
	FourOfAKind:   "four of a kind",
	StraightFlush: "straight flush",
	RoyalFlush:    "royal flush",
}

// String returns the name of the ranking
func (r Ranking) String() string {
	return rankingNames[r]
}

// A Hand is the highest poker hand derived from five or more cards.
type Hand struct {
	ranking     Ranking
	cards       []*Card
	description string
}

// A HandSorting is the sorting used to determine which hand is selected
// by NewHand and NewHandWithOptions.  Possible values include High and Low.
type HandSorting int

const (
	// High is a sorting method that will return the "high hand"
	High HandSorting = iota

	// Low is a sorting method that will return the "low hand"
	Low
)

// Options is a set of configuration that can be used in the NewHandWithOptions
// method to customize Hand selection.
type Options struct {
	Sorting         HandSorting
	IgnoreStraights bool
	IgnoreFlushes   bool
	AceIsLow        bool
}

// NewHand is a convience method for NewHandWithOptions using the Default Options.
func NewHand(cards []*Card) *Hand {
	return NewHandWithOptions(cards, Options{})
}

// NewHandWithOptions forms a hand with options to allow for
// customization of hand selection.  If less than five cards
// are given, blank cards will be inserted so that a value
// can still be calculated.
func NewHandWithOptions(cards []*Card, opts Options) *Hand {
	combos := cardCombos(cards)
	hands := []*Hand{}
	for _, c := range combos {
		hand := handForFiveCards(c, opts)
		hands = append(hands, hand)
	}
	index := len(hands) - 1
	if opts.Sorting == Low {
		index = 0
	}
	sort.Sort(ByHighHand(hands))
	return hands[index]
}

// Ranking returns the hand ranking of the hand.
func (h *Hand) Ranking() Ranking {
	return h.ranking
}

// Cards returns the five cards used in the best hand ranking for the hand.
func (h *Hand) Cards() []*Card {
	return h.cards
}

// Description returns a user displayable description of the hand such as
// "full house kings full of sixes".
func (h *Hand) Description() string {
	return h.description
}

// String returns the description followed by the cards used.
func (h *Hand) String() string {
	return fmt.Sprintf("%s %v", h.Description(), h.Cards())
}

// CompareTo returns a positive value if this hand beats the other hand, a
// negative value if this hand loses to the other hand, and zero if the hands
// are equal.
func (h *Hand) CompareTo(o *Hand) int {
	if h.Ranking() != o.Ranking() {
		return int(h.Ranking()) - int(o.Ranking())
	}
	hCards := h.Cards()
	oCards := o.Cards()
	for i := 0; i < 5; i++ {
		hCard, oCard := hCards[i], oCards[i]
		hIndex, oIndex := hCard.Rank().IndexOf(), oCard.Rank().IndexOf()
		if hIndex != oIndex {
			return hIndex - oIndex
		}
	}
	return 0
}

/* MarshalJSON implements the json.Marshaler interface.
   The json format is:
   {"ranking":9,"cards":["A♠","K♠","Q♠","J♠","T♠"],"description":"royal flush"}
*/
func (h *Hand) MarshalJSON() ([]byte, error) {
	cards := h.Cards()
	b, err := json.Marshal(&cards)
	if err != nil {
		return []byte{}, err
	}
	const format = `{"ranking":%v,"cards":%v,"description":"%v"}`
	s := fmt.Sprintf(format, h.Ranking(), string(b), h.Description())
	return []byte(s), nil
}

/* UnmarshalJSON implements the json.Unmarshaler interface.
   The json format is:
   {"ranking":9,"cards":["A♠","K♠","Q♠","J♠","T♠"],"description":"royal flush"}
*/
func (h *Hand) UnmarshalJSON(b []byte) error {
	type handJSON struct {
		Cards []*Card
	}
	m := &handJSON{}
	if err := json.Unmarshal(b, m); err != nil {
		return err
	}
	h = NewHand(m.Cards)
	return nil
}

// ByHighHand is a slice of hands sort in ascending value
type ByHighHand []*Hand

// Len implements the sort.Interface interface.
func (a ByHighHand) Len() int { return len(a) }

// Swap implements the sort.Interface interface.
func (a ByHighHand) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// Less implements the sort.Interface interface.
func (a ByHighHand) Less(i, j int) bool {
	iHand, jHand := a[i], a[j]
	return iHand.CompareTo(jHand) < 0
}

func handForFiveCards(cards []*Card, opts Options) *Hand {
	cards = formCards(cards, opts)
	for _, r := range rankings {
		if r.vFunc(cards, opts) {
			return &Hand{
				ranking:     r.r,
				cards:       cards,
				description: r.dFunc(cards),
			}
		}
	}
	panic("should never get here")
}

func cardCombos(cards []*Card) [][]*Card {
	cCombo := [][]*Card{}
	l := 5
	if len(cards) < 5 {
		l = len(cards)
	}
	indexCombos := combinations(len(cards), l)

	for _, combo := range indexCombos {
		cCards := []*Card{}
		for _, i := range combo {
			cCards = append(cCards, cards[i])
		}
		cCombo = append(cCombo, cCards)
	}
	return cCombo
}

type ranking struct {
	r     Ranking
	vFunc validFunc
	dFunc descFunc
}

type validFunc func([]*Card, Options) bool
type descFunc func([]*Card) string

var (
	highCard = ranking{
		r: HighCard,
		vFunc: func(cards []*Card, opts Options) bool {
			flush := hasFlush(cards)
			straight := hasStraight(cards)
			pairs := hasPairs(cards, []int{1, 1, 1, 1, 1})
			if !opts.IgnoreStraights {
				pairs = pairs && !straight
			}
			if !opts.IgnoreFlushes {
				pairs = pairs && !flush
			}
			return pairs
		},
		dFunc: func(cards []*Card) string {
			r := cards[0].Rank()
			return fmt.Sprintf("high card %v high", r.singularName())
		},
	}

	pair = ranking{
		r: Pair,
		vFunc: func(cards []*Card, opts Options) bool {
			return hasPairs(cards, []int{2, 2, 1, 1, 1})
		},
		dFunc: func(cards []*Card) string {
			r := cards[0].Rank()
			return fmt.Sprintf("pair of %v", r.pluralName())
		},
	}

	twoPair = ranking{
		r: TwoPair,
		vFunc: func(cards []*Card, opts Options) bool {
			return hasPairs(cards, []int{2, 2, 2, 2, 1})
		},
		dFunc: func(cards []*Card) string {
			r1 := cards[0].Rank()
			r2 := cards[2].Rank()
			return fmt.Sprintf("two pair %v and %v", r1.pluralName(), r2.pluralName())
		},
	}

	threeOfAKind = ranking{
		r: ThreeOfAKind,
		vFunc: func(cards []*Card, opts Options) bool {
			return hasPairs(cards, []int{3, 3, 3, 1, 1})
		},
		dFunc: func(cards []*Card) string {
			r := cards[0].Rank()
			return fmt.Sprintf("three of a kind %v", r.pluralName())
		},
	}

	straight = ranking{
		r: Straight,
		vFunc: func(cards []*Card, opts Options) bool {
			if opts.IgnoreStraights {
				return false
			}
			flush := hasFlush(cards)
			straight := hasStraight(cards)
			return !flush && straight
		},
		dFunc: func(cards []*Card) string {
			r := cards[0].Rank()
			return fmt.Sprintf("straight %v high", r.singularName())
		},
	}

	flush = ranking{
		r: Flush,
		vFunc: func(cards []*Card, opts Options) bool {
			if opts.IgnoreFlushes {
				return false
			}

			flush := hasFlush(cards)
			straight := hasStraight(cards)
			return flush && !straight
		},
		dFunc: func(cards []*Card) string {
			r1 := cards[0].Rank()
			return fmt.Sprintf("flush %v high", r1.singularName())
		},
	}

	fullHouse = ranking{
		r: FullHouse,
		vFunc: func(cards []*Card, opts Options) bool {
			return hasPairs(cards, []int{3, 3, 3, 2, 2})
		},
		dFunc: func(cards []*Card) string {
			r1 := cards[0].Rank()
			r2 := cards[3].Rank()
			return fmt.Sprintf("full house %v full of %v", r1.pluralName(), r2.pluralName())
		},
	}

	fourOfAKind = ranking{
		r: FourOfAKind,
		vFunc: func(cards []*Card, opts Options) bool {
			return hasPairs(cards, []int{4, 4, 4, 4, 1})
		},
		dFunc: func(cards []*Card) string {
			r := cards[0].Rank()
			return fmt.Sprintf("four of a kind %v", r.pluralName())
		},
	}

	straightFlush = ranking{
		r: StraightFlush,
		vFunc: func(cards []*Card, opts Options) bool {
			if opts.IgnoreStraights || opts.IgnoreFlushes {
				return false
			}
			flush := hasFlush(cards)
			straight := hasStraight(cards)
			return cards[0].Rank() != Ace && flush && straight
		},
		dFunc: func(cards []*Card) string {
			r := cards[0].Rank()
			return fmt.Sprintf("straight flush %v high", r.singularName())
		},
	}

	royalFlush = ranking{
		r: RoyalFlush,
		vFunc: func(cards []*Card, opts Options) bool {
			if opts.IgnoreStraights || opts.IgnoreFlushes {
				return false
			}
			flush := hasFlush(cards)
			straight := hasStraight(cards)
			return cards[0].Rank() == Ace && flush && straight
		},
		dFunc: func(cards []*Card) string {
			return "royal flush"
		},
	}

	rankings = []ranking{highCard, pair, twoPair, threeOfAKind,
		straight, flush, fullHouse, fourOfAKind, straightFlush, royalFlush}
)

func formCards(cards []*Card, opts Options) []*Card {
	var ranks []Rank
	if opts.AceIsLow {
		// sort cards staring w/ king
		sort.Sort(sort.Reverse(byAceLow(cards)))
		// sort ranks starting w/ king
		ranks = allRanks()
		sort.Sort(sort.Reverse(byAceLowRank(ranks)))
	} else {
		// sort cards staring w/ ace
		sort.Sort(sort.Reverse(byAceHigh(cards)))
		// sort ranks starting w/ ace
		ranks = allRanks()
		sort.Sort(sort.Reverse(byAceHighRank(ranks)))
	}

	// form cards starting w/ most paired
	formed := []*Card{}
	for i := 4; i > 0; i-- {
		for _, r := range ranks {
			rCards := cardsForRank(cards, r)
			if len(rCards) == i {
				formed = append(formed, rCards...)
			}
		}
	}

	dif := 5 - len(formed)
	for i := 0; i < dif; i++ {
		s := fmt.Sprintf("?%d", i+1)
		formed = append(formed, &Card{rank: Rank(s), suit: Suit(s)})
	}
	// check for low straight
	return formLowStraight(formed)
}

func hasPairs(cards []*Card, pairNums []int) bool {
	for i := 0; i < 5; i++ {
		card := cards[i]
		num := pairNums[i]
		if num != len(cardsForRank(cards, card.Rank())) {
			return false
		}
	}
	return true
}

func hasFlush(cards []*Card) bool {
	if hasBlankCards(cards) {
		return false
	}
	suit := cards[0].Suit()
	has := true
	for _, c := range cards {
		has = has && c.Suit() == suit
	}
	return has
}

func hasStraight(cards []*Card) bool {
	if hasBlankCards(cards) {
		return false
	}
	lastIndex := cards[0].Rank().IndexOf()
	straight := true
	for i := 1; i < 5; i++ {
		index := cards[i].Rank().IndexOf()
		straight = straight && (lastIndex == index+1)
		lastIndex = index
	}
	return straight || hasLowStraight(cards)
}

func hasLowStraight(cards []*Card) bool {
	return cards[0].Rank() == Five &&
		cards[1].Rank() == Four &&
		cards[2].Rank() == Three &&
		cards[3].Rank() == Two &&
		cards[4].Rank() == Ace
}

func formLowStraight(cards []*Card) []*Card {
	has := cards[0].Rank() == Ace &&
		cards[1].Rank() == Five &&
		cards[2].Rank() == Four &&
		cards[3].Rank() == Three &&
		cards[4].Rank() == Two
	if has {
		cards = []*Card{cards[1], cards[2], cards[3], cards[4], cards[0]}
	}
	return cards
}

func hasBlankCards(cards []*Card) bool {
	for _, c := range cards {
		if strings.Contains(string(c.Rank()), "?") {
			return true
		}
	}
	return false
}

func cardsForRank(cards []*Card, r Rank) []*Card {
	rCards := []*Card{}
	for _, c := range cards {
		if c.Rank() == r {
			rCards = append(rCards, c)
		}
	}
	return rCards
}