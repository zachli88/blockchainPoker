package main
import (
	// "fmt"
	// "github.com/zachli88/blockchainPoker/deck"
	"github.com/zachli88/blockchainPoker/p2p"
)

func main() {
	cfg := p2p.ServerConfig{
		Version: "1",
		ListenAddr: "127.0.0.1:3000",
	}
	server := p2p.NewServer(cfg)
	go server.Start()

	remoteCfg := p2p.ServerConfig{
		Version: "1",
		ListenAddr: "127.0.0.1:4000",
	}
	remoteServer := p2p.NewServer(remoteCfg)
	go remoteServer.Start()
	if err := remoteServer.Connect("127.0.0.1:3000"); err != nil {
		panic(err)
	}

	select{}
}
