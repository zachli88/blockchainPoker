package p2p

import (
	"net"
	"sync"
	"fmt"
	"io"
	"bytes"
)

type TCPTransport struct {

}

type ServerConfig struct {
	Version string
	ListenAddr string
}

type Server struct {
	ServerConfig
	handler Handler
	listener net.Listener
	mu sync.RWMutex
	peers map[net.Addr]*Peer
	addPeer chan *Peer
	delPeer chan *Peer
	msgCh chan *Message
}

func NewServer(cfg ServerConfig) *Server {
	return &Server{
		handler: &DefaultHandler{},
		ServerConfig: cfg,
		peers: make(map[net.Addr]*Peer),
		addPeer: make(chan *Peer),
		delPeer: make(chan *Peer),
		msgCh: make(chan *Message),
	}
}

func (s *Server) Start() {
	go s.loop()
	if err := s.listen(); err != nil {
		panic(err)
	}
	fmt.Printf("game server running on port %s\n", s.ListenAddr)
	go s.acceptLoop()
}

func (s *Server) Connect(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	peer := &Peer{
		conn: conn,
	}
	s.addPeer <-peer
	return peer.Send([]byte(s.Version))
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			panic(err)
		}
		peer := &Peer{
			conn: conn,
		}
		s.addPeer <-peer
		peer.Send([]byte(s.Version))
		go s.handleConn(peer)
	}
}

func (s *Server) loop() {
	for {
		select {
		case peer := <- s.addPeer:
			s.peers[peer.conn.RemoteAddr()] = peer
			fmt.Printf("new player connected %s\n", peer.conn.RemoteAddr())
		case msg := <- s.msgCh:
			if err := s.handler.HandleMessage(msg); err != nil {
				panic(err)
			}
		case peer := <- s.delPeer:
			addr := peer.conn.RemoteAddr()
			delete(s.peers, addr)
			fmt.Printf("player disconnected %s\n", addr)
		}
	}
}
