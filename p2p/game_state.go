package p2p

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

type PlayersList []*Player 

func (list PlayersList) Len() int {
	return len(list)
}

func (list PlayersList) Swap(i int, j int) {
	list[i], list[j] = list[j], list[i]
}

func (list PlayersList) Less(i int, j int) bool {
	_, portStrI, errI := net.SplitHostPort(list[i].listenAddr)
	_, portStrJ, errJ := net.SplitHostPort(list[j].listenAddr)
	if errI != nil || errJ != nil {
		return false
	}
	portI, _ := strconv.Atoi(portStrI)
	portJ, _ := strconv.Atoi(portStrJ)
	return portI < portJ
}

type Player struct {
	Status GameStatus
	listenAddr string
}

func (p *Player) String() string {
	return fmt.Sprintf("%s:%s", p.listenAddr, p.Status)
}

type GameState struct {
	listenAddr string
	broadcastch chan BroadcastTo
	isDealer bool
	gameStatus GameStatus
	playersLock sync.RWMutex
	playersList PlayersList
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
	g.AddPlayer(addr, GameStatusWaitingForCards)
	go g.loop()
	return g
}

func (g *GameState) SetStatus(status GameStatus) {
	if g.gameStatus != status {
		atomic.StoreInt32((*int32)(&g.gameStatus), (int32)(status))
		g.SetPlayerStatus(g.listenAddr, status)
	}
}

func (g *GameState) playersWaitingForCards() int {
	totalPlayers := 0
	for i := 0; i < len(g.playersList); i++ {
		if g.playersList[i].Status == GameStatusWaitingForCards {
			totalPlayers++;
		}
	}
	return totalPlayers
}

func (g *GameState) CheckStatus() {
	// playersWaiting := atomic.LoadInt32((&g.playersWaitingForCards))
	playersWaiting := g.playersWaitingForCards()

	if playersWaiting == len(g.players) && 
	g.isDealer && g.gameStatus == GameStatusWaitingForCards {
		logrus.WithFields(logrus.Fields{
			"addr": g.listenAddr,
		}).Info("deal cards")
		g.InitiateShuffleAndDeal()
	}
}

func (g *GameState) GetPlayerWithStatus(status GameStatus) []string {
	players := []string{}
	for addr, player := range g.players {
		if player.Status == status {
			players = append(players, addr)
		}
	}
	return players
}

func (g *GameState) getPositionOnTable() int {
	for i, player := range g.playersList {
		if player.listenAddr == g.listenAddr {
			return i
		}
	}
	panic("player does not exist")
}

func (g *GameState) getNextPositionOnTable() int {
	i := g.getPositionOnTable()
	return (i + 1) % len(g.playersList)
}

func (g *GameState) getPrevPositionOnTable() int {
	i := g.getPositionOnTable()
	return (i - 1 + len(g.playersList)) % len(g.playersList)
}

func (g *GameState) ShuffleAndEncrypt(from string, deck [][]byte) error {
	g.SetPlayerStatus(from, GameStatusShuffleAndDeal)
	prevPlayer := g.playersList[g.getPrevPositionOnTable()]
	if g.isDealer && prevPlayer.listenAddr == from {
		logrus.Info("end shuffle cycle")
		return nil
	}
	dealToPlayer := g.playersList[g.getNextPositionOnTable()]
	g.SendToPlayer(dealToPlayer.listenAddr, MessageEncDeck{Deck: [][]byte{}})
	g.SetStatus(GameStatusShuffleAndDeal)
	return nil
}

func (g *GameState) InitiateShuffleAndDeal() {
	dealToPlayer := g.playersList[g.getNextPositionOnTable()]
	g.SetStatus(GameStatusShuffleAndDeal)
	g.SendToPlayer(dealToPlayer.listenAddr, MessageEncDeck{Deck: [][]byte{}})
}

func (g *GameState) SendToPlayer(addr string, payload any) {
	g.broadcastch <- BroadcastTo {
		To: []string{addr},
		Payload: payload,
	}
	logrus.WithFields(logrus.Fields{
		"payload": payload,
		"player": addr,
	}).Info("sending payload to player")
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

	player := &Player{
		listenAddr: addr,
	}
	g.players[addr] = player
	g.playersList = append(g.playersList, player)
	sort.Sort(g.playersList)
	g.SetPlayerStatus(addr, status)
	logrus.WithFields(logrus.Fields{
		"addr": addr,
		"status": status,
	}).Info("new player joined")
}

func (g *GameState) loop() {
	ticker := time.NewTicker(time.Second * 5)
	for {
		<- ticker.C
		logrus.WithFields(logrus.Fields{
			"players connected": g.playersList,
			"status": g.gameStatus,
			"addr": g.listenAddr,
		}).Info()
	}
}
