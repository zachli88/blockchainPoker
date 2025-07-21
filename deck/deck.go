package deck

import (
	"fmt"
	"math/rand"
	"strconv"
)

type Suit int

func (s Suit) String() string {
	switch s {
	case Spades:
		return "SPADES"
	case Hearts:
		return "HEARTS"
	case Diamonds:
		return "DIAMONDS"
	case Clubs:
		return "CLUBS"
	default:
		panic("Invalid card suit")
	}
}

func (c Card) String() string {
	value := strconv.Itoa(c.Value)
	if c.Value == 1 {
		value = "ACE"
	}
	return fmt.Sprintf("%s of %s %s", value, c.Suit, suitToUnicode(c.Suit))
}

const (
	Spades Suit = iota
	Hearts
	Diamonds
	Clubs
)

type Card struct {
	Suit Suit
	Value int
}

func NewCard(s Suit, v int) Card {
	if v > 13 {
		panic("The value of the card cannot be greater than 13")
	}
	return Card{
		Suit: s,
		Value: v,
	}
}

type Deck [52]Card

func New() Deck {
	nSuits := 4
	nValues := 13
	d :=[52]Card{}
	for i:= 0; i < nSuits; i++ {
		for j := 0; j < nValues; j++ {
			d[i * nValues + j] = NewCard(Suit(i), j + 1)
		}
	}
	return shuffle(d)
}

func shuffle(d Deck) Deck {
	for i := 0; i < len(d); i++ {
		r := rand.Intn(i + 1)
		d[i], d[r] = d[r], d[i]
	}
	return d
}

func suitToUnicode(s Suit) string {
	switch s {
	case Spades:
		return "♠"
	case Hearts:
		return "♥"
	case Diamonds:
		return "♦"
	case Clubs:
		return "♣"
	default:
		panic("Invalid card suit")
	}
}
