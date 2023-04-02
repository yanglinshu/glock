package cli

import (
	"fmt"

	"github.com/yanglinshu/glock/internal/blockchain"
	"github.com/yanglinshu/glock/internal/errors"
	"github.com/yanglinshu/glock/internal/transaction"
)

// sendTransaction sends coins from one address to another
func sendTransaction(from, to string, amount int) error {
	if !transaction.ValidateAddress(from) {
		return errors.ErrInvalidAddress
	}

	if !transaction.ValidateAddress(to) {
		return errors.ErrInvalidAddress
	}

	bc, err := blockchain.NewBlockchain()
	defer bc.CloseDB()
	if err != nil {
		return err
	}

	UTXOSet := blockchain.UTXOSet{Blockchain: bc}

	tx, err := blockchain.NewUTXOTransaction(from, to, amount, &UTXOSet)
	if err != nil {
		return err
	}

	cbTx, err := transaction.NewCoinbaseTX(from, "")
	if err != nil {
		return err
	}

	newBlock, err := bc.MineBlock([]*transaction.Transaction{cbTx, tx})
	if err != nil {
		return err
	}

	UTXOSet.Update(newBlock)

	fmt.Println("Success!")
	return nil
}
