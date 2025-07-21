package p2p

import "github.com/zachli88/blockchainPoker/deck"

type Message struct {
	Payload any
	From string
}

func NewMessage(from string, payload any) *Message {
	return &Message{
		From: from,
		Payload: payload,
	}
}

type Handshake struct {
	GameVariant GameVariant
	Version string
	GameStatus GameStatus
	ListenAddr string
}

type MessagePeerList struct {
	Peers []string
}

type MessageCards struct {
	Deck deck.Deck
}
