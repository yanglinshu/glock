package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/yanglinshu/glock/internal/blockchain"
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
	fmt.Println("  get -balance ADDRESS - Get balance of ADDRESS")
	fmt.Println("  create -blockchain ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  create -wallet - Create a new wallet")
	fmt.Println("  show -blockchain - Print all the blocks of the blockchain")
	fmt.Println("  show -addresses - Print all the addresses in the wallet file")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
	fmt.Println("  update -UTXO - Update the UTXO set")
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
	// Get command, has subcommand balance
	getCmd := flag.NewFlagSet("get", flag.ExitOnError)
	getCmdBalance := getCmd.String("balance", "", "The address to get balance for")

	// Create command, has subcommand blockchain wallet
	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	createCmdBlockchain := createCmd.String("blockchain", "", "The address to send genesis block reward to")
	createCmdWallet := createCmd.Bool("wallet", false, "Create a new wallet")

	// Show command, has subcommand blockchain, addresses
	showCmd := flag.NewFlagSet("print", flag.ExitOnError)
	showCmdBlockchain := showCmd.Bool("blockchain", false, "Print all the blocks of the blockchain")
	showCmdAddresses := showCmd.Bool("addresses", false, "Print all the addresses in the wallet file")

	// Send command, defaultly create a send transaction, has parameters from, to, amount
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	sendCmdFrom := sendCmd.String("from", "", "Source wallet address")
	sendCmdTo := sendCmd.String("to", "", "Destination wallet address")
	sendCmdAmount := sendCmd.Int("amount", 0, "Amount to send")

	// Update command, has subcommand UTXO
	updateCmd := flag.NewFlagSet("update", flag.ExitOnError)
	updateCmdUTXO := updateCmd.Bool("UTXO", false, "Update the UTXO set")

	// Parse the command line arguments

	switch os.Args[1] {
	case "get":
		err := getCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "create":
		err := createCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "show":
		err := showCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "update":
		err := updateCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	// Execute the command get if it was parsed
	if getCmd.Parsed() {
		if *getCmdBalance == "" {
			getCmd.Usage()
			fmt.Println("Invalid address: ", *getCmdBalance)
			os.Exit(1)
		}
		err := getBalance(*getCmdBalance)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	// Execute the command create if it was parsed
	if createCmd.Parsed() {
		if *createCmdBlockchain != "" {
			err := createBlockchain(*createCmdBlockchain)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		} else if *createCmdWallet {
			err := createWallet()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		} else {
			createCmd.Usage()
			os.Exit(1)
		}
	}

	// Execute the command show if it was parsed
	if showCmd.Parsed() {
		if *showCmdBlockchain {
			err := showBlockchain()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		} else if *showCmdAddresses {
			err := showAddresses()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		} else {
			showCmd.Usage()
			os.Exit(1)
		}
	}

	// Execute the command send if it was parsed
	if sendCmd.Parsed() {
		if *sendCmdFrom == "" || *sendCmdTo == "" || *sendCmdAmount <= 0 {
			sendCmd.Usage()
			fmt.Println("Invalid address or amount")
			os.Exit(1)
		}
		err := sendTransaction(*sendCmdFrom, *sendCmdTo, *sendCmdAmount)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	// Execute the command update if it was parsed
	if updateCmd.Parsed() {
		if *updateCmdUTXO {
			err := updateUTXO()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		} else {
			updateCmd.Usage()
			os.Exit(1)
		}
	}

}
