package p2p

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

type Player struct {
	Status GameStatus
}

type GameState struct {
	listenAddr string
	broadcastch chan BroadcastTo
	isDealer bool
	gameStatus GameStatus
	playersWaitingForCards int32
	playersLock sync.RWMutex
	players map[string]*Player
	deckReceivedLock sync.RWMutex
	deckReceived map[string]bool
}

func NewGameState(addr string, broadcastch chan BroadcastTo) *GameState {
	g := &GameState{
		listenAddr: addr,
		broadcastch: broadcastch,
		isDealer: false,
		gameStatus: GameStatusWaitingForCards,
		players: make(map[string]*Player),
		deckReceived: make(map[string]bool),
	}
	go g.loop()
	return g
}

func (g *GameState) SetStatus(status GameStatus) {
	if g.gameStatus != status {
		atomic.StoreInt32((*int32)(&g.gameStatus), (int32)(status))
	}
}

func (g *GameState) AddPlayerWaitingForCards() {
	atomic.AddInt32(&g.playersWaitingForCards, 1)
}

func (g *GameState) CheckStatus() {
	playersWaiting := atomic.LoadInt32((&g.playersWaitingForCards))

	if playersWaiting == int32(len(g.players)) && 
	g.isDealer && g.gameStatus == GameStatusWaitingForCards {
		logrus.WithFields(logrus.Fields{
			"addr": g.listenAddr,
		}).Info("deal cards")
		g.InitiateShuffleAndDeal()
	}
}

func (g *GameState) GetPlayerWithStatus(status GameStatus) []string {
	players := []string{}
	for addr, _ := range g.players {
		players = append(players, addr)
	}
	return players
}

func (g *GameState) SetDeckReceived(from string) {
	g.deckReceivedLock.Lock()
	g.deckReceived[from] = true
	g.deckReceivedLock.Unlock()
}

func (g *GameState) ShuffleAndEncrypt(from string, deck [][]byte) error {
	g.SetStatus(GameStatusReceivingCards)
	g.SetDeckReceived(from)
	players := g.GetPlayerWithStatus(GameStatusReceivingCards)

	g.deckReceivedLock.RLock()
	for _, addr := range players {
		_, ok := g.deckReceived[addr] 
		if !ok {
			return nil
		}
	}
	g.deckReceivedLock.RUnlock()
	g.SetStatus(GameStatusPreFlop)
	g.SendToPlayersWithStatus(MessageEncDeck{Deck: [][]byte{}}, GameStatusReceivingCards)
	return nil
}

func (g *GameState) InitiateShuffleAndDeal() {
	g.SetStatus(GameStatusReceivingCards)
	// g.broadcastch <- MessageEncDeck{Deck: [][]byte{}}
	g.SendToPlayersWithStatus(MessageEncDeck{Deck:[][]byte{}}, GameStatusWaitingForCards)
}

func (g *GameState) SendToPlayersWithStatus(payload any, status GameStatus) {
	players := g.GetPlayerWithStatus(status)
	g.broadcastch <- BroadcastTo{
		To: players,
		Payload: payload,
	}
	logrus.WithFields(logrus.Fields{
		"payload": payload,
		"players": players,
	}).Info("sending to players")
}

func (g *GameState) DealCards() {
	// g.broadcastch <- MessageEncDeck{Deck: [][]byte{}}
}

func (g *GameState) SetPlayerStatus(addr string, status GameStatus) {
	player, ok := g.players[addr]
	if !ok {
		panic("player could not be found")
	}
	player.Status = status
	g.CheckStatus()
}

func (g *GameState) LenPlayersConnectedWithLock() int {
	g.playersLock.RLock()
	defer g.playersLock.RUnlock()
	return len(g.players)
}

func (g *GameState) AddPlayer(addr string, status GameStatus) {
	g.playersLock.Lock()
	defer g.playersLock.Unlock()

	if status == GameStatusWaitingForCards {
		g.AddPlayerWaitingForCards()
	}
	g.players[addr] = new(Player)

	g.SetPlayerStatus(addr, status)

	logrus.WithFields(logrus.Fields{
		"addr": addr,
		"status": status,
	}).Info("new player joined")
}

func (g *GameState) loop() {
	ticker := time.NewTicker(time.Second * 5)
	for {
		select {
		case <- ticker.C:
			logrus.WithFields(logrus.Fields{
				"players connected": g.LenPlayersConnectedWithLock(),
				"status": g.gameStatus,
				"addr": g.listenAddr,
			}).Info()
		default:
		}
	}
}
