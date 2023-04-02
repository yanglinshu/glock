package cli

import (
	"fmt"

	"github.com/yanglinshu/glock/internal/blockchain"
	"github.com/yanglinshu/glock/internal/errors"
	"github.com/yanglinshu/glock/internal/transaction"
)

// createBlockchain creates a new blockchain
func createBlockchain(address, nodeID string) error {
	if !transaction.ValidateAddress(address) {
		return errors.ErrInvalidAddress
	}

	bc, err := blockchain.CreateBlockchain(address, nodeID)
	if err != nil {
		return err
	}
	defer bc.CloseDB()

	UTXOSet := blockchain.UTXOSet{Blockchain: bc}
	UTXOSet.Reindex()

	fmt.Println("Done!")
	return nil
}

// createWallet creates a new wallet
func createWallet(nodeID string) error {
	wallets, _ := transaction.NewWallets(nodeID)

	address, err := wallets.CreateWallet()
	if err != nil {
		return err
	}

	wallets.SaveToFile(nodeID)

	fmt.Printf("Your new address: %s\n", address)
	return nil
}
