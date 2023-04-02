package cli

import (
	"fmt"
	"strconv"

	"github.com/yanglinshu/glock/internal/block"
	"github.com/yanglinshu/glock/internal/blockchain"
	"github.com/yanglinshu/glock/internal/transaction"
)

// showBlockchain prints the blockchain
func showBlockchain() error {
	bc, err := blockchain.NewBlockchain()
	defer bc.CloseDB()
	if err != nil {
		return err
	}

	bci := bc.Iterator()

	for {
		bl, err := bci.Next()
		if err != nil {
			return err
		}

		fmt.Printf("============ Block %x ============\n", bl.Hash)
		fmt.Printf("Prev. block: %x\n", bl.PrevBlockHash)
		pow := block.NewProofOfWork(bl)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range bl.Transactions {
			fmt.Println(tx)
		}
		fmt.Printf("\n\n")

		if len(bl.PrevBlockHash) == 0 {
			break
		}
	}

	return nil
}

// showAddresses lists all the addresses in the wallet file
func showAddresses() error {
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
