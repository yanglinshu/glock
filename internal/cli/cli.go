package cli

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/yanglinshu/glock/internal/blockchain"
	"github.com/yanglinshu/glock/internal/errors"
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
	fmt.Println("  new - create a new wallet")
	fmt.Println("  list - list all the addresses in the wallet file")
	fmt.Println("  print - print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - send AMOUNT of coins from FROM address to TO")
}

// createBlockchain creates a new blockchain
func (cli *CLI) createBlockchain(address string) error {
	if !transaction.ValidateAddress(address) {
		return errors.ErrorInvalidAddress
	}

	bc, err := blockchain.CreateBlockchain(address)
	defer bc.CloseDB()
	if err != nil {
		return err
	}

	fmt.Println("Done!")
	return nil
}

// createWallet creates a new wallet
func (cli *CLI) createWallet() error {
	wallets, _ := transaction.NewWallets()

	address, err := wallets.CreateWallet()
	if err != nil {
		return err
	}

	wallets.SaveToFile()

	fmt.Printf("Your new address: %s\n", address)
	return nil
}

// getBalance gets the balance of an address
func (cli *CLI) getBalance(address string) error {
	if !transaction.ValidateAddress(address) {
		return errors.ErrorInvalidAddress
	}

	bc, err := blockchain.NewBlockchain()
	defer bc.CloseDB()
	if err != nil {
		return err
	}

	balance := 0
	publicKeyHash := transaction.Base58Decode([]byte(address))
	publicKeyHash = publicKeyHash[1 : len(publicKeyHash)-4]
	UTXOs, err := bc.FindUTXO(publicKeyHash)
	if err != nil {
		return err
	}

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
	return nil
}

// listAddresses lists all the addresses in the wallet file
func (cli *CLI) listAddresses() error {
	wallets, err := transaction.NewWallets()
	if err != nil {
		return err
	}

	addresses := wallets.GetAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}

	return nil
}

// printChain prints the blockchain
func (cli *CLI) printChain() error {
	bc, err := blockchain.NewBlockchain()
	defer bc.CloseDB()
	if err != nil {
		return err
	}

	bci := bc.Iterator()

	for {
		block, err := bci.Next()
		if err != nil {
			return err
		}

		fmt.Printf("============ Block %x ============\n", block.Hash)
		fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
		pow := blockchain.NewProofOfWork(block)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Printf("\n\n")

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return nil
}

// sendCoin sends coins from one address to another
func (cli *CLI) sendCoin(from, to string, amount int) error {
	bc, err := blockchain.NewBlockchain()
	defer bc.CloseDB()
	if err != nil {
		return err
	}

	tx, err := blockchain.NewUTXOTransaction(from, to, amount, bc)
	if err != nil {
		return err
	}

	err = bc.MineBlock([]*transaction.Transaction{tx})
	if err != nil {
		return err
	}

	fmt.Println("Success!")
	return nil
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
	newCmd := flag.NewFlagSet("new", flag.ExitOnError)
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)

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
	case "new":
		err := newCmd.Parse(os.Args[2:])
		if err != nil {
			cli.printUsage()
			os.Exit(1)
		}
	case "list":
		err := listCmd.Parse(os.Args[2:])
		if err != nil {
			cli.printUsage()
			os.Exit(1)
		}
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

	// Execute the command new if it was parsed
	if newCmd.Parsed() {
		err := cli.createWallet()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	// Execute the command list if it was parsed
	if listCmd.Parsed() {
		err := cli.listAddresses()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	// Execute the command add if it was parsed
	if createCmd.Parsed() {
		if *createAddress == "" {
			createCmd.Usage()
			fmt.Println("Invalid address: ", *createAddress)
			os.Exit(1)
		}
		err := cli.createBlockchain(*createAddress)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	// Execute the command get if it was parsed
	if getCmd.Parsed() {
		if *getAddress == "" {
			getCmd.Usage()
			fmt.Println("Invalid address: ", *getAddress)
			os.Exit(1)
		}
		err := cli.getBalance(*getAddress)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	// Execute the command send if it was parsed
	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			fmt.Println("Invalid from/to/amount: ", *sendFrom, *sendTo, *sendAmount)
			os.Exit(1)
		}
		err := cli.sendCoin(*sendFrom, *sendTo, *sendAmount)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	// Execute the command print if it was parsed
	if printCmd.Parsed() {
		err := cli.printChain()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

}
