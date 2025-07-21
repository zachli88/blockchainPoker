package main

import (
	"time"

	"github.com/zachli88/blockchainPoker/p2p"
)

func main() {
	playerA := makeServerAndStart("127.0.0.1:3000")
	playerB := makeServerAndStart("127.0.0.1:4000")
	playerC := makeServerAndStart("127.0.0.1:5000")
	// playerD := makeServerAndStart("127.0.0.1:6000")
	// playerE := makeServerAndStart("127.0.0.1:7000")
	// playerF := makeServerAndStart("127.0.0.1:8000")
	time.Sleep(time.Second * 1)
	playerB.Connect(playerA.ListenAddr)
	time.Sleep(time.Second * 1)
	playerC.Connect(playerB.ListenAddr)
	time.Sleep(time.Second * 1)
	// playerD.Connect(playerC.ListenAddr)
	// time.Sleep(time.Second * 1)
	// playerE.Connect(playerA.ListenAddr)
	// time.Sleep(time.Second * 1)
	// playerF.Connect(playerC.ListenAddr)
	select{}
}


func makeServerAndStart(addr string) *p2p.Server {
	cfg := p2p.ServerConfig{
		Version: "1",
		ListenAddr: addr,
		GameVariant: p2p.TexasHoldem,
	}
	server := p2p.NewServer(cfg)
	go server.Start()
	return server
}
