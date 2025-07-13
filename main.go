package main
import (
	"fmt"
	"github.com/zachli88/blockchainPoker/deck"
)

func main() {
	for i := 0; i < 10; i++ {
		d := deck.New()
		fmt.Println(d)
		fmt.Println()
	}
}
