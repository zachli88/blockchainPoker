package main

import (
	// "fmt"
	// "github.com/zachli88/blockchainPoker/deck"
	"log"

	"github.com/zachli88/blockchainPoker/p2p"
)

func main() {
	playerA := makeServerAndStart("127.0.0.1:3000")
	playerB := makeServerAndStart("127.0.0.1:4000")
	playerC := makeServerAndStart("127.0.0.1:5000")
	playerD := makeServerAndStart("127.0.0.1:6000")
	if err := playerC.Connect(playerA.ListenAddr); err != nil {
		log.Fatal(err)
	}

	playerB.Connect(playerC.ListenAddr)
	playerD.Connect(playerC.ListenAddr)

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
