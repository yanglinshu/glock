package cli

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/yanglinshu/glock/internal/blockchain"
	"github.com/yanglinshu/glock/internal/transaction"
)

// CLI represents the command line interface
type CLI struct{}

// NewCLI creates a new CLI instance
func NewCLI(bc *blockchain.Blockchain) *CLI {
	return &CLI{}
}

// printUsage prints the usage of the CLI
func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  get -address ADDRESS - get the balance for an address")
	fmt.Println("  create -address ADDRESS - create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  print - print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - send AMOUNT of coins from FROM address to TO")
}

// createBlockchain creates a new blockchain
func (cli *CLI) createBlockchain(address string) {
	bc, err := blockchain.CreateBlockchain(address)
	defer bc.CloseDB()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Done!")
}

// getBalance gets the balance of an address
func (cli *CLI) getBalance(address string) {
	bc, err := blockchain.NewBlockchain()
	defer bc.CloseDB()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	balance := 0
	UTXOs, err := bc.FindUTXO(address)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

// printChain prints the blockchain
func (cli *CLI) printChain() {
	bc, err := blockchain.NewBlockchain()
	defer bc.CloseDB()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	bci := bc.Iterator()

	for {
		block, err := bci.Next()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := blockchain.NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}

	}
}

// sendCoin sends coins from one address to another
func (cli *CLI) sendCoin(from, to string, amount int) {
	bc, err := blockchain.NewBlockchain()
	defer bc.CloseDB()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	tx, err := blockchain.NewUTXOTransaction(from, to, amount, bc)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = bc.MineBlock([]*transaction.Transaction{tx})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Success!")
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
	getCmd := flag.NewFlagSet("get", flag.ExitOnError)
	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	printCmd := flag.NewFlagSet("print", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)

	// get command flags
	getAddress := getCmd.String("address", "", "The address to get balance for")

	// create command flags
	createAddress := createCmd.String("address", "", "The address to send genesis block reward to")

	// send command flags
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	// Parse the command line arguments
	switch os.Args[1] {
	case "get":
		err := getCmd.Parse(os.Args[2:])
		if err != nil {
			cli.printUsage()
			os.Exit(1)
		}
	case "create":
		err := createCmd.Parse(os.Args[2:])
		if err != nil {
			cli.printUsage()
			os.Exit(1)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			cli.printUsage()
			os.Exit(1)
		}
	case "print":
		err := printCmd.Parse(os.Args[2:])
		if err != nil {
			cli.printUsage()
			os.Exit(1)
		}
	default:
		cli.printUsage()
		fmt.Println("Invalid command: ", os.Args[1])
		os.Exit(1)
	}

	// Execute the command add if it was parsed
	if createCmd.Parsed() {
		if *createAddress == "" {
			createCmd.Usage()
			fmt.Println("Invalid address: ", *createAddress)
			os.Exit(1)
		}
		cli.createBlockchain(*createAddress)
	}

	// Execute the command get if it was parsed
	if getCmd.Parsed() {
		if *getAddress == "" {
			getCmd.Usage()
			fmt.Println("Invalid address: ", *getAddress)
			os.Exit(1)
		}
		cli.getBalance(*getAddress)
	}

	// Execute the command send if it was parsed
	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			fmt.Println("Invalid from/to/amount: ", *sendFrom, *sendTo, *sendAmount)
			os.Exit(1)
		}
		cli.sendCoin(*sendFrom, *sendTo, *sendAmount)
	}

	// Execute the command print if it was parsed
	if printCmd.Parsed() {
		cli.printChain()
	}

}
