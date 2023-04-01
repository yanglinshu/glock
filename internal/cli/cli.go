package cli

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/yanglinshu/glock/internal/blockchain"
)

// CLI represents the command line interface
type CLI struct {
	bc *blockchain.Blockchain // Blockchain instance
}

// NewCLI creates a new CLI instance
func NewCLI(bc *blockchain.Blockchain) *CLI {
	return &CLI{bc}
}

// printUsage prints the usage of the CLI
func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  add -data BLOCK_DATA - add a block to the blockchain")
	fmt.Println("  print - print all the blocks of the blockchain")
}

// validateArgs validates the command line arguments
func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// Run parses the command line arguments and executes the command
func (cli *CLI) Run() {
	cli.validateArgs()

	// CLI commands
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	printCmd := flag.NewFlagSet("print", flag.ExitOnError)
	addData := addCmd.String("data", "", "Block data")

	// Parse the command line arguments
	switch os.Args[1] {
	case "add":
		err := addCmd.Parse(os.Args[2:])
		if err != nil {
			cli.printUsage()
			log.Fatal(err)
		}
	case "print":
		err := printCmd.Parse(os.Args[2:])
		if err != nil {
			cli.printUsage()
			log.Fatal(err)
		}
	default:
		cli.printUsage()
		log.Fatalf("Invalid command: %s", os.Args[1])
	}

	// Execute the command add if it was parsed
	if addCmd.Parsed() {
		if *addData == "" {
			addCmd.Usage()
			log.Fatalf("Invalid data: %s", *addData)
		}
		cli.addBlock(*addData)
	}

	// Execute the command print if it was parsed
	if printCmd.Parsed() {
		cli.printChain()
	}

}

// addBlock adds a block to the blockchain
func (cli *CLI) addBlock(data string) {
	cli.bc.AddBlock(data)
	log.Printf("Added block %s", data)
}

// printChain prints the blockchain
func (cli *CLI) printChain() {
	bci := cli.bc.Iterator()

	for {
		block, err := bci.Next()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := blockchain.NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}

	}
}
