package main

import (
	"log"

	"github.com/yanglinshu/glock/internal/blockchain"
	"github.com/yanglinshu/glock/internal/cli"
)

func main() {
	bc, err := blockchain.NewBlockchain()
	if err != nil {
		log.Fatal(err)
	}
	defer bc.Close()

	cli := cli.NewCLI(bc)
	cli.Run()
}
