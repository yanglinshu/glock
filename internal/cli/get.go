package cli

import (
	"fmt"

	"github.com/yanglinshu/glock/internal/blockchain"
	"github.com/yanglinshu/glock/internal/errors"
	"github.com/yanglinshu/glock/internal/transaction"
)

// getBalance gets the balance of an address
func getBalance(address string) error {
	if !transaction.ValidateAddress(address) {
		return errors.ErrInvalidAddress
	}

	bc, err := blockchain.NewBlockchain()
	defer bc.CloseDB()
	if err != nil {
		return err
	}

	UTXOSet := blockchain.UTXOSet{Blockchain: bc}

	balance := 0
	publicKeyHash := transaction.Base58Decode([]byte(address))
	publicKeyHash = publicKeyHash[1 : len(publicKeyHash)-4]
	UTXOs, err := UTXOSet.FindUTXO(publicKeyHash)
	if err != nil {
		return err
	}

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
	return nil
}
