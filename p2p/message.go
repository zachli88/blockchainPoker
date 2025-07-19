package p2p

// "io"

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
}

type MessagePeerList struct {
	Peers []string
}
