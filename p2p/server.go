package p2p

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

type GameVariant int

const (
	TexasHoldem GameVariant = iota
	Omaha
)

func (gv GameVariant) String() string {
	switch gv {
	case TexasHoldem:
		return "Texas Hold'em"
	case Omaha:
		return "Omaha"
	default:
		return "unknown"
	}
}
type ServerConfig struct {
	Version string
	ListenAddr string
	GameVariant GameVariant
}

type Server struct {
	ServerConfig
	transport *TCPTransport
	peers map[net.Addr]*Peer
	addPeer chan *Peer
	delPeer chan *Peer
	msgCh chan *Message
	gameState *GameState
}

func NewServer(cfg ServerConfig) *Server {
	s := &Server{
		ServerConfig: cfg,
		peers: make(map[net.Addr]*Peer),
		addPeer: make(chan *Peer, 10),
		delPeer: make(chan *Peer),
		msgCh: make(chan *Message),
		gameState: NewGameState(),
	}
	tr := NewTCPTransport(s.ListenAddr)
	s.transport = tr
	tr.AddPeer = s.addPeer
	tr.DelPeer = s.delPeer
	return s
}

func (s *Server) Start() {
	go s.loop()
	fmt.Printf("game server running on port %s\n", s.ListenAddr)
	logrus.WithFields(logrus.Fields{
		"port": s.ListenAddr,
		"variant": s.GameVariant,
	}).Info("started new game server")
	s.transport.ListenAndAccept()
}

func (s *Server) sendPeerList(p *Peer) error {
	peerList := MessagePeerList{
		Peers: make([]string, len(s.peers)),
	}

	it := 0
	for addr := range s.peers {
		peerList.Peers[it] = addr.String()
		it++;
	}

	msg := NewMessage(s.ListenAddr, peerList)
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}
	return p.Send(buf.Bytes())
}

func (s *Server) SendHandshake(p *Peer) error {
	hs := &Handshake{
		GameVariant: s.GameVariant,
		Version: s.Version,
		GameStatus: s.gameState.gameStatus,
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(hs); err != nil {
		return err
	}

	return p.Send(buf.Bytes())
}

func (s *Server) Connect(addr string) error {
	var conn net.Conn
	var err error
	for i := 0; i < 5; i++ {
        conn, err = net.Dial("tcp", addr)
        if err == nil {
            break
        }
        time.Sleep(time.Second) // wait 1 second before retry
    }
	if err != nil {
		return err
	}
	peer := &Peer{
		conn: conn,
		outbound: true,
	}
	s.addPeer <-peer
	return s.SendHandshake(peer)
}

func (s *Server) loop() {
	for {
		select {
		case peer := <- s.addPeer:
			if err := s.handshake(peer); err != nil {
				logrus.Errorf("handshake with incoming player failed: %s", err)
				peer.conn.Close()
				delete(s.peers, peer.conn.RemoteAddr())
				continue
			}
			go peer.ReadLoop(s.msgCh)

			if !peer.outbound {
				if err := s.SendHandshake(peer); err != nil {
					logrus.Info("failed to send handshake with peer")
					peer.conn.Close()
					delete(s.peers, peer.conn.RemoteAddr())
					continue
				}
				if err := s.sendPeerList(peer); err != nil {
					logrus.Errorf("peerlist error: %s", err)
					continue
				}
			}
			logrus.WithFields(logrus.Fields{
				"addr": peer.conn.RemoteAddr(),
			}).Info("handshake successful - new player connected")
			s.peers[peer.conn.RemoteAddr()] = peer

		case msg := <- s.msgCh:
			if err := s.handleMessage(msg); err != nil {
				panic(err)
			}
			
		case peer := <- s.delPeer:
			logrus.WithFields(logrus.Fields{
				"addr": peer.conn.RemoteAddr(),
			}).Info("player disconnected")
			addr := peer.conn.RemoteAddr()
			delete(s.peers, addr)
			continue
		}
	}
}

func (s *Server) handshake(p *Peer) error {
	hs := &Handshake{}
	if err := gob.NewDecoder(p.conn).Decode(hs); err != nil {
		return err
	}
	if hs.GameVariant != s.GameVariant {
		return fmt.Errorf("game variant mismatch %s\n", hs.GameVariant)
	}
	if hs.Version != s.Version {
		return fmt.Errorf("invalid game version %s\n", hs.Version)
	}

	logrus.WithFields(logrus.Fields{
		"peer": p.conn.RemoteAddr(),
		"version": hs.Version,
		"variant": hs.GameVariant,
		"gameStatus": hs.GameStatus,
	}).Info("received handshake")
	return nil
}

func (s *Server) handleMessage(msg *Message) error {
	logrus.WithFields(logrus.Fields{
		"from": msg.From,
	}).Info("received message")

	switch v := msg.Payload.(type) {
	case MessagePeerList:
		return s.handlePeerList(v)
	}
	return nil
}

func (s *Server) handlePeerList(l MessagePeerList) error {
	fmt.Printf("peerList %+v\n", l)
	for i := 0; i < len(l.Peers); i++ {
		if err := s.Connect(l.Peers[i]); err != nil {
			logrus.Error("failed to dial peer: ", err)
			continue
		}
	}
	return  nil
}

func init() {
	gob.Register(MessagePeerList{})
}
